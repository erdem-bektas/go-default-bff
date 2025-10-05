package services

import (
	"context"
	"encoding/json"
	"fiber-app/pkg/config"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type AuthService struct {
	config         *config.ZitadelConfig
	oauthConfig    *oauth2.Config
	logger         *zap.Logger
	oidcService    *OIDCDiscoveryService
	jwksService    *JWKSService
	pkceService    *PKCEService
	sessionService *SessionService
	oidcConfig     *OIDCConfiguration
}

type ZitadelUserInfo struct {
	Sub               string   `json:"sub"`
	Name              string   `json:"name"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	PreferredUsername string   `json:"preferred_username"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Roles             []string `json:"urn:zitadel:iam:org:project:roles"`
}

type AuthURLResponse struct {
	URL   string `json:"url"`
	State string `json:"state"`
}

func NewAuthService(cfg *config.ZitadelConfig, logger *zap.Logger) *AuthService {
	service := &AuthService{
		config:         cfg,
		logger:         logger,
		oidcService:    NewOIDCDiscoveryService(logger),
		pkceService:    NewPKCEService(logger),
		sessionService: NewSessionService(logger),
	}

	// Initialize OIDC discovery
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.initializeOIDC(ctx); err != nil {
		logger.Error("Failed to initialize OIDC", zap.Error(err))
		// Continue with manual configuration as fallback
		service.setupFallbackConfig()
	}

	return service
}

func (as *AuthService) initializeOIDC(ctx context.Context) error {
	// Discover OIDC configuration
	oidcConfig, err := as.oidcService.DiscoverConfiguration(ctx, as.config.Issuer)
	if err != nil {
		return fmt.Errorf("OIDC discovery failed: %w", err)
	}

	as.oidcConfig = oidcConfig

	// Setup OAuth2 config with discovered endpoints
	as.oauthConfig = &oauth2.Config{
		ClientID:     as.config.ClientID,
		ClientSecret: as.config.ClientSecret,
		RedirectURL:  as.config.RedirectURL,
		Scopes:       as.config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oidcConfig.AuthorizationEndpoint,
			TokenURL: oidcConfig.TokenEndpoint,
		},
	}

	// Initialize JWKS service
	jwksURL := oidcConfig.JWKSUri
	if as.config.JWKSURL != "" {
		jwksURL = as.config.JWKSURL
	}
	as.jwksService = NewJWKSService(jwksURL, as.logger)

	as.logger.Info("OIDC initialized successfully",
		zap.String("issuer", oidcConfig.Issuer),
		zap.String("auth_endpoint", oidcConfig.AuthorizationEndpoint),
		zap.String("token_endpoint", oidcConfig.TokenEndpoint),
		zap.String("jwks_uri", jwksURL),
	)

	return nil
}

func (as *AuthService) setupFallbackConfig() {
	// Fallback to manual configuration
	as.oauthConfig = &oauth2.Config{
		ClientID:     as.config.ClientID,
		ClientSecret: as.config.ClientSecret,
		RedirectURL:  as.config.RedirectURL,
		Scopes:       as.config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/v2/authorize", as.config.Domain),
			TokenURL: fmt.Sprintf("%s/oauth/v2/token", as.config.Domain),
		},
	}

	// Setup JWKS service with manual URL
	jwksURL := as.config.JWKSURL
	if jwksURL == "" {
		jwksURL = fmt.Sprintf("%s/oauth/v2/keys", as.config.Domain)
	}
	as.jwksService = NewJWKSService(jwksURL, as.logger)

	as.logger.Info("Using fallback OIDC configuration",
		zap.String("auth_url", as.oauthConfig.Endpoint.AuthURL),
		zap.String("token_url", as.oauthConfig.Endpoint.TokenURL),
		zap.String("jwks_url", jwksURL),
	)
}

// GenerateAuthURL - OAuth2 authorization URL with PKCE
func (as *AuthService) GenerateAuthURL() (*AuthURLResponse, error) {
	// Generate PKCE challenge
	pkceChallenge, err := as.pkceService.GenerateChallenge()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE challenge: %w", err)
	}

	// Build authorization URL with PKCE parameters
	authURL := as.oauthConfig.AuthCodeURL(
		pkceChallenge.State,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	as.logger.Info("Generated OAuth2 authorization URL",
		zap.String("state", pkceChallenge.State),
	)

	return &AuthURLResponse{
		URL:   authURL,
		State: pkceChallenge.State,
	}, nil
}

// ExchangeCodeForToken - Authorization code'u token ile değiştir (PKCE)
func (as *AuthService) ExchangeCodeForToken(ctx context.Context, code, state string) (*oauth2.Token, error) {
	// Validate PKCE challenge
	pkceChallenge, err := as.pkceService.ValidateAndGetChallenge(state)
	if err != nil {
		return nil, fmt.Errorf("PKCE validation failed: %w", err)
	}

	// Exchange code for token with PKCE verifier
	token, err := as.oauthConfig.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", pkceChallenge.CodeVerifier),
	)
	if err != nil {
		as.logger.Error("Token exchange failed", zap.Error(err))
		return nil, err
	}

	as.logger.Info("Token exchange successful",
		zap.String("token_type", token.TokenType),
		zap.Time("expiry", token.Expiry),
	)

	return token, nil
}

// GetUserInfo - Access token ile kullanıcı bilgilerini al
func (as *AuthService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*ZitadelUserInfo, error) {
	client := as.oauthConfig.Client(ctx, token)

	userInfoURL := fmt.Sprintf("%s/oidc/v1/userinfo", as.config.Domain)
	resp, err := client.Get(userInfoURL)
	if err != nil {
		as.logger.Error("Failed to get user info", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		as.logger.Error("User info request failed",
			zap.Int("status_code", resp.StatusCode),
		)
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo ZitadelUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		as.logger.Error("Failed to decode user info", zap.Error(err))
		return nil, err
	}

	as.logger.Info("User info retrieved",
		zap.String("sub", userInfo.Sub),
		zap.String("email", userInfo.Email),
		zap.Strings("roles", userInfo.Roles),
	)

	return &userInfo, nil
}

// ValidateToken - JWT token'ı JWKS ile validate et
func (as *AuthService) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	if as.jwksService == nil {
		return nil, fmt.Errorf("JWKS service not initialized")
	}

	claims, err := as.jwksService.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	return claims, nil
}

// HasRole - Kullanıcının belirli bir role'ü var mı kontrol et
func (as *AuthService) HasRole(userInfo *ZitadelUserInfo, requiredRole string) bool {
	for _, role := range userInfo.Roles {
		if role == requiredRole {
			return true
		}
	}
	return false
}

// HasAnyRole - Kullanıcının herhangi bir role'ü var mı kontrol et
func (as *AuthService) HasAnyRole(userInfo *ZitadelUserInfo, requiredRoles []string) bool {
	for _, userRole := range userInfo.Roles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// CreateSession - Kullanıcı için session oluştur
func (as *AuthService) CreateSession(userInfo *ZitadelUserInfo) (*Session, error) {
	return as.sessionService.CreateSession(userInfo)
}

// GetSession - Session'ı getir
func (as *AuthService) GetSession(sessionID string) (*Session, error) {
	return as.sessionService.GetSession(sessionID)
}

// DeleteSession - Session'ı sil
func (as *AuthService) DeleteSession(sessionID string) error {
	return as.sessionService.DeleteSession(sessionID)
}

// ValidateSession - Session'ı validate et
func (as *AuthService) ValidateSession(sessionID string) (*Session, error) {
	return as.sessionService.ValidateSession(sessionID)
}

// RevokeToken - Token'ı iptal et
func (as *AuthService) RevokeToken(ctx context.Context, token *oauth2.Token) error {
	revokeURL := fmt.Sprintf("%s/oauth/v2/revoke", as.config.Domain)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", revokeURL, strings.NewReader(fmt.Sprintf("token=%s", token.AccessToken)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(as.config.ClientID, as.config.ClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		as.logger.Error("Token revocation failed", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		as.logger.Error("Token revocation request failed",
			zap.Int("status_code", resp.StatusCode),
		)
		return fmt.Errorf("token revocation failed with status: %d", resp.StatusCode)
	}

	as.logger.Info("Token revoked successfully")
	return nil
}

package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fiber-app/pkg/config"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type AuthService struct {
	config      *config.ZitadelConfig
	oauthConfig *oauth2.Config
	logger      *zap.Logger
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

type TokenClaims struct {
	Sub   string   `json:"sub"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"urn:zitadel:iam:org:project:roles"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg *config.ZitadelConfig, logger *zap.Logger) *AuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/v2/authorize", cfg.Domain),
			TokenURL: fmt.Sprintf("%s/oauth/v2/token", cfg.Domain),
		},
	}

	return &AuthService{
		config:      cfg,
		oauthConfig: oauthConfig,
		logger:      logger,
	}
}

// GenerateAuthURL - OAuth2 authorization URL oluştur
func (as *AuthService) GenerateAuthURL() (string, string, error) {
	// State parameter oluştur (CSRF koruması için)
	state, err := generateRandomString(32)
	if err != nil {
		return "", "", err
	}

	url := as.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, state, nil
}

// ExchangeCodeForToken - Authorization code'u token ile değiştir
func (as *AuthService) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := as.oauthConfig.Exchange(ctx, code)
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

// ValidateToken - JWT token'ı validate et
func (as *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	// Zitadel'den public key alınması gerekir, şimdilik basit validation
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Bu gerçek implementasyonda Zitadel'in public key'i kullanılmalı
		return []byte("your-secret-key"), nil
	})

	if err != nil {
		as.logger.Error("Token validation failed", zap.Error(err))
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
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

// CreateJWTToken - Kullanıcı için JWT token oluştur
func (as *AuthService) CreateJWTToken(userInfo *ZitadelUserInfo) (string, error) {
	claims := TokenClaims{
		Sub:   userInfo.Sub,
		Name:  userInfo.Name,
		Email: userInfo.Email,
		Roles: userInfo.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "fiber-app",
			Subject:   userInfo.Sub,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		as.logger.Error("Failed to create JWT token", zap.Error(err))
		return "", err
	}

	return tokenString, nil
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

// generateRandomString - Güvenli random string oluştur
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

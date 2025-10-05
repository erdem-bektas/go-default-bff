package services

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type JWKSService struct {
	logger    *zap.Logger
	client    *http.Client
	jwksURL   string
	keys      map[string]*rsa.PublicKey
	keysMutex sync.RWMutex
	lastFetch time.Time
	cacheTTL  time.Duration
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

type TokenClaims struct {
	Sub   string   `json:"sub"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"urn:zitadel:iam:org:project:roles"`
	jwt.RegisteredClaims
}

func NewJWKSService(jwksURL string, logger *zap.Logger) *JWKSService {
	return &JWKSService{
		logger:   logger,
		client:   &http.Client{Timeout: 10 * time.Second},
		jwksURL:  jwksURL,
		keys:     make(map[string]*rsa.PublicKey),
		cacheTTL: time.Hour, // 1 hour cache
	}
}

func (s *JWKSService) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	// Parse token to get the kid
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get key ID from header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		// Get public key for this kid
		publicKey, err := s.getPublicKey(ctx, kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		return publicKey, nil
	})

	if err != nil {
		s.logger.Error("Token validation failed", zap.Error(err))
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		s.logger.Info("Token validated successfully",
			zap.String("sub", claims.Sub),
			zap.String("email", claims.Email),
		)
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *JWKSService) getPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	s.keysMutex.RLock()
	if key, exists := s.keys[kid]; exists && time.Since(s.lastFetch) < s.cacheTTL {
		s.keysMutex.RUnlock()
		return key, nil
	}
	s.keysMutex.RUnlock()

	// Need to fetch/refresh keys
	if err := s.fetchJWKS(ctx); err != nil {
		return nil, err
	}

	s.keysMutex.RLock()
	defer s.keysMutex.RUnlock()

	key, exists := s.keys[kid]
	if !exists {
		return nil, fmt.Errorf("key with kid %s not found", kid)
	}

	return key, nil
}

func (s *JWKSService) fetchJWKS(ctx context.Context) error {
	s.logger.Info("Fetching JWKS", zap.String("jwks_url", s.jwksURL))

	req, err := http.NewRequestWithContext(ctx, "GET", s.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create JWKS request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS fetch failed with status: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	s.keysMutex.Lock()
	defer s.keysMutex.Unlock()

	// Clear existing keys
	s.keys = make(map[string]*rsa.PublicKey)

	// Convert JWKs to RSA public keys
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			continue
		}

		publicKey, err := s.jwkToRSAPublicKey(jwk)
		if err != nil {
			s.logger.Warn("Failed to convert JWK to RSA public key",
				zap.String("kid", jwk.Kid),
				zap.Error(err),
			)
			continue
		}

		s.keys[jwk.Kid] = publicKey
		s.logger.Info("Added public key",
			zap.String("kid", jwk.Kid),
			zap.String("alg", jwk.Alg),
		)
	}

	s.lastFetch = time.Now()
	s.logger.Info("JWKS fetched successfully", zap.Int("key_count", len(s.keys)))

	return nil
}

func (s *JWKSService) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode n (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode e (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

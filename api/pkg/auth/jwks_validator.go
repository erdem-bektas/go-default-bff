package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// TokenClaims represents the JWT token claims with Zitadel-specific fields
type TokenClaims struct {
	Sub       string   `json:"sub"`
	Iss       string   `json:"iss"`
	Aud       []string `json:"aud"`
	Exp       int64    `json:"exp"`
	Iat       int64    `json:"iat"`
	AuthTime  int64    `json:"auth_time"`
	Nonce     string   `json:"nonce,omitempty"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Roles     []string `json:"urn:zitadel:iam:org:project:roles"`
	OrgID     string   `json:"urn:zitadel:iam:org:id"`
	ProjectID string   `json:"urn:zitadel:iam:org:project:id"`
	jwt.RegisteredClaims
}

// JWKSValidator interface for JWT validation using JWKS
type JWKSValidator interface {
	ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error)
	RefreshJWKS(ctx context.Context) error
	GetPublicKey(keyID string) (interface{}, error)
}

// JWKSValidatorConfig holds configuration for the JWKS validator
type JWKSValidatorConfig struct {
	JWKSURL            string
	Issuer             string
	Audience           string
	ProjectID          string
	RoleClaimName      string
	CacheTTL           time.Duration
	ClockSkewTolerance time.Duration
	HTTPTimeout        time.Duration
}

// jwksValidator implements the JWKSValidator interface
type jwksValidator struct {
	config    JWKSValidatorConfig
	keySet    jwk.Set
	cacheTime time.Time
	mutex     sync.RWMutex
}

// NewJWKSValidator creates a new JWKS validator instance
func NewJWKSValidator(config JWKSValidatorConfig) JWKSValidator {
	if config.CacheTTL == 0 {
		config.CacheTTL = time.Hour // Default 1 hour cache TTL
	}
	if config.ClockSkewTolerance == 0 {
		config.ClockSkewTolerance = 2 * time.Minute // Default 2 minute clock skew tolerance
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 30 * time.Second // Default 30 second HTTP timeout
	}
	if config.RoleClaimName == "" {
		config.RoleClaimName = "urn:zitadel:iam:org:project:roles"
	}

	return &jwksValidator{
		config: config,
	}
}

// ValidateToken validates a JWT token using JWKS
func (v *jwksValidator) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	// Create token claims extractor
	extractor := NewTokenClaimsExtractor(v.config.RoleClaimName, v.config.ClockSkewTolerance)

	// Define key function for JWT parsing
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Get the key ID from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("token missing kid header")
		}

		// Validate the signing method
		switch token.Method.(type) {
		case *jwt.SigningMethodRSA:
			// RS256 is supported
		case *jwt.SigningMethodECDSA:
			// ES256 is supported
		default:
			return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
		}

		// Get the public key for this kid
		publicKey, err := v.GetPublicKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key for kid %s: %w", kid, err)
		}

		return publicKey, nil
	}

	// Extract claims using the extractor
	claims, err := extractor.ExtractClaims(tokenString, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Validate issuer
	if err := extractor.ValidateIssuer(claims, v.config.Issuer); err != nil {
		return nil, err
	}

	// Validate audience with multi-tenant support
	if err := extractor.ValidateAudience(claims, v.config.Audience, v.config.ProjectID); err != nil {
		return nil, err
	}

	// Validate time claims
	if err := extractor.ValidateTimeClaims(claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// GetPublicKey retrieves a public key by key ID, refreshing JWKS if necessary
func (v *jwksValidator) GetPublicKey(keyID string) (interface{}, error) {
	v.mutex.RLock()

	// Check if we have a cached keyset and it's still valid
	if v.keySet != nil && time.Since(v.cacheTime) < v.config.CacheTTL {
		if key, found := v.keySet.LookupKeyID(keyID); found {
			v.mutex.RUnlock()

			// Convert JWK to crypto public key
			var publicKey interface{}
			if err := key.Raw(&publicKey); err != nil {
				return nil, fmt.Errorf("failed to extract public key: %w", err)
			}
			return publicKey, nil
		}
	}
	v.mutex.RUnlock()

	// Need to refresh JWKS
	ctx, cancel := context.WithTimeout(context.Background(), v.config.HTTPTimeout)
	defer cancel()

	if err := v.RefreshJWKS(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
	}

	// Try to get the key again after refresh
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	if v.keySet != nil {
		if key, found := v.keySet.LookupKeyID(keyID); found {
			// Convert JWK to crypto public key
			var publicKey interface{}
			if err := key.Raw(&publicKey); err != nil {
				return nil, fmt.Errorf("failed to extract public key: %w", err)
			}
			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("key with ID %s not found in JWKS", keyID)
}

// RefreshJWKS fetches and caches the JWKS from the configured URL
func (v *jwksValidator) RefreshJWKS(ctx context.Context) error {
	// Create a context with timeout for the HTTP request
	fetchCtx, cancel := context.WithTimeout(ctx, v.config.HTTPTimeout)
	defer cancel()

	// Fetch JWKS using the jwx library
	keySet, err := jwk.Fetch(fetchCtx, v.config.JWKSURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS from %s: %w", v.config.JWKSURL, err)
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.keySet = keySet
	v.cacheTime = time.Now()

	return nil
}

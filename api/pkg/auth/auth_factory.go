package auth

import (
	"context"
	"time"
)

// AuthConfig represents the authentication configuration
type AuthConfig struct {
	JWKSURL            string
	Issuer             string
	Audience           string
	ProjectID          string
	OrgID              string
	RoleClaimName      string
	CacheTTL           time.Duration
	ClockSkewTolerance time.Duration
	HTTPTimeout        time.Duration
}

// NewJWKSValidatorFromConfig creates a JWKS validator from auth configuration
func NewJWKSValidatorFromConfig(config AuthConfig) JWKSValidator {
	validatorConfig := JWKSValidatorConfig{
		JWKSURL:            config.JWKSURL,
		Issuer:             config.Issuer,
		Audience:           config.Audience,
		ProjectID:          config.ProjectID,
		RoleClaimName:      config.RoleClaimName,
		CacheTTL:           config.CacheTTL,
		ClockSkewTolerance: config.ClockSkewTolerance,
		HTTPTimeout:        config.HTTPTimeout,
	}

	return NewJWKSValidator(validatorConfig)
}

// ValidateTokenWithContext validates a token with additional context validation
func ValidateTokenWithContext(validator JWKSValidator, tokenString string, expectedOrgID, expectedProjectID string) (*TokenClaims, error) {
	claims, err := validator.ValidateToken(context.TODO(), tokenString)
	if err != nil {
		return nil, err
	}

	// Create extractor for additional validations
	extractor := NewTokenClaimsExtractor("", 0) // Use defaults

	// Validate organization context if specified
	if expectedOrgID != "" {
		if err := extractor.ValidateOrgContext(claims, expectedOrgID); err != nil {
			return nil, err
		}
	}

	// Validate project context if specified
	if expectedProjectID != "" {
		if err := extractor.ValidateProjectContext(claims, expectedProjectID); err != nil {
			return nil, err
		}
	}

	return claims, nil
}

// ExtractUserInfoFromToken extracts user information from a validated token
func ExtractUserInfoFromToken(validator JWKSValidator, tokenString string) (*ZitadelUserInfo, error) {
	claims, err := validator.ValidateToken(context.TODO(), tokenString)
	if err != nil {
		return nil, err
	}

	extractor := NewTokenClaimsExtractor("", 0)
	return extractor.ExtractUserInfo(claims), nil
}

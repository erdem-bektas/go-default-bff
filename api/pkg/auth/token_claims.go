package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ZitadelUserInfo represents user information from Zitadel
type ZitadelUserInfo struct {
	Sub               string              `json:"sub"`
	Name              string              `json:"name"`
	GivenName         string              `json:"given_name"`
	FamilyName        string              `json:"family_name"`
	PreferredUsername string              `json:"preferred_username"`
	Email             string              `json:"email"`
	EmailVerified     bool                `json:"email_verified"`
	Roles             []string            `json:"urn:zitadel:iam:org:project:roles"`
	OrgID             string              `json:"urn:zitadel:iam:org:id"`
	ProjectRoles      map[string][]string `json:"urn:zitadel:iam:org:project:roles:audience"`
}

// TokenClaimsExtractor handles extraction and validation of token claims
type TokenClaimsExtractor struct {
	roleClaimName      string
	clockSkewTolerance time.Duration
}

// NewTokenClaimsExtractor creates a new token claims extractor
func NewTokenClaimsExtractor(roleClaimName string, clockSkewTolerance time.Duration) *TokenClaimsExtractor {
	if roleClaimName == "" {
		roleClaimName = "urn:zitadel:iam:org:project:roles"
	}
	if clockSkewTolerance == 0 {
		clockSkewTolerance = 2 * time.Minute
	}

	return &TokenClaimsExtractor{
		roleClaimName:      roleClaimName,
		clockSkewTolerance: clockSkewTolerance,
	}
}

// ExtractClaims extracts and validates claims from a JWT token string
func (e *TokenClaimsExtractor) ExtractClaims(tokenString string, keyFunc jwt.Keyfunc) (*TokenClaims, error) {
	// Parse token with custom claims
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token or claims")
	}

	// If using a custom role claim name, extract roles from MapClaims
	if e.roleClaimName != "urn:zitadel:iam:org:project:roles" {
		mapClaims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(tokenString, &mapClaims, keyFunc)
		if err == nil {
			if roles, exists := mapClaims[e.roleClaimName]; exists {
				claims.Roles = e.extractRolesFromInterface(roles)
			}
		}
	}

	return claims, nil
}

// ValidateTimeClaims validates the time-based claims with clock skew tolerance
func (e *TokenClaimsExtractor) ValidateTimeClaims(claims *TokenClaims) error {
	now := time.Now()

	// Check expiration (exp)
	if claims.Exp > 0 {
		expTime := time.Unix(claims.Exp, 0)
		if now.After(expTime.Add(e.clockSkewTolerance)) {
			return fmt.Errorf("token expired at %v (current time: %v, tolerance: %v)",
				expTime, now, e.clockSkewTolerance)
		}
	}

	// Check issued at (iat)
	if claims.Iat > 0 {
		iatTime := time.Unix(claims.Iat, 0)
		if now.Before(iatTime.Add(-e.clockSkewTolerance)) {
			return fmt.Errorf("token used before issued at %v (current time: %v, tolerance: %v)",
				iatTime, now, e.clockSkewTolerance)
		}
	}

	// Check not before (nbf) if present
	if claims.NotBefore != nil && claims.NotBefore.Unix() > 0 {
		nbfTime := claims.NotBefore.Time
		if now.Before(nbfTime.Add(-e.clockSkewTolerance)) {
			return fmt.Errorf("token not valid before %v (current time: %v, tolerance: %v)",
				nbfTime, now, e.clockSkewTolerance)
		}
	}

	return nil
}

// ValidateIssuer validates the issuer claim
func (e *TokenClaimsExtractor) ValidateIssuer(claims *TokenClaims, expectedIssuer string) error {
	if claims.Iss != expectedIssuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, claims.Iss)
	}
	return nil
}

// ValidateAudience validates the audience claim with multi-tenant support
func (e *TokenClaimsExtractor) ValidateAudience(claims *TokenClaims, expectedAudience, projectID string) error {
	expectedAudiences := []string{expectedAudience}

	// Add fallback audience if using client_id@project format
	if projectID != "" {
		// Extract client_id from client_id@project format
		clientID := expectedAudience
		projectSuffix := "@" + projectID
		if len(clientID) > len(projectSuffix) && clientID[len(clientID)-len(projectSuffix):] == projectSuffix {
			expectedAudiences = append(expectedAudiences, clientID[:len(clientID)-len(projectSuffix)])
		}
	}

	// Check if any of the token audiences match our expected audiences
	for _, tokenAud := range claims.Aud {
		for _, expectedAud := range expectedAudiences {
			if tokenAud == expectedAud {
				return nil
			}
		}
	}

	return fmt.Errorf("invalid audience: expected one of %v, got %v", expectedAudiences, claims.Aud)
}

// ExtractUserInfo extracts user information from token claims
func (e *TokenClaimsExtractor) ExtractUserInfo(claims *TokenClaims) *ZitadelUserInfo {
	return &ZitadelUserInfo{
		Sub:          claims.Sub,
		Name:         claims.Name,
		Email:        claims.Email,
		Roles:        claims.Roles,
		OrgID:        claims.OrgID,
		ProjectRoles: make(map[string][]string), // This would need to be populated from additional claims
	}
}

// extractRolesFromInterface safely extracts roles from an interface{} value
func (e *TokenClaimsExtractor) extractRolesFromInterface(roles interface{}) []string {
	switch v := roles.(type) {
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, role := range v {
			if roleStr, ok := role.(string); ok {
				result = append(result, roleStr)
			}
		}
		return result
	case []string:
		return v
	case string:
		return []string{v}
	default:
		return []string{}
	}
}

// ValidateNonce validates the nonce claim if present
func (e *TokenClaimsExtractor) ValidateNonce(claims *TokenClaims, expectedNonce string) error {
	if expectedNonce != "" && claims.Nonce != expectedNonce {
		return fmt.Errorf("invalid nonce: expected %s, got %s", expectedNonce, claims.Nonce)
	}
	return nil
}

// ValidateAuthTime validates the auth_time claim with tolerance
func (e *TokenClaimsExtractor) ValidateAuthTime(claims *TokenClaims, maxAge time.Duration) error {
	if maxAge > 0 && claims.AuthTime > 0 {
		authTime := time.Unix(claims.AuthTime, 0)
		if time.Since(authTime) > maxAge {
			return fmt.Errorf("authentication too old: auth_time %v, max_age %v", authTime, maxAge)
		}
	}
	return nil
}

// GetProjectRoles extracts project-specific roles from claims
func (e *TokenClaimsExtractor) GetProjectRoles(claims *TokenClaims, projectID string) []string {
	// This is a simplified implementation
	// In a real Zitadel setup, you might have project-specific role claims
	// that need to be extracted based on the project ID

	// For now, return all roles if they match the project context
	if claims.ProjectID == projectID {
		return claims.Roles
	}

	return []string{}
}

// ValidateOrgContext validates that the token is valid for the given organization
func (e *TokenClaimsExtractor) ValidateOrgContext(claims *TokenClaims, expectedOrgID string) error {
	if expectedOrgID != "" && claims.OrgID != expectedOrgID {
		return fmt.Errorf("invalid organization context: expected %s, got %s", expectedOrgID, claims.OrgID)
	}
	return nil
}

// ValidateProjectContext validates that the token is valid for the given project
func (e *TokenClaimsExtractor) ValidateProjectContext(claims *TokenClaims, expectedProjectID string) error {
	if expectedProjectID != "" && claims.ProjectID != expectedProjectID {
		return fmt.Errorf("invalid project context: expected %s, got %s", expectedProjectID, claims.ProjectID)
	}
	return nil
}

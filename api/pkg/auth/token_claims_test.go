package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenClaimsExtractor_ExtractClaims_Success(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	// Create test claims
	now := time.Now()
	expectedClaims := &TokenClaims{
		Sub:       "test-user-123",
		Iss:       "https://test.zitadel.cloud",
		Aud:       []string{"test-client@test-project"},
		Exp:       now.Add(time.Hour).Unix(),
		Iat:       now.Unix(),
		AuthTime:  now.Unix(),
		Email:     "test@example.com",
		Name:      "Test User",
		Roles:     []string{"user", "admin"},
		OrgID:     "test-org-123",
		ProjectID: "test-project-456",
	}

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, expectedClaims)
	token.Header["kid"] = testKeyID
	tokenString, err := token.SignedString(testPrivateKey)
	require.NoError(t, err)

	// Define key function
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return &testPrivateKey.PublicKey, nil
	}

	// Extract claims
	claims, err := extractor.ExtractClaims(tokenString, keyFunc)
	require.NoError(t, err)

	// Verify claims
	assert.Equal(t, expectedClaims.Sub, claims.Sub)
	assert.Equal(t, expectedClaims.Iss, claims.Iss)
	assert.Equal(t, expectedClaims.Aud, claims.Aud)
	assert.Equal(t, expectedClaims.Email, claims.Email)
	assert.Equal(t, expectedClaims.Name, claims.Name)
	assert.Equal(t, expectedClaims.Roles, claims.Roles)
	assert.Equal(t, expectedClaims.OrgID, claims.OrgID)
	assert.Equal(t, expectedClaims.ProjectID, claims.ProjectID)
}

func TestTokenClaimsExtractor_ExtractClaims_CustomRoleClaim(t *testing.T) {
	// Use custom role claim name
	customRoleClaimName := "custom:roles"
	extractor := NewTokenClaimsExtractor(customRoleClaimName, 2*time.Minute)

	// Create token with custom role claim
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":                            "test-user-123",
		"iss":                            "https://test.zitadel.cloud",
		"aud":                            []string{"test-client@test-project"},
		"exp":                            now.Add(time.Hour).Unix(),
		"iat":                            now.Unix(),
		"auth_time":                      now.Unix(),
		"email":                          "test@example.com",
		"name":                           "Test User",
		customRoleClaimName:              []string{"custom-role1", "custom-role2"},
		"urn:zitadel:iam:org:id":         "test-org-123",
		"urn:zitadel:iam:org:project:id": "test-project-456",
	}

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID
	tokenString, err := token.SignedString(testPrivateKey)
	require.NoError(t, err)

	// Define key function
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return &testPrivateKey.PublicKey, nil
	}

	// Extract claims
	extractedClaims, err := extractor.ExtractClaims(tokenString, keyFunc)
	require.NoError(t, err)

	// Verify custom roles were extracted
	assert.Equal(t, []string{"custom-role1", "custom-role2"}, extractedClaims.Roles)
}

func TestTokenClaimsExtractor_ExtractClaims_InvalidToken(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	// Define key function
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return &testPrivateKey.PublicKey, nil
	}

	// Test with invalid token string
	_, err := extractor.ExtractClaims("invalid.token.string", keyFunc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestTokenClaimsExtractor_ValidateTimeClaims_Success(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	now := time.Now()
	claims := &TokenClaims{
		Exp:      now.Add(time.Hour).Unix(),
		Iat:      now.Add(-time.Minute).Unix(),
		AuthTime: now.Add(-time.Minute).Unix(),
	}
	claims.NotBefore = jwt.NewNumericDate(now.Add(-time.Minute))

	err := extractor.ValidateTimeClaims(claims)
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_ValidateTimeClaims_ExpiredToken(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	now := time.Now()
	claims := &TokenClaims{
		Exp: now.Add(-time.Hour).Unix(), // Expired 1 hour ago
		Iat: now.Add(-2 * time.Hour).Unix(),
	}

	err := extractor.ValidateTimeClaims(claims)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token expired")
}

func TestTokenClaimsExtractor_ValidateTimeClaims_ClockSkewTolerance(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 5*time.Minute)

	now := time.Now()
	claims := &TokenClaims{
		Exp: now.Add(-3 * time.Minute).Unix(), // Expired 3 minutes ago, within tolerance
		Iat: now.Add(-time.Hour).Unix(),
	}

	err := extractor.ValidateTimeClaims(claims)
	assert.NoError(t, err, "Should accept token within clock skew tolerance")

	// Test outside tolerance
	claims.Exp = now.Add(-7 * time.Minute).Unix() // Expired 7 minutes ago, outside tolerance
	err = extractor.ValidateTimeClaims(claims)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token expired")
}

func TestTokenClaimsExtractor_ValidateTimeClaims_IssuedInFuture(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	now := time.Now()
	claims := &TokenClaims{
		Exp: now.Add(time.Hour).Unix(),
		Iat: now.Add(5 * time.Minute).Unix(), // Issued 5 minutes in the future, outside tolerance
	}

	err := extractor.ValidateTimeClaims(claims)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token used before issued")
}

func TestTokenClaimsExtractor_ValidateTimeClaims_NotBeforeInFuture(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	now := time.Now()
	claims := &TokenClaims{
		Exp: now.Add(time.Hour).Unix(),
		Iat: now.Add(-time.Minute).Unix(),
	}
	claims.NotBefore = jwt.NewNumericDate(now.Add(5 * time.Minute)) // Not valid for 5 minutes

	err := extractor.ValidateTimeClaims(claims)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token not valid before")
}

func TestTokenClaimsExtractor_ValidateIssuer_Success(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Iss: "https://test.zitadel.cloud",
	}

	err := extractor.ValidateIssuer(claims, "https://test.zitadel.cloud")
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_ValidateIssuer_Mismatch(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Iss: "https://malicious.issuer.com",
	}

	err := extractor.ValidateIssuer(claims, "https://test.zitadel.cloud")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid issuer")
	assert.Contains(t, err.Error(), "expected https://test.zitadel.cloud")
	assert.Contains(t, err.Error(), "got https://malicious.issuer.com")
}

func TestTokenClaimsExtractor_ValidateAudience_ExactMatch(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Aud: []string{"test-client@test-project"},
	}

	err := extractor.ValidateAudience(claims, "test-client@test-project", "test-project")
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_ValidateAudience_FallbackMatch(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Aud: []string{"test-client"}, // Without @project suffix
	}

	err := extractor.ValidateAudience(claims, "test-client@test-project", "test-project")
	assert.NoError(t, err, "Should accept fallback audience (client_id only)")
}

func TestTokenClaimsExtractor_ValidateAudience_MultipleAudiences(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Aud: []string{"other-client", "test-client@test-project", "another-client"},
	}

	err := extractor.ValidateAudience(claims, "test-client@test-project", "test-project")
	assert.NoError(t, err, "Should match one of multiple audiences")
}

func TestTokenClaimsExtractor_ValidateAudience_NoMatch(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Aud: []string{"wrong-client@wrong-project"},
	}

	err := extractor.ValidateAudience(claims, "test-client@test-project", "test-project")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid audience")
}

func TestTokenClaimsExtractor_ValidateAudience_EmptyAudience(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Aud: []string{},
	}

	err := extractor.ValidateAudience(claims, "test-client@test-project", "test-project")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid audience")
}

func TestTokenClaimsExtractor_ValidateNonce_Success(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Nonce: "test-nonce-123",
	}

	err := extractor.ValidateNonce(claims, "test-nonce-123")
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_ValidateNonce_Mismatch(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Nonce: "wrong-nonce",
	}

	err := extractor.ValidateNonce(claims, "expected-nonce")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nonce")
}

func TestTokenClaimsExtractor_ValidateNonce_EmptyExpected(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Nonce: "some-nonce",
	}

	// When expected nonce is empty, validation should pass
	err := extractor.ValidateNonce(claims, "")
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_ValidateAuthTime_Success(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	now := time.Now()
	claims := &TokenClaims{
		AuthTime: now.Add(-5 * time.Minute).Unix(), // Authenticated 5 minutes ago
	}

	err := extractor.ValidateAuthTime(claims, 10*time.Minute) // Max age 10 minutes
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_ValidateAuthTime_TooOld(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	now := time.Now()
	claims := &TokenClaims{
		AuthTime: now.Add(-15 * time.Minute).Unix(), // Authenticated 15 minutes ago
	}

	err := extractor.ValidateAuthTime(claims, 10*time.Minute) // Max age 10 minutes
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication too old")
}

func TestTokenClaimsExtractor_ValidateAuthTime_NoMaxAge(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	now := time.Now()
	claims := &TokenClaims{
		AuthTime: now.Add(-time.Hour).Unix(), // Authenticated 1 hour ago
	}

	// When max age is 0, validation should pass
	err := extractor.ValidateAuthTime(claims, 0)
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_ValidateOrgContext_Success(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		OrgID: "test-org-123",
	}

	err := extractor.ValidateOrgContext(claims, "test-org-123")
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_ValidateOrgContext_Mismatch(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		OrgID: "wrong-org-456",
	}

	err := extractor.ValidateOrgContext(claims, "test-org-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid organization context")
}

func TestTokenClaimsExtractor_ValidateOrgContext_EmptyExpected(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		OrgID: "some-org-123",
	}

	// When expected org ID is empty, validation should pass
	err := extractor.ValidateOrgContext(claims, "")
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_ValidateProjectContext_Success(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		ProjectID: "test-project-456",
	}

	err := extractor.ValidateProjectContext(claims, "test-project-456")
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_ValidateProjectContext_Mismatch(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		ProjectID: "wrong-project-789",
	}

	err := extractor.ValidateProjectContext(claims, "test-project-456")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project context")
}

func TestTokenClaimsExtractor_ValidateProjectContext_EmptyExpected(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		ProjectID: "some-project-456",
	}

	// When expected project ID is empty, validation should pass
	err := extractor.ValidateProjectContext(claims, "")
	assert.NoError(t, err)
}

func TestTokenClaimsExtractor_GetProjectRoles_MatchingProject(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		ProjectID: "test-project-456",
		Roles:     []string{"admin", "user", "viewer"},
	}

	roles := extractor.GetProjectRoles(claims, "test-project-456")
	assert.Equal(t, []string{"admin", "user", "viewer"}, roles)
}

func TestTokenClaimsExtractor_GetProjectRoles_NonMatchingProject(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		ProjectID: "different-project-789",
		Roles:     []string{"admin", "user", "viewer"},
	}

	roles := extractor.GetProjectRoles(claims, "test-project-456")
	assert.Equal(t, []string{}, roles)
}

func TestTokenClaimsExtractor_ExtractUserInfo(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	claims := &TokenClaims{
		Sub:   "test-user-123",
		Name:  "Test User",
		Email: "test@example.com",
		Roles: []string{"admin", "user"},
		OrgID: "test-org-123",
	}

	userInfo := extractor.ExtractUserInfo(claims)
	assert.Equal(t, "test-user-123", userInfo.Sub)
	assert.Equal(t, "Test User", userInfo.Name)
	assert.Equal(t, "test@example.com", userInfo.Email)
	assert.Equal(t, []string{"admin", "user"}, userInfo.Roles)
	assert.Equal(t, "test-org-123", userInfo.OrgID)
	assert.NotNil(t, userInfo.ProjectRoles)
}

func TestTokenClaimsExtractor_ExtractRolesFromInterface(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "string slice",
			input:    []string{"role1", "role2", "role3"},
			expected: []string{"role1", "role2", "role3"},
		},
		{
			name:     "interface slice with strings",
			input:    []interface{}{"role1", "role2", "role3"},
			expected: []string{"role1", "role2", "role3"},
		},
		{
			name:     "interface slice with mixed types",
			input:    []interface{}{"role1", 123, "role2"},
			expected: []string{"role1", "role2"}, // Non-strings filtered out
		},
		{
			name:     "single string",
			input:    "single-role",
			expected: []string{"single-role"},
		},
		{
			name:     "unsupported type",
			input:    123,
			expected: []string{},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.extractRolesFromInterface(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenClaimsExtractor_DefaultValues(t *testing.T) {
	// Test with empty role claim name
	extractor1 := NewTokenClaimsExtractor("", 0)
	assert.Equal(t, "urn:zitadel:iam:org:project:roles", extractor1.roleClaimName)
	assert.Equal(t, 2*time.Minute, extractor1.clockSkewTolerance)

	// Test with custom values
	extractor2 := NewTokenClaimsExtractor("custom:roles", 5*time.Minute)
	assert.Equal(t, "custom:roles", extractor2.roleClaimName)
	assert.Equal(t, 5*time.Minute, extractor2.clockSkewTolerance)
}

// Edge case tests for audience validation with various project ID formats
func TestTokenClaimsExtractor_ValidateAudience_EdgeCases(t *testing.T) {
	extractor := NewTokenClaimsExtractor("urn:zitadel:iam:org:project:roles", 2*time.Minute)

	tests := []struct {
		name             string
		tokenAudiences   []string
		expectedAudience string
		projectID        string
		shouldPass       bool
	}{
		{
			name:             "exact match with project suffix",
			tokenAudiences:   []string{"client@project"},
			expectedAudience: "client@project",
			projectID:        "project",
			shouldPass:       true,
		},
		{
			name:             "fallback match without project suffix",
			tokenAudiences:   []string{"client"},
			expectedAudience: "client@project",
			projectID:        "project",
			shouldPass:       true,
		},
		{
			name:             "no project ID provided",
			tokenAudiences:   []string{"client"},
			expectedAudience: "client",
			projectID:        "",
			shouldPass:       true,
		},
		{
			name:             "malformed expected audience",
			tokenAudiences:   []string{"client"},
			expectedAudience: "@project", // Malformed
			projectID:        "project",
			shouldPass:       false,
		},
		{
			name:             "empty project ID with @ in audience",
			tokenAudiences:   []string{"client@project"},
			expectedAudience: "client@project",
			projectID:        "", // No project ID
			shouldPass:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &TokenClaims{
				Aud: tt.tokenAudiences,
			}

			err := extractor.ValidateAudience(claims, tt.expectedAudience, tt.projectID)
			if tt.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

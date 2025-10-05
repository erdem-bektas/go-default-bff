package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test RSA key pair for testing
var (
	testPrivateKey *rsa.PrivateKey
	testPublicKey  *rsa.PublicKey
	testKeyID      = "test-key-id-1"
	testKeyID2     = "test-key-id-2"
)

// Initialize test keys
func init() {
	var err error
	testPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate test RSA key: %v", err))
	}
	testPrivateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate second test RSA key: %v", err))
	}
	testPublicKey = &testPrivateKey.PublicKey
	testPublicKey2 = &testPrivateKey2.PublicKey

	// Store both keys for rotation testing
	testKeys = map[string]*rsa.PrivateKey{
		testKeyID:  testPrivateKey,
		testKeyID2: testPrivateKey2,
	}
}

var (
	testPrivateKey2 *rsa.PrivateKey
	testPublicKey2  *rsa.PublicKey
	testKeys        map[string]*rsa.PrivateKey
)

// Helper functions to encode RSA public key components
func encodeRSAPublicKeyN(key *rsa.PublicKey) string {
	return base64.RawURLEncoding.EncodeToString(key.N.Bytes())
}

func encodeRSAPublicKeyE(key *rsa.PublicKey) string {
	eBytes := make([]byte, 4)
	eBytes[0] = byte(key.E >> 24)
	eBytes[1] = byte(key.E >> 16)
	eBytes[2] = byte(key.E >> 8)
	eBytes[3] = byte(key.E)

	// Remove leading zeros
	for len(eBytes) > 1 && eBytes[0] == 0 {
		eBytes = eBytes[1:]
	}

	return base64.RawURLEncoding.EncodeToString(eBytes)
}

// createTestJWKS creates a test JWKS response with the given key IDs
func createTestJWKS(keyIDs ...string) map[string]interface{} {
	keys := make([]map[string]interface{}, 0, len(keyIDs))

	for _, kid := range keyIDs {
		var publicKey *rsa.PublicKey
		if kid == testKeyID {
			publicKey = testPublicKey
		} else if kid == testKeyID2 {
			publicKey = testPublicKey2
		} else {
			continue // Skip unknown key IDs
		}

		// Convert RSA public key to JWK format
		key := map[string]interface{}{
			"kty": "RSA",
			"use": "sig",
			"kid": kid,
			"alg": "RS256",
			"n":   encodeRSAPublicKeyN(publicKey),
			"e":   encodeRSAPublicKeyE(publicKey),
		}
		keys = append(keys, key)
	}

	return map[string]interface{}{
		"keys": keys,
	}
}

// createTestJWKSServer creates a test HTTP server that serves JWKS
func createTestJWKSServer(t testing.TB, keyIDs ...string) *httptest.Server {
	jwks := createTestJWKS(keyIDs...)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
}

// createTestToken creates a test JWT token with the given claims and key ID
func createTestToken(t testing.TB, claims *TokenClaims, keyID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID

	privateKey, exists := testKeys[keyID]
	if !exists {
		t.Fatalf("Test private key not found for kid: %s", keyID)
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	return tokenString
}

// createValidTestClaims creates valid test claims
func createValidTestClaims() *TokenClaims {
	now := time.Now()
	return &TokenClaims{
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
}

func TestJWKSValidator_ValidateToken_Success(t *testing.T) {
	// Create test JWKS server
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create valid test token
	claims := createValidTestClaims()
	tokenString := createTestToken(t, claims, testKeyID)

	// Validate token
	ctx := context.Background()
	validatedClaims, err := validator.ValidateToken(ctx, tokenString)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, claims.Sub, validatedClaims.Sub)
	assert.Equal(t, claims.Iss, validatedClaims.Iss)
	assert.Equal(t, claims.Aud, validatedClaims.Aud)
	assert.Equal(t, claims.Email, validatedClaims.Email)
	assert.Equal(t, claims.Name, validatedClaims.Name)
	assert.Equal(t, claims.Roles, validatedClaims.Roles)
	assert.Equal(t, claims.OrgID, validatedClaims.OrgID)
	assert.Equal(t, claims.ProjectID, validatedClaims.ProjectID)
}

func TestJWKSValidator_ValidateToken_ExpiredToken(t *testing.T) {
	// Create test JWKS server
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create expired token
	claims := createValidTestClaims()
	claims.Exp = time.Now().Add(-time.Hour).Unix() // Expired 1 hour ago
	tokenString := createTestToken(t, claims, testKeyID)

	// Validate token
	ctx := context.Background()
	_, err := validator.ValidateToken(ctx, tokenString)

	// Should fail with expiration error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token expired")
}

func TestJWKSValidator_ValidateToken_ClockSkewTolerance(t *testing.T) {
	// Create test JWKS server
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator with 5 minute clock skew tolerance
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 5 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create token that expired 3 minutes ago (within tolerance)
	claims := createValidTestClaims()
	claims.Exp = time.Now().Add(-3 * time.Minute).Unix()
	tokenString := createTestToken(t, claims, testKeyID)

	// Validate token - should succeed due to clock skew tolerance
	ctx := context.Background()
	_, err := validator.ValidateToken(ctx, tokenString)
	require.NoError(t, err)

	// Create token that expired 7 minutes ago (outside tolerance)
	claims.Exp = time.Now().Add(-7 * time.Minute).Unix()
	tokenString = createTestToken(t, claims, testKeyID)

	// Validate token - should fail
	_, err = validator.ValidateToken(ctx, tokenString)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token expired")
}

func TestJWKSValidator_ValidateToken_InvalidIssuer(t *testing.T) {
	// Create test JWKS server
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://expected.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create token with different issuer
	claims := createValidTestClaims()
	claims.Iss = "https://malicious.issuer.com"
	tokenString := createTestToken(t, claims, testKeyID)

	// Validate token
	ctx := context.Background()
	_, err := validator.ValidateToken(ctx, tokenString)

	// Should fail with issuer validation error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid issuer")
}

func TestJWKSValidator_ValidateToken_InvalidAudience(t *testing.T) {
	// Create test JWKS server
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "expected-client@expected-project",
		ProjectID:          "expected-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create token with different audience
	claims := createValidTestClaims()
	claims.Aud = []string{"wrong-client@wrong-project"}
	tokenString := createTestToken(t, claims, testKeyID)

	// Validate token
	ctx := context.Background()
	_, err := validator.ValidateToken(ctx, tokenString)

	// Should fail with audience validation error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid audience")
}

func TestJWKSValidator_ValidateToken_AudienceFallback(t *testing.T) {
	// Create test JWKS server
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator with client_id@project format
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create token with fallback audience (client_id only)
	claims := createValidTestClaims()
	claims.Aud = []string{"test-client"} // Without @project suffix
	tokenString := createTestToken(t, claims, testKeyID)

	// Validate token - should succeed with fallback
	ctx := context.Background()
	_, err := validator.ValidateToken(ctx, tokenString)
	require.NoError(t, err)
}

func TestJWKSValidator_ValidateToken_MissingKID(t *testing.T) {
	// Create test JWKS server
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create token without kid header
	claims := createValidTestClaims()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// Don't set kid header
	tokenString, err := token.SignedString(testPrivateKey)
	require.NoError(t, err)

	// Validate token
	ctx := context.Background()
	_, err = validator.ValidateToken(ctx, tokenString)

	// Should fail with missing kid error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token missing kid header")
}

func TestJWKSValidator_ValidateToken_UnsupportedSigningMethod(t *testing.T) {
	// Create test JWKS server
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create token with HMAC signing method (unsupported)
	claims := createValidTestClaims()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = testKeyID
	tokenString, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)

	// Validate token
	ctx := context.Background()
	_, err = validator.ValidateToken(ctx, tokenString)

	// Should fail with unsupported signing method error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported signing method")
}

func TestJWKSValidator_JWKSCaching(t *testing.T) {
	// Create test JWKS server with request counter
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		jwks := createTestJWKS(testKeyID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	// Create validator with short cache TTL for testing
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           100 * time.Millisecond, // Very short for testing
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create test token
	claims := createValidTestClaims()
	tokenString := createTestToken(t, claims, testKeyID)

	ctx := context.Background()

	// First validation - should fetch JWKS
	_, err := validator.ValidateToken(ctx, tokenString)
	require.NoError(t, err)
	assert.Equal(t, 1, requestCount, "Should have made 1 JWKS request")

	// Second validation immediately - should use cache
	_, err = validator.ValidateToken(ctx, tokenString)
	require.NoError(t, err)
	assert.Equal(t, 1, requestCount, "Should still have made only 1 JWKS request (cached)")

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Third validation - should fetch JWKS again
	_, err = validator.ValidateToken(ctx, tokenString)
	require.NoError(t, err)
	assert.Equal(t, 2, requestCount, "Should have made 2 JWKS requests (cache expired)")
}

func TestJWKSValidator_KIDRotation(t *testing.T) {
	// Create test JWKS server that initially serves only the first key
	currentKeys := []string{testKeyID}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwks := createTestJWKS(currentKeys...)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour, // Long cache TTL
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	ctx := context.Background()

	// First, validate a token with the first key (should populate cache)
	claims1 := createValidTestClaims()
	tokenString1 := createTestToken(t, claims1, testKeyID)
	_, err := validator.ValidateToken(ctx, tokenString1)
	require.NoError(t, err)

	// Now simulate key rotation - server now serves both keys
	currentKeys = []string{testKeyID, testKeyID2}

	// Create token with the new key ID
	claims2 := createValidTestClaims()
	tokenString2 := createTestToken(t, claims2, testKeyID2)

	// Validate token with new key - should trigger JWKS refresh
	_, err = validator.ValidateToken(ctx, tokenString2)
	require.NoError(t, err, "Should successfully validate token with new key after rotation")

	// Verify that old key still works (should be in refreshed cache)
	_, err = validator.ValidateToken(ctx, tokenString1)
	require.NoError(t, err, "Should still validate token with old key after rotation")
}

func TestJWKSValidator_KIDNotFound(t *testing.T) {
	// Create test JWKS server with only one key
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create token with non-existent key ID manually (not using createTestToken)
	claims := createValidTestClaims()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "non-existent-key-id"

	// Sign with existing key but claim it's from non-existent key
	tokenString, err := token.SignedString(testPrivateKey)
	require.NoError(t, err)

	// Validate token
	ctx := context.Background()
	_, err = validator.ValidateToken(ctx, tokenString)

	// Should fail with key not found error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key with ID non-existent-key-id not found")
}

func TestJWKSValidator_RefreshJWKS_HTTPError(t *testing.T) {
	// Create validator with invalid JWKS URL
	config := JWKSValidatorConfig{
		JWKSURL:            "http://invalid-url-that-does-not-exist.com/jwks",
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        1 * time.Second, // Short timeout
	}
	validator := NewJWKSValidator(config)

	// Try to refresh JWKS
	ctx := context.Background()
	err := validator.RefreshJWKS(ctx)

	// Should fail with HTTP error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch JWKS")
}

func TestJWKSValidator_RefreshJWKS_Timeout(t *testing.T) {
	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than timeout
		jwks := createTestJWKS(testKeyID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	// Create validator with short timeout
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        500 * time.Millisecond, // Short timeout
	}
	validator := NewJWKSValidator(config)

	// Try to refresh JWKS
	ctx := context.Background()
	err := validator.RefreshJWKS(ctx)

	// Should fail with timeout error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch JWKS")
}

func TestJWKSValidator_GetPublicKey_CacheHit(t *testing.T) {
	// Create test JWKS server
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		jwks := createTestJWKS(testKeyID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// First call - should fetch JWKS
	key1, err := validator.GetPublicKey(testKeyID)
	require.NoError(t, err)
	require.NotNil(t, key1)
	assert.Equal(t, 1, requestCount)

	// Second call - should use cache
	key2, err := validator.GetPublicKey(testKeyID)
	require.NoError(t, err)
	require.NotNil(t, key2)
	assert.Equal(t, 1, requestCount, "Should not have made additional JWKS request")

	// Keys should be the same
	assert.Equal(t, key1, key2)
}

func TestJWKSValidator_GetPublicKey_CacheMiss(t *testing.T) {
	// Create test JWKS server
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Try to get key that doesn't exist
	_, err := validator.GetPublicKey("non-existent-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key with ID non-existent-key not found")
}

func TestJWKSValidator_ConcurrentAccess(t *testing.T) {
	// Create test JWKS server
	server := createTestJWKSServer(t, testKeyID)
	defer server.Close()

	// Create validator
	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	// Create test token
	claims := createValidTestClaims()
	tokenString := createTestToken(t, claims, testKeyID)

	ctx := context.Background()

	// Run multiple goroutines concurrently
	const numGoroutines = 10
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := validator.ValidateToken(ctx, tokenString)
			errors <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		assert.NoError(t, err, "Concurrent validation should succeed")
	}
}

func TestJWKSValidator_DefaultConfiguration(t *testing.T) {
	// Test that default values are set correctly
	config := JWKSValidatorConfig{
		JWKSURL:  "https://test.example.com/jwks",
		Issuer:   "https://test.example.com",
		Audience: "test-client",
	}

	validator := NewJWKSValidator(config)

	// Cast to concrete type to access config
	v := validator.(*jwksValidator)

	// Check default values
	assert.Equal(t, time.Hour, v.config.CacheTTL)
	assert.Equal(t, 2*time.Minute, v.config.ClockSkewTolerance)
	assert.Equal(t, 30*time.Second, v.config.HTTPTimeout)
	assert.Equal(t, "urn:zitadel:iam:org:project:roles", v.config.RoleClaimName)
}

// Benchmark tests
func BenchmarkJWKSValidator_ValidateToken_CacheHit(b *testing.B) {
	// Setup
	server := createTestJWKSServer(b, testKeyID)
	defer server.Close()

	config := JWKSValidatorConfig{
		JWKSURL:            server.URL,
		Issuer:             "https://test.zitadel.cloud",
		Audience:           "test-client@test-project",
		ProjectID:          "test-project",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}
	validator := NewJWKSValidator(config)

	claims := createValidTestClaims()
	tokenString := createTestToken(b, claims, testKeyID)
	ctx := context.Background()

	// Prime the cache
	_, err := validator.ValidateToken(ctx, tokenString)
	require.NoError(b, err)

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateToken(ctx, tokenString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

package services

import (
	"testing"

	"fiber-app/internal/models"
	"fiber-app/pkg/cache"
	"fiber-app/pkg/config"

	"go.uber.org/zap"
)

func TestSessionService_CreateAndGetSession(t *testing.T) {
	// Setup test logger
	logger := zap.NewNop()

	// Setup test config and Redis connection
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       1, // Use test database
		},
	}

	// Connect to Redis (skip if Redis not available)
	if err := cache.Connect(cfg, logger); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// Clean up test data
	defer cache.FlushDB()

	// Create session service
	sessionService := NewSessionService(
		logger,
		"test-encryption-key-32-bytes-long",
		"test-csrf-secret-32-bytes-long-key",
	)

	// Test user info
	userInfo := &models.ZitadelUserInfo{
		Sub:           "test-user-123",
		Name:          "Test User",
		Email:         "test@example.com",
		EmailVerified: true,
		Roles:         []string{"user", "admin"},
		OrgID:         "test-org",
	}

	// Create session
	session, err := sessionService.CreateSession(userInfo, "refresh-token-123", "test-project")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session data
	if session.UserID != userInfo.Sub {
		t.Errorf("Expected UserID %s, got %s", userInfo.Sub, session.UserID)
	}

	if session.Name != userInfo.Name {
		t.Errorf("Expected Name %s, got %s", userInfo.Name, session.Name)
	}

	if session.Email != userInfo.Email {
		t.Errorf("Expected Email %s, got %s", userInfo.Email, session.Email)
	}

	if session.OrgID != "test-org" {
		t.Errorf("Expected OrgID test-org, got %s", session.OrgID)
	}

	if session.ProjectID != "test-project" {
		t.Errorf("Expected ProjectID test-project, got %s", session.ProjectID)
	}

	if session.RefreshTokenID != "refresh-token-123" {
		t.Errorf("Expected RefreshTokenID refresh-token-123, got %s", session.RefreshTokenID)
	}

	// Get session
	retrievedSession, err := sessionService.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	// Verify retrieved session
	if retrievedSession.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, retrievedSession.ID)
	}

	if retrievedSession.UserID != session.UserID {
		t.Errorf("Expected UserID %s, got %s", session.UserID, retrievedSession.UserID)
	}
}

func TestSessionService_CSRFToken(t *testing.T) {
	logger := zap.NewNop()
	sessionService := NewSessionService(
		logger,
		"test-encryption-key-32-bytes-long",
		"test-csrf-secret-32-bytes-long-key",
	)

	sessionID := "test-session-123"

	// Generate CSRF token
	token1 := sessionService.generateCSRFToken(sessionID)
	token2 := sessionService.generateCSRFToken(sessionID)

	// Tokens should be consistent for same session
	if token1 != token2 {
		t.Errorf("CSRF tokens should be consistent for same session")
	}

	// Validate CSRF token
	if !sessionService.ValidateCSRFToken(sessionID, token1) {
		t.Errorf("CSRF token validation failed")
	}

	// Invalid token should fail
	if sessionService.ValidateCSRFToken(sessionID, "invalid-token") {
		t.Errorf("Invalid CSRF token should not validate")
	}
}

func TestSessionService_Encryption(t *testing.T) {
	logger := zap.NewNop()
	sessionService := NewSessionService(
		logger,
		"test-encryption-key-32-bytes-long",
		"test-csrf-secret-32-bytes-long-key",
	)

	testData := []byte("sensitive session data")

	// Encrypt data
	encrypted, err := sessionService.encryptData(testData)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	// Decrypt data
	decrypted, err := sessionService.decryptData(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	// Verify data integrity
	if string(decrypted) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(decrypted))
	}
}

func TestSessionService_Fingerprinting(t *testing.T) {
	logger := zap.NewNop()
	sessionService := NewSessionService(
		logger,
		"test-encryption-key-32-bytes-long",
		"test-csrf-secret-32-bytes-long-key",
	)

	session := &Session{
		ID:        "test-session",
		UserID:    "test-user",
		RiskScore: 0,
	}

	fingerprint1 := "Mozilla/5.0|en-US|gzip"
	fingerprint2 := "Chrome/91.0|en-US|gzip"

	// First fingerprint should initialize
	action, err := sessionService.ValidateFingerprint(session, fingerprint1)
	if err != nil {
		t.Fatalf("Failed to validate fingerprint: %v", err)
	}

	if action.Action != "allow" || action.Reason != "fingerprint_initialized" {
		t.Errorf("Expected fingerprint initialization, got %s: %s", action.Action, action.Reason)
	}

	// Same fingerprint should allow
	action, err = sessionService.ValidateFingerprint(session, fingerprint1)
	if err != nil {
		t.Fatalf("Failed to validate fingerprint: %v", err)
	}

	if action.Action != "allow" || action.Reason != "fingerprint_match" {
		t.Errorf("Expected fingerprint match, got %s: %s", action.Action, action.Reason)
	}

	// Different fingerprint should increase risk
	action, err = sessionService.ValidateFingerprint(session, fingerprint2)
	if err != nil {
		t.Fatalf("Failed to validate fingerprint: %v", err)
	}

	if action.Action != "allow" || action.Reason != "fingerprint_mismatch_low_risk" {
		t.Errorf("Expected low risk mismatch, got %s: %s", action.Action, action.Reason)
	}

	if session.RiskScore != 10 {
		t.Errorf("Expected risk score 10, got %d", session.RiskScore)
	}
}

func TestSessionService_RefreshTokenReuse(t *testing.T) {
	logger := zap.NewNop()

	// Setup test config and Redis connection
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       1, // Use test database
		},
	}

	// Connect to Redis (skip if Redis not available)
	if err := cache.Connect(cfg, logger); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// Clean up test data
	defer cache.FlushDB()

	sessionService := NewSessionService(
		logger,
		"test-encryption-key-32-bytes-long",
		"test-csrf-secret-32-bytes-long-key",
	)

	refreshTokenID := "test-refresh-token-123"

	// First use should not be reused
	reused, err := sessionService.IsRefreshTokenReused(refreshTokenID)
	if err != nil {
		t.Fatalf("Failed to check token reuse: %v", err)
	}

	if reused {
		t.Errorf("Token should not be marked as reused on first use")
	}

	// Second use should be detected as reused
	reused, err = sessionService.IsRefreshTokenReused(refreshTokenID)
	if err != nil {
		t.Fatalf("Failed to check token reuse: %v", err)
	}

	if !reused {
		t.Errorf("Token should be marked as reused on second use")
	}
}

package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"fiber-app/pkg/cache"

	"go.uber.org/zap"
)

type PKCEService struct {
	logger *zap.Logger
}

type PKCEChallenge struct {
	CodeVerifier  string    `json:"code_verifier"`
	CodeChallenge string    `json:"code_challenge"`
	State         string    `json:"state"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewPKCEService(logger *zap.Logger) *PKCEService {
	return &PKCEService{
		logger: logger,
	}
}

func (s *PKCEService) GenerateChallenge() (*PKCEChallenge, error) {
	// Generate code verifier (43-128 characters)
	codeVerifier, err := s.generateRandomString(128)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	// Generate code challenge (SHA256 hash of verifier, base64url encoded)
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// Generate state parameter
	state, err := s.generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	challenge := &PKCEChallenge{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
		State:         state,
		CreatedAt:     time.Now(),
	}

	// Store in Redis for 10 minutes
	stateKey := fmt.Sprintf("pkce_state:%s", state)
	if err := cache.Set(stateKey, challenge, 10*time.Minute); err != nil {
		s.logger.Error("Failed to store PKCE challenge",
			zap.String("state", state),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to store PKCE challenge: %w", err)
	}

	s.logger.Info("PKCE challenge generated",
		zap.String("state", state),
		zap.String("code_challenge", codeChallenge),
	)

	return challenge, nil
}

func (s *PKCEService) ValidateAndGetChallenge(state string) (*PKCEChallenge, error) {
	stateKey := fmt.Sprintf("pkce_state:%s", state)

	var challenge PKCEChallenge
	if err := cache.Get(stateKey, &challenge); err != nil {
		s.logger.Warn("PKCE state not found or expired",
			zap.String("state", state),
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid or expired state: %w", err)
	}

	// Delete the state after use (one-time use)
	if err := cache.Delete(stateKey); err != nil {
		s.logger.Warn("Failed to delete used PKCE state",
			zap.String("state", state),
			zap.Error(err),
		)
	}

	// Check if not expired (10 minutes)
	if time.Since(challenge.CreatedAt) > 10*time.Minute {
		return nil, fmt.Errorf("PKCE challenge expired")
	}

	s.logger.Info("PKCE challenge validated",
		zap.String("state", state),
	)

	return &challenge, nil
}

func (s *PKCEService) generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}

package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"fiber-app/internal/models"
	"fiber-app/pkg/cache"

	"go.uber.org/zap"
)

type SessionService struct {
	logger        *zap.Logger
	encryptionKey []byte
	csrfSecret    []byte
}

type Session struct {
	ID             string            `json:"id"`
	UserID         string            `json:"user_id"`
	Name           string            `json:"name"`
	Email          string            `json:"email"`
	Roles          []string          `json:"roles"`
	OrgID          string            `json:"org_id"`
	ProjectID      string            `json:"project_id"`
	LoginTime      time.Time         `json:"login_time"`
	LastActivity   time.Time         `json:"last_activity"`
	CSRFToken      string            `json:"csrf_token"`
	Fingerprint    string            `json:"fingerprint"`
	RefreshTokenID string            `json:"refresh_token_id"` // For rotation tracking
	RefreshToken   string            `json:"refresh_token"`    // Encrypted refresh token storage
	RiskScore      int               `json:"risk_score"`       // For adaptive security
	Metadata       map[string]string `json:"metadata"`
}

type SecurityAction struct {
	Action    string `json:"action"` // "allow", "challenge", "deny"
	Reason    string `json:"reason"` // "fingerprint_mismatch", "suspicious_activity"
	RiskScore int    `json:"risk_score"`
}

func NewSessionService(logger *zap.Logger, encryptionKey, csrfSecret string) *SessionService {
	// Decode encryption key from hex or base64
	var keyBytes []byte
	var err error

	// Try hex first, then base64
	if keyBytes, err = hex.DecodeString(encryptionKey); err != nil {
		if keyBytes, err = base64.StdEncoding.DecodeString(encryptionKey); err != nil {
			// If both fail, use the string as-is (for development)
			keyBytes = []byte(encryptionKey)
		}
	}

	// Ensure key is 32 bytes for AES-256
	if len(keyBytes) != 32 {
		hash := sha256.Sum256(keyBytes)
		keyBytes = hash[:]
	}

	// Decode CSRF secret
	var csrfBytes []byte
	if csrfBytes, err = hex.DecodeString(csrfSecret); err != nil {
		if csrfBytes, err = base64.StdEncoding.DecodeString(csrfSecret); err != nil {
			csrfBytes = []byte(csrfSecret)
		}
	}

	// Ensure CSRF secret is 32 bytes
	if len(csrfBytes) != 32 {
		hash := sha256.Sum256(csrfBytes)
		csrfBytes = hash[:]
	}

	return &SessionService{
		logger:        logger,
		encryptionKey: keyBytes,
		csrfSecret:    csrfBytes,
	}
}

// encryptData encrypts data using AES-256-GCM
func (s *SessionService) encryptData(data []byte) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptData decrypts data using AES-256-GCM
func (s *SessionService) decryptData(encryptedData string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// generateCSRFToken generates a CSRF token for the session
func (s *SessionService) generateCSRFToken(sessionID string) string {
	// Create a hash of session ID + CSRF secret + timestamp
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	data := sessionID + string(s.csrfSecret) + timestamp
	hash := sha256.Sum256([]byte(data))
	return base64.URLEncoding.EncodeToString(hash[:16]) // Use first 16 bytes for shorter token
}

// generateFingerprint creates a fingerprint from request metadata
func (s *SessionService) generateFingerprint(userAgent, acceptLanguage, acceptEncoding string) string {
	data := fmt.Sprintf("%s|%s|%s", userAgent, acceptLanguage, acceptEncoding)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for shorter fingerprint
}

// generateSessionID generates a cryptographically secure session ID
func (s *SessionService) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (s *SessionService) CreateSession(userInfo *models.ZitadelUserInfo, refreshTokenID, projectID string) (*Session, error) {
	// Generate secure session ID
	sessionID, err := s.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Generate CSRF token
	csrfToken := s.generateCSRFToken(sessionID)

	// Create session key with org/project isolation
	// Note: OrgID is mapped from "urn:zitadel:iam:org:id" claim
	orgID := userInfo.OrgID
	if orgID == "" {
		orgID = "default"
	}

	// ProjectID is passed as parameter or use default
	if projectID == "" {
		projectID = "default"
	}

	sessionKey := fmt.Sprintf("session:%s:%s:%s", orgID, projectID, sessionID)

	session := &Session{
		ID:             sessionID,
		UserID:         userInfo.Sub,
		Name:           userInfo.Name,
		Email:          userInfo.Email,
		Roles:          userInfo.Roles,
		OrgID:          orgID,
		ProjectID:      projectID,
		LoginTime:      time.Now(),
		LastActivity:   time.Now(),
		CSRFToken:      csrfToken,
		RefreshTokenID: refreshTokenID,
		RiskScore:      0, // Start with low risk
		Metadata:       make(map[string]string),
	}

	// Encrypt session data before storing
	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	encryptedData, err := s.encryptData(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt session: %w", err)
	}

	// Store encrypted session in Redis for 24 hours
	if err := cache.Set(sessionKey, encryptedData, 24*time.Hour); err != nil {
		s.logger.Error("Failed to create session",
			zap.String("session_id", sessionID),
			zap.String("org_id", orgID),
			zap.String("project_id", projectID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info("Session created successfully",
		zap.String("session_id", sessionID),
		zap.String("user_id", userInfo.Sub),
		zap.String("org_id", orgID),
		zap.String("project_id", projectID),
		zap.String("email", s.maskEmail(userInfo.Email)),
	)

	return session, nil
}

func (s *SessionService) GetSession(sessionID string) (*Session, error) {
	// Try to find session with different org/project combinations
	// This is a fallback for when we don't know the org/project context
	pattern := fmt.Sprintf("session:*:*:%s", sessionID)
	keys, err := cache.Keys(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search for session: %w", err)
	}

	if len(keys) == 0 {
		s.logger.Debug("Session not found",
			zap.String("session_id", sessionID),
		)
		return nil, fmt.Errorf("session not found")
	}

	// Use the first matching key (there should only be one)
	sessionKey := keys[0]

	var encryptedData string
	if err := cache.Get(sessionKey, &encryptedData); err != nil {
		s.logger.Debug("Session not found in cache",
			zap.String("session_key", sessionKey),
			zap.Error(err),
		)
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Decrypt session data
	sessionData, err := s.decryptData(encryptedData)
	if err != nil {
		s.logger.Error("Failed to decrypt session",
			zap.String("session_key", sessionKey),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to decrypt session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(sessionData, &session); err != nil {
		s.logger.Error("Failed to unmarshal session",
			zap.String("session_key", sessionKey),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Update last activity
	session.LastActivity = time.Now()

	// Re-encrypt and store updated session
	updatedData, err := json.Marshal(&session)
	if err != nil {
		s.logger.Warn("Failed to marshal updated session", zap.Error(err))
	} else {
		encryptedUpdated, err := s.encryptData(updatedData)
		if err != nil {
			s.logger.Warn("Failed to encrypt updated session", zap.Error(err))
		} else {
			if err := cache.Set(sessionKey, encryptedUpdated, 24*time.Hour); err != nil {
				s.logger.Warn("Failed to update session activity",
					zap.String("session_key", sessionKey),
					zap.Error(err),
				)
			}
		}
	}

	return &session, nil
}

// GetSessionByContext gets session with known org/project context for better performance
func (s *SessionService) GetSessionByContext(sessionID, orgID, projectID string) (*Session, error) {
	sessionKey := fmt.Sprintf("session:%s:%s:%s", orgID, projectID, sessionID)

	var encryptedData string
	if err := cache.Get(sessionKey, &encryptedData); err != nil {
		s.logger.Debug("Session not found in cache",
			zap.String("session_key", sessionKey),
			zap.Error(err),
		)
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Decrypt session data
	sessionData, err := s.decryptData(encryptedData)
	if err != nil {
		s.logger.Error("Failed to decrypt session",
			zap.String("session_key", sessionKey),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to decrypt session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(sessionData, &session); err != nil {
		s.logger.Error("Failed to unmarshal session",
			zap.String("session_key", sessionKey),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Update last activity
	session.LastActivity = time.Now()

	// Re-encrypt and store updated session
	updatedData, err := json.Marshal(&session)
	if err != nil {
		s.logger.Warn("Failed to marshal updated session", zap.Error(err))
	} else {
		encryptedUpdated, err := s.encryptData(updatedData)
		if err != nil {
			s.logger.Warn("Failed to encrypt updated session", zap.Error(err))
		} else {
			if err := cache.Set(sessionKey, encryptedUpdated, 24*time.Hour); err != nil {
				s.logger.Warn("Failed to update session activity",
					zap.String("session_key", sessionKey),
					zap.Error(err),
				)
			}
		}
	}

	return &session, nil
}

func (s *SessionService) DeleteSession(sessionID string) error {
	// Find session key pattern
	pattern := fmt.Sprintf("session:*:*:%s", sessionID)
	keys, err := cache.Keys(pattern)
	if err != nil {
		return fmt.Errorf("failed to search for session: %w", err)
	}

	if len(keys) == 0 {
		s.logger.Debug("Session not found for deletion",
			zap.String("session_id", sessionID),
		)
		return nil // Not an error if session doesn't exist
	}

	// Delete all matching sessions (should be only one)
	for _, key := range keys {
		if err := cache.Delete(key); err != nil {
			s.logger.Error("Failed to delete session",
				zap.String("session_key", key),
				zap.Error(err),
			)
			return fmt.Errorf("failed to delete session: %w", err)
		}
	}

	s.logger.Info("Session deleted successfully",
		zap.String("session_id", sessionID),
		zap.Int("keys_deleted", len(keys)),
	)

	return nil
}

// DeleteSessionByContext deletes session with known context for better performance
func (s *SessionService) DeleteSessionByContext(sessionID, orgID, projectID string) error {
	sessionKey := fmt.Sprintf("session:%s:%s:%s", orgID, projectID, sessionID)

	if err := cache.Delete(sessionKey); err != nil {
		s.logger.Error("Failed to delete session",
			zap.String("session_key", sessionKey),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.logger.Info("Session deleted successfully",
		zap.String("session_id", sessionID),
		zap.String("org_id", orgID),
		zap.String("project_id", projectID),
	)

	return nil
}

// ValidateFingerprint validates session fingerprint and returns security action
func (s *SessionService) ValidateFingerprint(session *Session, currentFingerprint string) (*SecurityAction, error) {
	if session.Fingerprint == "" {
		// First time setting fingerprint
		session.Fingerprint = currentFingerprint
		return &SecurityAction{
			Action:    "allow",
			Reason:    "fingerprint_initialized",
			RiskScore: 0,
		}, nil
	}

	if session.Fingerprint == currentFingerprint {
		// Fingerprint matches
		return &SecurityAction{
			Action:    "allow",
			Reason:    "fingerprint_match",
			RiskScore: session.RiskScore,
		}, nil
	}

	// Fingerprint mismatch - increase risk score
	session.RiskScore += 10

	// Risk-based decision
	if session.RiskScore >= 50 {
		return &SecurityAction{
			Action:    "deny",
			Reason:    "fingerprint_mismatch_high_risk",
			RiskScore: session.RiskScore,
		}, nil
	} else if session.RiskScore >= 20 {
		return &SecurityAction{
			Action:    "challenge",
			Reason:    "fingerprint_mismatch_medium_risk",
			RiskScore: session.RiskScore,
		}, nil
	}

	// Low risk - allow but log
	return &SecurityAction{
		Action:    "allow",
		Reason:    "fingerprint_mismatch_low_risk",
		RiskScore: session.RiskScore,
	}, nil
}

// RotateSessionID creates a new session ID for session fixation protection
func (s *SessionService) RotateSessionID(oldSessionID string) (*Session, error) {
	// Get existing session
	session, err := s.GetSession(oldSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing session: %w", err)
	}

	// Generate new session ID
	newSessionID, err := s.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new session ID: %w", err)
	}

	// Update session with new ID and CSRF token
	session.ID = newSessionID
	session.CSRFToken = s.generateCSRFToken(newSessionID)
	session.LastActivity = time.Now()

	// Create new session key
	newSessionKey := fmt.Sprintf("session:%s:%s:%s", session.OrgID, session.ProjectID, newSessionID)

	// Encrypt and store new session
	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	encryptedData, err := s.encryptData(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt session: %w", err)
	}

	if err := cache.Set(newSessionKey, encryptedData, 24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store new session: %w", err)
	}

	// Delete old session
	if err := s.DeleteSession(oldSessionID); err != nil {
		s.logger.Warn("Failed to delete old session during rotation",
			zap.String("old_session_id", oldSessionID),
			zap.Error(err),
		)
	}

	s.logger.Info("Session ID rotated successfully",
		zap.String("old_session_id", oldSessionID),
		zap.String("new_session_id", newSessionID),
		zap.String("user_id", session.UserID),
	)

	return session, nil
}

// StoreRefreshToken encrypts and stores refresh token in session
func (s *SessionService) StoreRefreshToken(sessionID, refreshToken string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Encrypt refresh token
	encryptedToken, err := s.encryptData([]byte(refreshToken))
	if err != nil {
		return fmt.Errorf("failed to encrypt refresh token: %w", err)
	}

	session.RefreshToken = encryptedToken
	session.LastActivity = time.Now()

	// Re-encrypt and store session
	sessionKey := fmt.Sprintf("session:%s:%s:%s", session.OrgID, session.ProjectID, sessionID)
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	encryptedData, err := s.encryptData(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encrypt session: %w", err)
	}

	if err := cache.Set(sessionKey, encryptedData, 24*time.Hour); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves and decrypts refresh token from session
func (s *SessionService) GetRefreshToken(sessionID string) (string, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	if session.RefreshToken == "" {
		return "", fmt.Errorf("no refresh token in session")
	}

	// Decrypt refresh token
	tokenData, err := s.decryptData(session.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	return string(tokenData), nil
}

// maskEmail masks email for logging (PII protection)
func (s *SessionService) maskEmail(email string) string {
	if email == "" {
		return ""
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return "***@" + domain
	}

	return username[:1] + "***@" + domain
}

func (s *SessionService) ValidateSession(sessionID string) (*Session, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Check if session is expired (24 hours)
	if time.Since(session.LoginTime) > 24*time.Hour {
		s.DeleteSession(sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// ValidateCSRFToken validates CSRF token against session
func (s *SessionService) ValidateCSRFToken(sessionID, token string) bool {
	expectedToken := s.generateCSRFToken(sessionID)
	return expectedToken == token
}

// CleanupExpiredSessions removes expired sessions from Redis
func (s *SessionService) CleanupExpiredSessions() error {
	pattern := "session:*:*:*"
	keys, err := cache.Keys(pattern)
	if err != nil {
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	expiredCount := 0
	for _, key := range keys {
		var encryptedData string
		if err := cache.Get(key, &encryptedData); err != nil {
			continue // Skip if can't get data
		}

		sessionData, err := s.decryptData(encryptedData)
		if err != nil {
			// Delete corrupted sessions
			cache.Delete(key)
			expiredCount++
			continue
		}

		var session Session
		if err := json.Unmarshal(sessionData, &session); err != nil {
			// Delete corrupted sessions
			cache.Delete(key)
			expiredCount++
			continue
		}

		// Check if expired (24 hours)
		if time.Since(session.LoginTime) > 24*time.Hour {
			cache.Delete(key)
			expiredCount++
		}
	}

	s.logger.Info("Expired sessions cleaned up",
		zap.Int("expired_count", expiredCount),
		zap.Int("total_checked", len(keys)),
	)

	return nil
}

// HandleRefreshTokenReuse handles refresh token reuse detection and revokes all sessions
func (s *SessionService) HandleRefreshTokenReuse(refreshTokenID string) error {
	// Find all sessions with this refresh token ID
	pattern := "session:*:*:*"
	keys, err := cache.Keys(pattern)
	if err != nil {
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	revokedCount := 0
	for _, key := range keys {
		var encryptedData string
		if err := cache.Get(key, &encryptedData); err != nil {
			continue
		}

		sessionData, err := s.decryptData(encryptedData)
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(sessionData, &session); err != nil {
			continue
		}

		// Check if this session has the reused refresh token ID
		if session.RefreshTokenID == refreshTokenID {
			if err := cache.Delete(key); err != nil {
				s.logger.Error("Failed to revoke session during token reuse handling",
					zap.String("session_key", key),
					zap.Error(err),
				)
			} else {
				revokedCount++
			}
		}
	}

	s.logger.Warn("Refresh token reuse detected - all sessions revoked",
		zap.String("refresh_token_id", refreshTokenID),
		zap.Int("sessions_revoked", revokedCount),
	)

	return nil
}

// RotateRefreshToken updates session with new refresh token and ID
func (s *SessionService) RotateRefreshToken(sessionID, newRefreshToken, newRefreshTokenID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Encrypt new refresh token
	encryptedToken, err := s.encryptData([]byte(newRefreshToken))
	if err != nil {
		return fmt.Errorf("failed to encrypt refresh token: %w", err)
	}

	// Update session with new refresh token data
	session.RefreshToken = encryptedToken
	session.RefreshTokenID = newRefreshTokenID
	session.LastActivity = time.Now()

	// Store updated session
	sessionKey := fmt.Sprintf("session:%s:%s:%s", session.OrgID, session.ProjectID, sessionID)
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	encryptedData, err := s.encryptData(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encrypt session: %w", err)
	}

	if err := cache.Set(sessionKey, encryptedData, 24*time.Hour); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	s.logger.Info("Refresh token rotated successfully",
		zap.String("session_id", sessionID),
		zap.String("user_id", session.UserID),
		zap.String("new_refresh_token_id", newRefreshTokenID),
	)

	return nil
}

// IsRefreshTokenReused checks if a refresh token ID has been used before
func (s *SessionService) IsRefreshTokenReused(refreshTokenID string) (bool, error) {
	// Store used refresh token IDs in Redis with TTL
	key := fmt.Sprintf("used_refresh_token:%s", refreshTokenID)

	// Check if token ID already exists
	exists := cache.Exists(key)
	if exists {
		return true, nil
	}

	// Mark token as used (store for 48 hours to detect reuse)
	if err := cache.Set(key, "used", 48*time.Hour); err != nil {
		return false, fmt.Errorf("failed to mark refresh token as used: %w", err)
	}

	return false, nil
}

// GetSessionsByUser gets all sessions for a specific user (for session management)
func (s *SessionService) GetSessionsByUser(userID string) ([]*Session, error) {
	pattern := "session:*:*:*"
	keys, err := cache.Keys(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %w", err)
	}

	var sessions []*Session
	for _, key := range keys {
		var encryptedData string
		if err := cache.Get(key, &encryptedData); err != nil {
			continue
		}

		sessionData, err := s.decryptData(encryptedData)
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(sessionData, &session); err != nil {
			continue
		}

		if session.UserID == userID {
			sessions = append(sessions, &session)
		}
	}

	return sessions, nil
}

// RevokeAllUserSessions revokes all sessions for a specific user
func (s *SessionService) RevokeAllUserSessions(userID string) error {
	sessions, err := s.GetSessionsByUser(userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	revokedCount := 0
	for _, session := range sessions {
		if err := s.DeleteSessionByContext(session.ID, session.OrgID, session.ProjectID); err != nil {
			s.logger.Error("Failed to revoke user session",
				zap.String("session_id", session.ID),
				zap.String("user_id", userID),
				zap.Error(err),
			)
		} else {
			revokedCount++
		}
	}

	s.logger.Info("All user sessions revoked",
		zap.String("user_id", userID),
		zap.Int("sessions_revoked", revokedCount),
	)

	return nil
}

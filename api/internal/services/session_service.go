package services

import (
	"fmt"
	"time"

	"fiber-app/pkg/cache"

	"go.uber.org/zap"
)

type SessionService struct {
	logger *zap.Logger
}

type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Roles        []string  `json:"roles"`
	LoginTime    time.Time `json:"login_time"`
	LastActivity time.Time `json:"last_activity"`
}

func NewSessionService(logger *zap.Logger) *SessionService {
	return &SessionService{
		logger: logger,
	}
}

func (s *SessionService) CreateSession(userInfo *ZitadelUserInfo) (*Session, error) {
	sessionID := fmt.Sprintf("session:%s", userInfo.Sub)

	session := &Session{
		ID:           sessionID,
		UserID:       userInfo.Sub,
		Name:         userInfo.Name,
		Email:        userInfo.Email,
		Roles:        userInfo.Roles,
		LoginTime:    time.Now(),
		LastActivity: time.Now(),
	}

	// Store session in Redis for 24 hours
	if err := cache.Set(sessionID, session, 24*time.Hour); err != nil {
		s.logger.Error("Failed to create session",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info("Session created successfully",
		zap.String("session_id", sessionID),
		zap.String("user_id", userInfo.Sub),
		zap.String("email", userInfo.Email),
	)

	return session, nil
}

func (s *SessionService) GetSession(sessionID string) (*Session, error) {
	var session Session
	if err := cache.Get(sessionID, &session); err != nil {
		s.logger.Debug("Session not found",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Update last activity
	session.LastActivity = time.Now()
	if err := cache.Set(sessionID, &session, 24*time.Hour); err != nil {
		s.logger.Warn("Failed to update session activity",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
	}

	return &session, nil
}

func (s *SessionService) DeleteSession(sessionID string) error {
	if err := cache.Delete(sessionID); err != nil {
		s.logger.Error("Failed to delete session",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.logger.Info("Session deleted successfully",
		zap.String("session_id", sessionID),
	)

	return nil
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

package services

import (
	"fiber-app/internal/models"
	"fiber-app/pkg/cache"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	// Cache key prefixes
	UserCachePrefix   = "user:"
	UserRolePrefix    = "user_roles:"
	ZitadelUserPrefix = "zitadel_user:"

	// Cache TTL
	DefaultCacheTTL = 15 * time.Minute
	RoleCacheTTL    = 30 * time.Minute
)

type CacheService struct {
	logger *zap.Logger
}

func NewCacheService(logger *zap.Logger) *CacheService {
	return &CacheService{
		logger: logger,
	}
}

// User Cache Operations

// GetUser - Get user from cache by ID
func (cs *CacheService) GetUser(userID uint) (*models.User, error) {
	key := fmt.Sprintf("%s%d", UserCachePrefix, userID)

	var user models.User
	err := cache.Get(key, &user)
	if err != nil {
		cs.logger.Debug("User cache miss",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return nil, err
	}

	cs.logger.Debug("User cache hit",
		zap.Uint("user_id", userID),
	)

	return &user, nil
}

// GetUserByZitadelID - Get user from cache by Zitadel ID
func (cs *CacheService) GetUserByZitadelID(zitadelID string) (*models.User, error) {
	key := fmt.Sprintf("%s%s", ZitadelUserPrefix, zitadelID)

	var user models.User
	err := cache.Get(key, &user)
	if err != nil {
		cs.logger.Debug("Zitadel user cache miss",
			zap.String("zitadel_id", zitadelID),
			zap.Error(err),
		)
		return nil, err
	}

	cs.logger.Debug("Zitadel user cache hit",
		zap.String("zitadel_id", zitadelID),
	)

	return &user, nil
}

// SetUser - Save user to cache
func (cs *CacheService) SetUser(user *models.User) error {
	// Cache by ID
	key := fmt.Sprintf("%s%d", UserCachePrefix, user.ID)
	err := cache.Set(key, user, DefaultCacheTTL)
	if err != nil {
		cs.logger.Error("User cache set failed",
			zap.Uint("user_id", user.ID),
			zap.Error(err),
		)
		return err
	}

	// Cache by Zitadel ID
	zitadelKey := fmt.Sprintf("%s%s", ZitadelUserPrefix, user.ZitadelID)
	err = cache.Set(zitadelKey, user, DefaultCacheTTL)
	if err != nil {
		cs.logger.Error("Zitadel user cache set failed",
			zap.String("zitadel_id", user.ZitadelID),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("User cached",
		zap.Uint("user_id", user.ID),
		zap.String("zitadel_id", user.ZitadelID),
	)

	return nil
}

// DeleteUser - Delete user cache
func (cs *CacheService) DeleteUser(userID uint) error {
	key := fmt.Sprintf("%s%d", UserCachePrefix, userID)

	err := cache.Delete(key)
	if err != nil {
		cs.logger.Error("User cache delete failed",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("User cache deleted",
		zap.Uint("user_id", userID),
	)

	return nil
}

// DeleteUserByZitadelID - Delete user cache by Zitadel ID
func (cs *CacheService) DeleteUserByZitadelID(zitadelID string) error {
	key := fmt.Sprintf("%s%s", ZitadelUserPrefix, zitadelID)

	err := cache.Delete(key)
	if err != nil {
		cs.logger.Error("Zitadel user cache delete failed",
			zap.String("zitadel_id", zitadelID),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("Zitadel user cache deleted",
		zap.String("zitadel_id", zitadelID),
	)

	return nil
}

// User Role Cache Operations

// GetUserRoles - Get user roles from cache
func (cs *CacheService) GetUserRoles(userID uint) ([]models.UserRole, error) {
	key := fmt.Sprintf("%s%d", UserRolePrefix, userID)

	var roles []models.UserRole
	err := cache.Get(key, &roles)
	if err != nil {
		cs.logger.Debug("User roles cache miss",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return nil, err
	}

	cs.logger.Debug("User roles cache hit",
		zap.Uint("user_id", userID),
	)

	return roles, nil
}

// SetUserRoles - Save user roles to cache
func (cs *CacheService) SetUserRoles(userID uint, roles []models.UserRole) error {
	key := fmt.Sprintf("%s%d", UserRolePrefix, userID)

	err := cache.Set(key, roles, RoleCacheTTL)
	if err != nil {
		cs.logger.Error("User roles cache set failed",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("User roles cached",
		zap.Uint("user_id", userID),
		zap.Int("role_count", len(roles)),
	)

	return nil
}

// DeleteUserRoles - Delete user roles cache
func (cs *CacheService) DeleteUserRoles(userID uint) error {
	key := fmt.Sprintf("%s%d", UserRolePrefix, userID)

	err := cache.Delete(key)
	if err != nil {
		cs.logger.Error("User roles cache delete failed",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("User roles cache deleted",
		zap.Uint("user_id", userID),
	)

	return nil
}

// Session Cache Operations (for future use with Zitadel sessions)

// GetSession - Get session from cache
func (cs *CacheService) GetSession(sessionID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	var session map[string]interface{}
	err := cache.Get(key, &session)
	if err != nil {
		cs.logger.Debug("Session cache miss",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return nil, err
	}

	cs.logger.Debug("Session cache hit",
		zap.String("session_id", sessionID),
	)

	return session, nil
}

// SetSession - Save session to cache
func (cs *CacheService) SetSession(sessionID string, session map[string]interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)

	err := cache.Set(key, session, ttl)
	if err != nil {
		cs.logger.Error("Session cache set failed",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("Session cached",
		zap.String("session_id", sessionID),
	)

	return nil
}

// DeleteSession - Delete session cache
func (cs *CacheService) DeleteSession(sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)

	err := cache.Delete(key)
	if err != nil {
		cs.logger.Error("Session cache delete failed",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("Session cache deleted",
		zap.String("session_id", sessionID),
	)

	return nil
}

// Cache Management

// InvalidateUserCaches - Invalidate all user-related caches
func (cs *CacheService) InvalidateUserCaches(userID uint, zitadelID string) error {
	// Delete user cache by ID
	if err := cs.DeleteUser(userID); err != nil {
		cs.logger.Error("Failed to delete user cache", zap.Error(err))
	}

	// Delete user cache by Zitadel ID
	if zitadelID != "" {
		if err := cs.DeleteUserByZitadelID(zitadelID); err != nil {
			cs.logger.Error("Failed to delete zitadel user cache", zap.Error(err))
		}
	}

	// Delete user roles cache
	if err := cs.DeleteUserRoles(userID); err != nil {
		cs.logger.Error("Failed to delete user roles cache", zap.Error(err))
	}

	cs.logger.Info("User caches invalidated",
		zap.Uint("user_id", userID),
		zap.String("zitadel_id", zitadelID),
	)

	return nil
}

// InvalidateAllUserRoleCaches - Invalidate all user role caches (when roles change globally)
func (cs *CacheService) InvalidateAllUserRoleCaches() error {
	// Delete all user role caches
	pattern := fmt.Sprintf("%s*", UserRolePrefix)
	if err := cache.DeletePattern(pattern); err != nil {
		cs.logger.Error("Failed to delete user role caches", zap.Error(err))
		return err
	}

	cs.logger.Info("All user role caches invalidated")
	return nil
}

// GetCacheStats - Get cache statistics
func (cs *CacheService) GetCacheStats() (map[string]interface{}, error) {
	dbSize, err := cache.DBSize()
	if err != nil {
		return nil, err
	}

	userKeys, _ := cache.Keys(UserCachePrefix + "*")
	zitadelUserKeys, _ := cache.Keys(ZitadelUserPrefix + "*")
	userRoleKeys, _ := cache.Keys(UserRolePrefix + "*")
	sessionKeys, _ := cache.Keys("session:*")

	stats := map[string]interface{}{
		"total_keys":        dbSize,
		"user_keys":         len(userKeys),
		"zitadel_user_keys": len(zitadelUserKeys),
		"user_role_keys":    len(userRoleKeys),
		"session_keys":      len(sessionKeys),
	}

	return stats, nil
}

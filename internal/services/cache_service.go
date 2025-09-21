package services

import (
	"fiber-app/internal/models"
	"fiber-app/pkg/cache"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// Cache key prefixes
	UserCachePrefix = "user:"
	RoleCachePrefix = "role:"
	UserRolePrefix  = "user_role:"

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

// GetUser - Cache'den user getir
func (cs *CacheService) GetUser(userID uuid.UUID) (*models.User, error) {
	key := fmt.Sprintf("%s%s", UserCachePrefix, userID.String())

	var user models.User
	err := cache.Get(key, &user)
	if err != nil {
		cs.logger.Debug("User cache miss",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	cs.logger.Debug("User cache hit",
		zap.String("user_id", userID.String()),
	)

	return &user, nil
}

// SetUser - User'ı cache'e kaydet
func (cs *CacheService) SetUser(user *models.User) error {
	key := fmt.Sprintf("%s%s", UserCachePrefix, user.ID.String())

	err := cache.Set(key, user, DefaultCacheTTL)
	if err != nil {
		cs.logger.Error("User cache set failed",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("User cached",
		zap.String("user_id", user.ID.String()),
	)

	return nil
}

// DeleteUser - User cache'ini sil
func (cs *CacheService) DeleteUser(userID uuid.UUID) error {
	key := fmt.Sprintf("%s%s", UserCachePrefix, userID.String())

	err := cache.Delete(key)
	if err != nil {
		cs.logger.Error("User cache delete failed",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("User cache deleted",
		zap.String("user_id", userID.String()),
	)

	return nil
}

// Role Cache Operations

// GetRole - Cache'den role getir
func (cs *CacheService) GetRole(roleID uuid.UUID) (*models.Role, error) {
	key := fmt.Sprintf("%s%s", RoleCachePrefix, roleID.String())

	var role models.Role
	err := cache.Get(key, &role)
	if err != nil {
		cs.logger.Debug("Role cache miss",
			zap.String("role_id", roleID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	cs.logger.Debug("Role cache hit",
		zap.String("role_id", roleID.String()),
	)

	return &role, nil
}

// SetRole - Role'ü cache'e kaydet
func (cs *CacheService) SetRole(role *models.Role) error {
	key := fmt.Sprintf("%s%s", RoleCachePrefix, role.ID.String())

	err := cache.Set(key, role, RoleCacheTTL)
	if err != nil {
		cs.logger.Error("Role cache set failed",
			zap.String("role_id", role.ID.String()),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("Role cached",
		zap.String("role_id", role.ID.String()),
	)

	return nil
}

// DeleteRole - Role cache'ini sil
func (cs *CacheService) DeleteRole(roleID uuid.UUID) error {
	key := fmt.Sprintf("%s%s", RoleCachePrefix, roleID.String())

	err := cache.Delete(key)
	if err != nil {
		cs.logger.Error("Role cache delete failed",
			zap.String("role_id", roleID.String()),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("Role cache deleted",
		zap.String("role_id", roleID.String()),
	)

	return nil
}

// GetAllRoles - Tüm rolleri cache'den getir
func (cs *CacheService) GetAllRoles() ([]models.Role, error) {
	key := "all_roles"

	var roles []models.Role
	err := cache.Get(key, &roles)
	if err != nil {
		cs.logger.Debug("All roles cache miss", zap.Error(err))
		return nil, err
	}

	cs.logger.Debug("All roles cache hit")
	return roles, nil
}

// SetAllRoles - Tüm rolleri cache'e kaydet
func (cs *CacheService) SetAllRoles(roles []models.Role) error {
	key := "all_roles"

	err := cache.Set(key, roles, RoleCacheTTL)
	if err != nil {
		cs.logger.Error("All roles cache set failed", zap.Error(err))
		return err
	}

	cs.logger.Debug("All roles cached")
	return nil
}

// User-Role Relationship Cache

// GetUserRole - User'ın role bilgisini cache'den getir
func (cs *CacheService) GetUserRole(userID uuid.UUID) (*models.Role, error) {
	key := fmt.Sprintf("%s%s", UserRolePrefix, userID.String())

	var role models.Role
	err := cache.Get(key, &role)
	if err != nil {
		cs.logger.Debug("User role cache miss",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	cs.logger.Debug("User role cache hit",
		zap.String("user_id", userID.String()),
		zap.String("role", role.Name),
	)

	return &role, nil
}

// SetUserRole - User'ın role bilgisini cache'e kaydet
func (cs *CacheService) SetUserRole(userID uuid.UUID, role *models.Role) error {
	key := fmt.Sprintf("%s%s", UserRolePrefix, userID.String())

	err := cache.Set(key, role, DefaultCacheTTL)
	if err != nil {
		cs.logger.Error("User role cache set failed",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("User role cached",
		zap.String("user_id", userID.String()),
		zap.String("role", role.Name),
	)

	return nil
}

// DeleteUserRole - User'ın role cache'ini sil
func (cs *CacheService) DeleteUserRole(userID uuid.UUID) error {
	key := fmt.Sprintf("%s%s", UserRolePrefix, userID.String())

	err := cache.Delete(key)
	if err != nil {
		cs.logger.Error("User role cache delete failed",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("User role cache deleted",
		zap.String("user_id", userID.String()),
	)

	return nil
}

// Cache Management

// InvalidateUserCaches - User ile ilgili tüm cache'leri sil
func (cs *CacheService) InvalidateUserCaches(userID uuid.UUID) error {
	// User cache'ini sil
	if err := cs.DeleteUser(userID); err != nil {
		cs.logger.Error("Failed to delete user cache", zap.Error(err))
	}

	// User role cache'ini sil
	if err := cs.DeleteUserRole(userID); err != nil {
		cs.logger.Error("Failed to delete user role cache", zap.Error(err))
	}

	cs.logger.Info("User caches invalidated",
		zap.String("user_id", userID.String()),
	)

	return nil
}

// InvalidateRoleCaches - Role ile ilgili tüm cache'leri sil
func (cs *CacheService) InvalidateRoleCaches(roleID uuid.UUID) error {
	// Role cache'ini sil
	if err := cs.DeleteRole(roleID); err != nil {
		cs.logger.Error("Failed to delete role cache", zap.Error(err))
	}

	// All roles cache'ini sil
	if err := cache.Delete("all_roles"); err != nil {
		cs.logger.Error("Failed to delete all roles cache", zap.Error(err))
	}

	// Bu role'ü kullanan user'ların role cache'lerini sil
	pattern := fmt.Sprintf("%s*", UserRolePrefix)
	if err := cache.DeletePattern(pattern); err != nil {
		cs.logger.Error("Failed to delete user role caches", zap.Error(err))
	}

	cs.logger.Info("Role caches invalidated",
		zap.String("role_id", roleID.String()),
	)

	return nil
}

// GetCacheStats - Cache istatistikleri
func (cs *CacheService) GetCacheStats() (map[string]interface{}, error) {
	dbSize, err := cache.DBSize()
	if err != nil {
		return nil, err
	}

	userKeys, _ := cache.Keys(UserCachePrefix + "*")
	roleKeys, _ := cache.Keys(RoleCachePrefix + "*")
	userRoleKeys, _ := cache.Keys(UserRolePrefix + "*")

	stats := map[string]interface{}{
		"total_keys":     dbSize,
		"user_keys":      len(userKeys),
		"role_keys":      len(roleKeys),
		"user_role_keys": len(userRoleKeys),
	}

	return stats, nil
}

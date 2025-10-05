package repository

import (
	"fiber-app/internal/models"
	"fmt"

	"gorm.io/gorm"
)

// UserRepository interface for user data operations
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByZitadelID(zitadelID string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateUser(user *models.User) error
	UpdateUserRoles(userID uint, roles []models.UserRole) error
	GetUserRoles(userID uint) ([]models.UserRole, error)
	DeleteUser(id uint) error
	ListUsers(orgID, projectID string, limit, offset int) ([]models.User, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *userRepository) CreateUser(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByZitadelID retrieves a user by their Zitadel ID
func (r *userRepository) GetUserByZitadelID(zitadelID string) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Roles").Where("zitadel_id = ?", zitadelID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found, return nil without error
		}
		return nil, fmt.Errorf("failed to get user by zitadel_id: %w", err)
	}
	return &user, nil
}

// GetUserByID retrieves a user by their internal ID
func (r *userRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Roles").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User not found, return nil without error
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

// UpdateUser updates an existing user
func (r *userRepository) UpdateUser(user *models.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// UpdateUserRoles updates user roles (replaces existing roles)
func (r *userRepository) UpdateUserRoles(userID uint, roles []models.UserRole) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing roles for the user
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing roles: %w", err)
		}

		// Create new roles
		for i := range roles {
			roles[i].UserID = userID
			if err := tx.Create(&roles[i]).Error; err != nil {
				return fmt.Errorf("failed to create role: %w", err)
			}
		}

		return nil
	})
}

// GetUserRoles retrieves all roles for a user
func (r *userRepository) GetUserRoles(userID uint) ([]models.UserRole, error) {
	var roles []models.UserRole
	if err := r.db.Where("user_id = ?", userID).Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

// DeleteUser deletes a user and their associated roles (cascade)
func (r *userRepository) DeleteUser(id uint) error {
	if err := r.db.Delete(&models.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers retrieves users with pagination and filtering
func (r *userRepository) ListUsers(orgID, projectID string, limit, offset int) ([]models.User, error) {
	var users []models.User
	query := r.db.Preload("Roles")

	if orgID != "" {
		query = query.Where("org_id = ?", orgID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

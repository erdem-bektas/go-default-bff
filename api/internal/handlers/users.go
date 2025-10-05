package handlers

import (
	"errors"
	"fiber-app/internal/models"
	"fiber-app/pkg/database"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetUsers - List all users with pagination and filtering
// @Summary List users
// @Description List users with pagination and search support
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Records per page" default(10)
// @Param search query string false "Search term (name or email)"
// @Param org_id query string false "Organization ID filter"
// @Param project_id query string false "Project ID filter"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users [get]
func GetUsers(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	// Query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")
	orgID := c.Query("org_id", "")
	projectID := c.Query("project_id", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	zapLogger.Info("Users list requested",
		zap.String("trace_id", traceID),
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("search", search),
		zap.String("org_id", orgID),
		zap.String("project_id", projectID),
	)

	var users []models.User
	var total int64

	query := database.DB.Model(&models.User{}).Preload("Roles")

	// Search filter
	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Organization filter
	if orgID != "" {
		query = query.Where("org_id = ?", orgID)
	}

	// Project filter
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	// Total count
	if err := query.Count(&total).Error; err != nil {
		zapLogger.Error("Users count error",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	// Fetch data with pagination
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		zapLogger.Error("Users list error",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
		"trace_id": traceID,
	})
}

// GetUser - Get single user by ID
// @Summary Get user details
// @Description Get user details by ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (integer)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/{id} [get]
func GetUser(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	userIDStr := c.Params("id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "User ID required",
			"trace_id": traceID,
		})
	}

	// Convert to uint
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid User ID format",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("User details requested",
		zap.String("trace_id", traceID),
		zap.Uint64("user_id", userID),
	)

	var user models.User
	if err := database.DB.Preload("Roles").First(&user, uint(userID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "User not found",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("User fetch error",
			zap.String("trace_id", traceID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	return c.JSON(fiber.Map{
		"user":     user,
		"trace_id": traceID,
	})
}

// CreateUser - Create new user for Zitadel integration
// @Summary Create new user
// @Description Create new user record for Zitadel integration
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User information"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users [post]
func CreateUser(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		zapLogger.Error("User create body parse error",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid JSON format",
			"trace_id": traceID,
		})
	}

	// Basic validation
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Name field is required",
			"trace_id": traceID,
		})
	}

	if req.ZitadelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Zitadel ID field is required",
			"trace_id": traceID,
		})
	}

	if req.OrgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Organization ID field is required",
			"trace_id": traceID,
		})
	}

	if req.ProjectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Project ID field is required",
			"trace_id": traceID,
		})
	}

	emailStr := ""
	if req.Email != nil {
		emailStr = *req.Email
	}

	zapLogger.Info("Creating new user",
		zap.String("trace_id", traceID),
		zap.String("name", req.Name),
		zap.String("email", emailStr),
		zap.String("zitadel_id", req.ZitadelID),
		zap.String("org_id", req.OrgID),
		zap.String("project_id", req.ProjectID),
	)

	user := models.User{
		ZitadelID:     req.ZitadelID,
		Email:         req.Email,
		EmailVerified: req.EmailVerified,
		Name:          req.Name,
		GivenName:     req.GivenName,
		FamilyName:    req.FamilyName,
		Username:      req.Username,
		OrgID:         req.OrgID,
		ProjectID:     req.ProjectID,
		IsActive:      true,
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := database.DB.Create(&user).Error; err != nil {
		zapLogger.Error("User creation error",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)

		// Email unique constraint error
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "email") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":    "This email address is already in use",
				"trace_id": traceID,
			})
		}

		// Zitadel ID unique constraint error
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "zitadel_id") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":    "This Zitadel ID is already in use",
				"trace_id": traceID,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	// Load roles information
	database.DB.Preload("Roles").First(&user, user.ID)

	zapLogger.Info("User created successfully",
		zap.String("trace_id", traceID),
		zap.Uint("user_id", user.ID),
	)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "User created successfully",
		"user":     user,
		"trace_id": traceID,
	})
}

// UpdateUser - Update user information
// @Summary Update user
// @Description Update existing user information
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (integer)"
// @Param user body models.UpdateUserRequest true "User information to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/{id} [put]
func UpdateUser(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	userIDStr := c.Params("id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "User ID required",
			"trace_id": traceID,
		})
	}

	// Convert to uint
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid User ID format",
			"trace_id": traceID,
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		zapLogger.Error("User update body parse error",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid JSON format",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Updating user",
		zap.String("trace_id", traceID),
		zap.Uint64("user_id", userID),
	)

	// Check if user exists
	var user models.User
	if err := database.DB.Preload("Roles").First(&user, uint(userID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "User not found",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("User lookup error",
			zap.String("trace_id", traceID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	// Prepare update data
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.EmailVerified != nil {
		updates["email_verified"] = *req.EmailVerified
	}
	if req.GivenName != nil {
		updates["given_name"] = *req.GivenName
	}
	if req.FamilyName != nil {
		updates["family_name"] = *req.FamilyName
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "No fields to update",
			"trace_id": traceID,
		})
	}

	// Update
	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		zapLogger.Error("User update error",
			zap.String("trace_id", traceID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
		)

		// Email unique constraint error
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "email") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":    "This email address is already in use",
				"trace_id": traceID,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	// Get updated user
	if err := database.DB.Preload("Roles").First(&user, uint(userID)).Error; err != nil {
		zapLogger.Error("Updated user fetch error",
			zap.String("trace_id", traceID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("User updated successfully",
		zap.String("trace_id", traceID),
		zap.Uint64("user_id", userID),
	)

	return c.JSON(fiber.Map{
		"message":  "User updated successfully",
		"user":     user,
		"trace_id": traceID,
	})
}

// DeleteUser - Delete user
// @Summary Delete user
// @Description Delete user from system
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (integer)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/{id} [delete]
func DeleteUser(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	userIDStr := c.Params("id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "User ID required",
			"trace_id": traceID,
		})
	}

	// Convert to uint
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid User ID format",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Deleting user",
		zap.String("trace_id", traceID),
		zap.Uint64("user_id", userID),
	)

	// Check if user exists
	var user models.User
	if err := database.DB.Preload("Roles").First(&user, uint(userID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "User not found",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("User lookup error",
			zap.String("trace_id", traceID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	// Delete (roles will be cascade deleted due to foreign key constraint)
	if err := database.DB.Delete(&user).Error; err != nil {
		zapLogger.Error("User deletion error",
			zap.String("trace_id", traceID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("User deleted successfully",
		zap.String("trace_id", traceID),
		zap.Uint64("user_id", userID),
	)

	return c.JSON(fiber.Map{
		"message":  "User deleted successfully",
		"trace_id": traceID,
	})
}

// GetUserByZitadelID - Get user by Zitadel ID
// @Summary Get user by Zitadel ID
// @Description Get user details by Zitadel ID
// @Tags Users
// @Accept json
// @Produce json
// @Param zitadel_id path string true "Zitadel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/zitadel/{zitadel_id} [get]
func GetUserByZitadelID(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zitadelID := c.Params("zitadel_id")
	if zitadelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Zitadel ID required",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("User details requested by Zitadel ID",
		zap.String("trace_id", traceID),
		zap.String("zitadel_id", zitadelID),
	)

	var user models.User
	if err := database.DB.Preload("Roles").Where("zitadel_id = ?", zitadelID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "User not found",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("User fetch error",
			zap.String("trace_id", traceID),
			zap.String("zitadel_id", zitadelID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	return c.JSON(fiber.Map{
		"user":     user,
		"trace_id": traceID,
	})
}

// AssignUserRole - Assign role to user
// @Summary Assign role to user
// @Description Assign a role to a user for specific org/project
// @Tags Users
// @Accept json
// @Produce json
// @Param user_role body models.UserRoleRequest true "User role assignment"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/roles [post]
func AssignUserRole(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	var req models.UserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		zapLogger.Error("User role assignment body parse error",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid JSON format",
			"trace_id": traceID,
		})
	}

	// Validation
	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "User ID is required",
			"trace_id": traceID,
		})
	}

	if req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Role is required",
			"trace_id": traceID,
		})
	}

	if req.OrgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Organization ID is required",
			"trace_id": traceID,
		})
	}

	if req.ProjectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Project ID is required",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Assigning role to user",
		zap.String("trace_id", traceID),
		zap.Uint("user_id", req.UserID),
		zap.String("role", req.Role),
		zap.String("org_id", req.OrgID),
		zap.String("project_id", req.ProjectID),
	)

	// Check if user exists
	var user models.User
	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":    "User not found",
				"trace_id": traceID,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	userRole := models.UserRole{
		UserID:    req.UserID,
		Role:      req.Role,
		OrgID:     req.OrgID,
		ProjectID: req.ProjectID,
	}

	if err := database.DB.Create(&userRole).Error; err != nil {
		zapLogger.Error("User role assignment error",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)

		// Unique constraint error
		if strings.Contains(err.Error(), "duplicate key") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":    "User already has this role for this org/project",
				"trace_id": traceID,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Role assigned successfully",
		zap.String("trace_id", traceID),
		zap.Uint("user_role_id", userRole.ID),
	)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Role assigned successfully",
		"user_role": userRole,
		"trace_id":  traceID,
	})
}

// RemoveUserRole - Remove role from user
// @Summary Remove role from user
// @Description Remove a role from a user for specific org/project
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User Role ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/roles/{id} [delete]
func RemoveUserRole(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	userRoleIDStr := c.Params("id")
	if userRoleIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "User Role ID required",
			"trace_id": traceID,
		})
	}

	// Convert to uint
	userRoleID, err := strconv.ParseUint(userRoleIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid User Role ID format",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Removing user role",
		zap.String("trace_id", traceID),
		zap.Uint64("user_role_id", userRoleID),
	)

	// Check if user role exists
	var userRole models.UserRole
	if err := database.DB.First(&userRole, uint(userRoleID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "User role not found",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("User role lookup error",
			zap.String("trace_id", traceID),
			zap.Uint64("user_role_id", userRoleID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	// Delete
	if err := database.DB.Delete(&userRole).Error; err != nil {
		zapLogger.Error("User role deletion error",
			zap.String("trace_id", traceID),
			zap.Uint64("user_role_id", userRoleID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database error",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("User role removed successfully",
		zap.String("trace_id", traceID),
		zap.Uint64("user_role_id", userRoleID),
	)

	return c.JSON(fiber.Map{
		"message":  "User role removed successfully",
		"trace_id": traceID,
	})
}

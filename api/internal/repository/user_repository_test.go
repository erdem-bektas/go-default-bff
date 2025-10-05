package repository

import (
	"fiber-app/internal/models"
	"fiber-app/pkg/config"
	"fiber-app/pkg/database"
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDB   *gorm.DB
	testRepo UserRepository
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) {
	// Use existing database configuration but with test table prefix
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			DBName:   getEnvOrDefault("DB_NAME", "fiber_app"), // Use existing database
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		},
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	var err error
	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping test - failed to connect to database: %v", err)
	}

	// Configure connection pool for tests
	if err := database.ConfigureConnectionPool(testDB); err != nil {
		t.Fatalf("Failed to configure connection pool: %v", err)
	}

	// Run migrations for test - use existing tables
	if err := testDB.AutoMigrate(&models.User{}, &models.UserRole{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	testRepo = NewUserRepository(testDB)
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T) {
	if testDB != nil {
		// Clean up test data - delete all test records
		testDB.Exec("DELETE FROM user_roles WHERE user_id IN (SELECT id FROM users WHERE zitadel_id LIKE 'zitadel_%test%')")
		testDB.Exec("DELETE FROM users WHERE zitadel_id LIKE 'zitadel_%test%'")

		// Close connection
		sqlDB, err := testDB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Setup
	code := m.Run()

	// Teardown would happen here if needed
	os.Exit(code)
}

// TestUserRepository_CreateUser tests user creation functionality
func TestUserRepository_CreateUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "create user with email",
			user: &models.User{
				ZitadelID:     "zitadel_123",
				Email:         stringPtr("test@example.com"),
				EmailVerified: true,
				Name:          "Test User",
				GivenName:     "Test",
				FamilyName:    "User",
				Username:      "testuser",
				OrgID:         "org_123",
				ProjectID:     "project_123",
				IsActive:      true,
			},
			wantErr: false,
		},
		{
			name: "create user without email (social login)",
			user: &models.User{
				ZitadelID:     "zitadel_456",
				Email:         nil,
				EmailVerified: false,
				Name:          "Social User",
				GivenName:     "Social",
				FamilyName:    "User",
				Username:      "socialuser",
				OrgID:         "org_123",
				ProjectID:     "project_123",
				IsActive:      true,
			},
			wantErr: false,
		},
		{
			name: "create user with duplicate zitadel_id",
			user: &models.User{
				ZitadelID:     "zitadel_123", // Duplicate
				Email:         stringPtr("another@example.com"),
				EmailVerified: true,
				Name:          "Another User",
				GivenName:     "Another",
				FamilyName:    "User",
				Username:      "anotheruser",
				OrgID:         "org_123",
				ProjectID:     "project_123",
				IsActive:      true,
			},
			wantErr: true, // Should fail due to unique constraint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRepo.CreateUser(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify user was created
				if tt.user.ID == 0 {
					t.Error("Expected user ID to be set after creation")
				}
				if tt.user.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set after creation")
				}
			}
		})
	}
}

// TestUserRepository_GetUserByZitadelID tests user retrieval by Zitadel ID
func TestUserRepository_GetUserByZitadelID(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user
	testUser := &models.User{
		ZitadelID:     "zitadel_get_test",
		Email:         stringPtr("get@example.com"),
		EmailVerified: true,
		Name:          "Get Test User",
		GivenName:     "Get",
		FamilyName:    "User",
		Username:      "getuser",
		OrgID:         "org_123",
		ProjectID:     "project_123",
		IsActive:      true,
	}

	err := testRepo.CreateUser(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name      string
		zitadelID string
		wantUser  bool
		wantErr   bool
	}{
		{
			name:      "get existing user",
			zitadelID: "zitadel_get_test",
			wantUser:  true,
			wantErr:   false,
		},
		{
			name:      "get non-existent user",
			zitadelID: "zitadel_nonexistent",
			wantUser:  false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := testRepo.GetUserByZitadelID(tt.zitadelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByZitadelID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantUser && user == nil {
				t.Error("Expected user to be found, got nil")
			}
			if !tt.wantUser && user != nil {
				t.Error("Expected user to be nil, got user")
			}

			if user != nil {
				if user.ZitadelID != tt.zitadelID {
					t.Errorf("Expected ZitadelID %s, got %s", tt.zitadelID, user.ZitadelID)
				}
			}
		})
	}
}

// TestUserRepository_UpdateUser tests user update functionality
func TestUserRepository_UpdateUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user
	testUser := &models.User{
		ZitadelID:     "zitadel_update_test",
		Email:         stringPtr("update@example.com"),
		EmailVerified: false,
		Name:          "Update Test User",
		GivenName:     "Update",
		FamilyName:    "User",
		Username:      "updateuser",
		OrgID:         "org_123",
		ProjectID:     "project_123",
		IsActive:      true,
	}

	err := testRepo.CreateUser(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Update user
	testUser.EmailVerified = true
	testUser.Name = "Updated User Name"
	newEmail := "updated@example.com"
	testUser.Email = &newEmail

	err = testRepo.UpdateUser(testUser)
	if err != nil {
		t.Errorf("UpdateUser() error = %v", err)
	}

	// Verify update
	updatedUser, err := testRepo.GetUserByZitadelID(testUser.ZitadelID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if !updatedUser.EmailVerified {
		t.Error("Expected EmailVerified to be true after update")
	}
	if updatedUser.Name != "Updated User Name" {
		t.Errorf("Expected Name to be 'Updated User Name', got %s", updatedUser.Name)
	}
	if updatedUser.Email == nil || *updatedUser.Email != "updated@example.com" {
		t.Errorf("Expected Email to be 'updated@example.com', got %v", updatedUser.Email)
	}
}

// TestUserRepository_RoleAssignment tests user role assignment functionality
func TestUserRepository_RoleAssignment(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user
	testUser := &models.User{
		ZitadelID:     "zitadel_role_test",
		Email:         stringPtr("role@example.com"),
		EmailVerified: true,
		Name:          "Role Test User",
		GivenName:     "Role",
		FamilyName:    "User",
		Username:      "roleuser",
		OrgID:         "org_123",
		ProjectID:     "project_123",
		IsActive:      true,
	}

	err := testRepo.CreateUser(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test role assignment
	roles := []models.UserRole{
		{
			Role:      "admin",
			OrgID:     "org_123",
			ProjectID: "project_123",
		},
		{
			Role:      "user",
			OrgID:     "org_123",
			ProjectID: "project_123",
		},
	}

	err = testRepo.UpdateUserRoles(testUser.ID, roles)
	if err != nil {
		t.Errorf("UpdateUserRoles() error = %v", err)
	}

	// Verify roles were assigned
	assignedRoles, err := testRepo.GetUserRoles(testUser.ID)
	if err != nil {
		t.Errorf("GetUserRoles() error = %v", err)
	}

	if len(assignedRoles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(assignedRoles))
	}

	// Check role names
	roleNames := make(map[string]bool)
	for _, role := range assignedRoles {
		roleNames[role.Role] = true
		if role.UserID != testUser.ID {
			t.Errorf("Expected UserID %d, got %d", testUser.ID, role.UserID)
		}
		if role.OrgID != "org_123" {
			t.Errorf("Expected OrgID 'org_123', got %s", role.OrgID)
		}
		if role.ProjectID != "project_123" {
			t.Errorf("Expected ProjectID 'project_123', got %s", role.ProjectID)
		}
	}

	if !roleNames["admin"] || !roleNames["user"] {
		t.Error("Expected both 'admin' and 'user' roles to be assigned")
	}

	// Test role update (replace existing roles)
	newRoles := []models.UserRole{
		{
			Role:      "moderator",
			OrgID:     "org_123",
			ProjectID: "project_123",
		},
	}

	err = testRepo.UpdateUserRoles(testUser.ID, newRoles)
	if err != nil {
		t.Errorf("UpdateUserRoles() (replacement) error = %v", err)
	}

	// Verify old roles were replaced
	updatedRoles, err := testRepo.GetUserRoles(testUser.ID)
	if err != nil {
		t.Errorf("GetUserRoles() (after replacement) error = %v", err)
	}

	if len(updatedRoles) != 1 {
		t.Errorf("Expected 1 role after replacement, got %d", len(updatedRoles))
	}

	if updatedRoles[0].Role != "moderator" {
		t.Errorf("Expected role 'moderator', got %s", updatedRoles[0].Role)
	}
}

// TestUserRepository_DeleteUser tests user deletion with cascade
func TestUserRepository_DeleteUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user with roles
	testUser := &models.User{
		ZitadelID:     "zitadel_delete_test",
		Email:         stringPtr("delete@example.com"),
		EmailVerified: true,
		Name:          "Delete Test User",
		GivenName:     "Delete",
		FamilyName:    "User",
		Username:      "deleteuser",
		OrgID:         "org_123",
		ProjectID:     "project_123",
		IsActive:      true,
	}

	err := testRepo.CreateUser(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Assign roles
	roles := []models.UserRole{
		{
			Role:      "admin",
			OrgID:     "org_123",
			ProjectID: "project_123",
		},
	}

	err = testRepo.UpdateUserRoles(testUser.ID, roles)
	if err != nil {
		t.Fatalf("Failed to assign roles: %v", err)
	}

	// Delete user
	err = testRepo.DeleteUser(testUser.ID)
	if err != nil {
		t.Errorf("DeleteUser() error = %v", err)
	}

	// Verify user was deleted
	deletedUser, err := testRepo.GetUserByID(testUser.ID)
	if err != nil {
		t.Errorf("GetUserByID() after deletion error = %v", err)
	}
	if deletedUser != nil {
		t.Error("Expected user to be deleted, but still found")
	}

	// Verify roles were also deleted (cascade)
	deletedRoles, err := testRepo.GetUserRoles(testUser.ID)
	if err != nil {
		t.Errorf("GetUserRoles() after user deletion error = %v", err)
	}
	if len(deletedRoles) != 0 {
		t.Errorf("Expected 0 roles after user deletion, got %d", len(deletedRoles))
	}
}

// TestUserRepository_ListUsers tests user listing with pagination and filtering
func TestUserRepository_ListUsers(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test users in different orgs/projects
	users := []*models.User{
		{
			ZitadelID: "zitadel_list_1",
			Email:     stringPtr("list1@example.com"),
			Name:      "List User 1",
			OrgID:     "org_123",
			ProjectID: "project_123",
			IsActive:  true,
		},
		{
			ZitadelID: "zitadel_list_2",
			Email:     stringPtr("list2@example.com"),
			Name:      "List User 2",
			OrgID:     "org_123",
			ProjectID: "project_456",
			IsActive:  true,
		},
		{
			ZitadelID: "zitadel_list_3",
			Email:     stringPtr("list3@example.com"),
			Name:      "List User 3",
			OrgID:     "org_456",
			ProjectID: "project_123",
			IsActive:  true,
		},
	}

	for _, user := range users {
		err := testRepo.CreateUser(user)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
	}

	tests := []struct {
		name        string
		orgID       string
		projectID   string
		limit       int
		offset      int
		expectedLen int
	}{
		{
			name:        "list all users",
			orgID:       "",
			projectID:   "",
			limit:       10,
			offset:      0,
			expectedLen: 3,
		},
		{
			name:        "filter by org_id",
			orgID:       "org_123",
			projectID:   "",
			limit:       10,
			offset:      0,
			expectedLen: 2,
		},
		{
			name:        "filter by project_id",
			orgID:       "",
			projectID:   "project_123",
			limit:       10,
			offset:      0,
			expectedLen: 2,
		},
		{
			name:        "filter by org and project",
			orgID:       "org_123",
			projectID:   "project_123",
			limit:       10,
			offset:      0,
			expectedLen: 1,
		},
		{
			name:        "pagination - limit 2",
			orgID:       "",
			projectID:   "",
			limit:       2,
			offset:      0,
			expectedLen: 2,
		},
		{
			name:        "pagination - offset 2",
			orgID:       "",
			projectID:   "",
			limit:       10,
			offset:      2,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testRepo.ListUsers(tt.orgID, tt.projectID, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("ListUsers() error = %v", err)
				return
			}

			if len(result) != tt.expectedLen {
				t.Errorf("Expected %d users, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

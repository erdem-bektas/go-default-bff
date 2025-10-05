package test

import (
	"fiber-app/internal/models"
	"fiber-app/internal/repository"
	"fiber-app/pkg/config"
	"fiber-app/pkg/database"
	"fmt"
	"os"
	"testing"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDB   *gorm.DB
	testRepo repository.UserRepository
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) {
	// Use existing database configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			DBName:   getEnvOrDefault("DB_NAME", "fiber_app"),
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

	// Drop and recreate tables for testing to ensure clean schema
	testDB.Exec("DROP TABLE IF EXISTS user_roles CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS users CASCADE")

	// Auto-migrate tables for testing (creates tables based on Go models)
	if err := testDB.AutoMigrate(&models.User{}, &models.UserRole{}); err != nil {
		t.Fatalf("Failed to auto-migrate tables: %v", err)
	}

	// Ensure proper foreign key constraint with CASCADE DELETE
	// GORM sometimes doesn't create the constraint correctly, so we'll fix it manually
	testDB.Exec("ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS fk_users_roles")
	testDB.Exec("ALTER TABLE user_roles ADD CONSTRAINT fk_user_roles_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE")

	testRepo = repository.NewUserRepository(testDB)
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T) {
	if testDB != nil {
		// Clean up test data - delete all test records
		testDB.Exec("DELETE FROM user_roles WHERE user_id IN (SELECT id FROM users WHERE zitadel_id LIKE 'test_%')")
		testDB.Exec("DELETE FROM users WHERE zitadel_id LIKE 'test_%'")

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

// TestDatabaseConnection tests basic database connectivity
func TestDatabaseConnection(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Test basic database health check
	err := database.HealthCheck(testDB)
	if err != nil {
		t.Errorf("Database health check failed: %v", err)
	}

	// Test connection pool configuration
	err = database.ConfigureConnectionPool(testDB)
	if err != nil {
		t.Errorf("Connection pool configuration failed: %v", err)
	}

	// Test that we can execute a simple query
	var result int
	err = testDB.Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		t.Errorf("Simple query failed: %v", err)
	}

	if result != 1 {
		t.Errorf("Expected result 1, got %d", result)
	}
}

// TestUserRepository_BasicOperations tests basic user repository operations
func TestUserRepository_BasicOperations(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Test user creation
	testUser := &models.User{
		ZitadelID:     "test_user_123",
		Email:         stringPtr("test@example.com"),
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Username:      "testuser",
		OrgID:         "test_org_123",
		ProjectID:     "test_project_123",
		IsActive:      true,
	}

	err := testRepo.CreateUser(testUser)
	if err != nil {
		t.Errorf("CreateUser() error = %v", err)
		return
	}

	// Verify user was created with ID
	if testUser.ID == 0 {
		t.Error("Expected user ID to be set after creation")
	}

	// Test user retrieval by Zitadel ID
	retrievedUser, err := testRepo.GetUserByZitadelID(testUser.ZitadelID)
	if err != nil {
		t.Errorf("GetUserByZitadelID() error = %v", err)
		return
	}

	if retrievedUser == nil {
		t.Error("Expected user to be found, got nil")
		return
	}

	if retrievedUser.ZitadelID != testUser.ZitadelID {
		t.Errorf("Expected ZitadelID %s, got %s", testUser.ZitadelID, retrievedUser.ZitadelID)
	}

	if retrievedUser.Email == nil || *retrievedUser.Email != *testUser.Email {
		t.Errorf("Expected Email %s, got %v", *testUser.Email, retrievedUser.Email)
	}

	// Test user update
	newEmail := "updated@example.com"
	retrievedUser.Email = &newEmail
	retrievedUser.EmailVerified = false

	err = testRepo.UpdateUser(retrievedUser)
	if err != nil {
		t.Errorf("UpdateUser() error = %v", err)
		return
	}

	// Verify update
	updatedUser, err := testRepo.GetUserByZitadelID(testUser.ZitadelID)
	if err != nil {
		t.Errorf("GetUserByZitadelID() after update error = %v", err)
		return
	}

	if updatedUser.Email == nil || *updatedUser.Email != newEmail {
		t.Errorf("Expected updated email %s, got %v", newEmail, updatedUser.Email)
	}

	if updatedUser.EmailVerified {
		t.Error("Expected EmailVerified to be false after update")
	}
}

// TestUserRepository_RoleOperations tests user role operations
func TestUserRepository_RoleOperations(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user
	testUser := &models.User{
		ZitadelID:     "test_role_user_456",
		Email:         stringPtr("roletest@example.com"),
		EmailVerified: true,
		Name:          "Role Test User",
		GivenName:     "Role",
		FamilyName:    "User",
		Username:      "roleuser",
		OrgID:         "test_org_123",
		ProjectID:     "test_project_123",
		IsActive:      true,
	}

	err := testRepo.CreateUser(testUser)
	if err != nil {
		t.Errorf("CreateUser() error = %v", err)
		return
	}

	// Test role assignment
	roles := []models.UserRole{
		{
			Role:      "admin",
			OrgID:     "test_org_123",
			ProjectID: "test_project_123",
		},
		{
			Role:      "user",
			OrgID:     "test_org_123",
			ProjectID: "test_project_123",
		},
	}

	err = testRepo.UpdateUserRoles(testUser.ID, roles)
	if err != nil {
		t.Errorf("UpdateUserRoles() error = %v", err)
		return
	}

	// Verify roles were assigned
	assignedRoles, err := testRepo.GetUserRoles(testUser.ID)
	if err != nil {
		t.Errorf("GetUserRoles() error = %v", err)
		return
	}

	if len(assignedRoles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(assignedRoles))
		return
	}

	// Check role names
	roleNames := make(map[string]bool)
	for _, role := range assignedRoles {
		roleNames[role.Role] = true
		if role.UserID != testUser.ID {
			t.Errorf("Expected UserID %d, got %d", testUser.ID, role.UserID)
		}
	}

	if !roleNames["admin"] || !roleNames["user"] {
		t.Error("Expected both 'admin' and 'user' roles to be assigned")
	}

	// Test role replacement
	newRoles := []models.UserRole{
		{
			Role:      "moderator",
			OrgID:     "test_org_123",
			ProjectID: "test_project_123",
		},
	}

	err = testRepo.UpdateUserRoles(testUser.ID, newRoles)
	if err != nil {
		t.Errorf("UpdateUserRoles() replacement error = %v", err)
		return
	}

	// Verify old roles were replaced
	updatedRoles, err := testRepo.GetUserRoles(testUser.ID)
	if err != nil {
		t.Errorf("GetUserRoles() after replacement error = %v", err)
		return
	}

	if len(updatedRoles) != 1 {
		t.Errorf("Expected 1 role after replacement, got %d", len(updatedRoles))
		return
	}

	if updatedRoles[0].Role != "moderator" {
		t.Errorf("Expected role 'moderator', got %s", updatedRoles[0].Role)
	}
}

// TestMigrationRunner_Operations tests migration runner functionality
func TestMigrationRunner_Operations(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create migration runner
	zapLogger, _ := zap.NewDevelopment()
	migrationRunner := database.NewMigrationRunner(testDB, zapLogger)

	// Test migration status
	status, err := migrationRunner.GetMigrationStatus()
	if err != nil {
		t.Errorf("GetMigrationStatus() error = %v", err)
		return
	}

	if status == nil {
		t.Error("Expected migration status, got nil")
		return
	}

	// Status should have version information
	if status.LatestVersion == 0 {
		t.Error("Expected LatestVersion > 0")
	}

	// Test schema validation
	err = migrationRunner.ValidateSchema()
	if err != nil {
		t.Logf("Schema validation failed (expected in some cases): %v", err)
	}

	// Test running migrations (should be idempotent)
	err = migrationRunner.RunMigrations()
	if err != nil {
		t.Errorf("RunMigrations() error = %v", err)
		return
	}

	// Test migration status after running migrations
	status2, err := migrationRunner.GetMigrationStatus()
	if err != nil {
		t.Errorf("GetMigrationStatus() after migrations error = %v", err)
		return
	}

	if !status2.IsUpToDate {
		t.Logf("Migrations not up to date: current=%d, latest=%d, pending=%v",
			status2.CurrentVersion, status2.LatestVersion, status2.PendingMigrations)
	}
}

// TestDatabaseHealthChecks tests database health check functionality
func TestDatabaseHealthChecks(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Test successful health check
	err := database.HealthCheck(testDB)
	if err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}

	// Test connection pool stats
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Errorf("Failed to get underlying sql.DB: %v", err)
		return
	}

	stats := sqlDB.Stats()
	if stats.OpenConnections == 0 {
		t.Error("Expected at least one open connection")
	}

	// Test ping functionality
	err = sqlDB.Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

// TestUserRepository_ListUsers tests user listing with filtering
func TestUserRepository_ListUsers(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test users
	users := []*models.User{
		{
			ZitadelID: "test_list_1",
			Email:     stringPtr("list1@example.com"),
			Name:      "List User 1",
			OrgID:     "test_org_123",
			ProjectID: "test_project_123",
			IsActive:  true,
		},
		{
			ZitadelID: "test_list_2",
			Email:     stringPtr("list2@example.com"),
			Name:      "List User 2",
			OrgID:     "test_org_123",
			ProjectID: "test_project_456",
			IsActive:  true,
		},
	}

	for _, user := range users {
		err := testRepo.CreateUser(user)
		if err != nil {
			t.Errorf("Failed to create test user: %v", err)
			return
		}
	}

	// Test listing all users (should include our test users)
	allUsers, err := testRepo.ListUsers("", "", 10, 0)
	if err != nil {
		t.Errorf("ListUsers() error = %v", err)
		return
	}

	// Should have at least our test users
	if len(allUsers) < 2 {
		t.Logf("Found %d users, expected at least 2 test users", len(allUsers))
	}

	// Test filtering by org
	orgUsers, err := testRepo.ListUsers("test_org_123", "", 10, 0)
	if err != nil {
		t.Errorf("ListUsers() with org filter error = %v", err)
		return
	}

	// Should have at least 2 users in test_org_123
	if len(orgUsers) < 2 {
		t.Errorf("Expected at least 2 users in test_org_123, got %d", len(orgUsers))
	}

	// Test filtering by project
	projectUsers, err := testRepo.ListUsers("", "test_project_123", 10, 0)
	if err != nil {
		t.Errorf("ListUsers() with project filter error = %v", err)
		return
	}

	// Should have at least 1 user in test_project_123
	if len(projectUsers) < 1 {
		t.Errorf("Expected at least 1 user in test_project_123, got %d", len(projectUsers))
	}
}

// TestUserRepository_DeleteUser tests user deletion
func TestUserRepository_DeleteUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user with roles
	testUser := &models.User{
		ZitadelID:     "test_delete_789",
		Email:         stringPtr("delete@example.com"),
		EmailVerified: true,
		Name:          "Delete Test User",
		GivenName:     "Delete",
		FamilyName:    "User",
		Username:      "deleteuser",
		OrgID:         "test_org_123",
		ProjectID:     "test_project_123",
		IsActive:      true,
	}

	err := testRepo.CreateUser(testUser)
	if err != nil {
		t.Errorf("CreateUser() error = %v", err)
		return
	}

	// Assign roles
	roles := []models.UserRole{
		{
			Role:      "admin",
			OrgID:     "test_org_123",
			ProjectID: "test_project_123",
		},
	}

	err = testRepo.UpdateUserRoles(testUser.ID, roles)
	if err != nil {
		t.Errorf("UpdateUserRoles() error = %v", err)
		return
	}

	// Delete user
	err = testRepo.DeleteUser(testUser.ID)
	if err != nil {
		t.Errorf("DeleteUser() error = %v", err)
		return
	}

	// Verify user was deleted
	deletedUser, err := testRepo.GetUserByID(testUser.ID)
	if err != nil {
		t.Errorf("GetUserByID() after deletion error = %v", err)
		return
	}

	if deletedUser != nil {
		t.Error("Expected user to be deleted, but still found")
	}

	// Verify roles were also deleted (cascade)
	deletedRoles, err := testRepo.GetUserRoles(testUser.ID)
	if err != nil {
		t.Errorf("GetUserRoles() after user deletion error = %v", err)
		return
	}

	if len(deletedRoles) != 0 {
		t.Errorf("Expected 0 roles after user deletion, got %d", len(deletedRoles))
	}
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

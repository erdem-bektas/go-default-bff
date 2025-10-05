package database

import (
	"fmt"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testMigrationDB     *gorm.DB
	testMigrationRunner MigrationRunner
)

// setupMigrationTestDB creates a test database connection for migration tests
func setupMigrationTestDB(t *testing.T) {
	// Use test database configuration
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "postgres")
	dbname := getEnvOrDefault("DB_NAME", "fiber_app")
	sslmode := getEnvOrDefault("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	var err error
	testMigrationDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Configure connection pool for tests
	if err := ConfigureConnectionPool(testMigrationDB); err != nil {
		t.Fatalf("Failed to configure connection pool: %v", err)
	}

	// Create migration runner
	zapLogger, _ := zap.NewDevelopment()
	testMigrationRunner = NewMigrationRunner(testMigrationDB, zapLogger)
}

// teardownMigrationTestDB cleans up the migration test database
func teardownMigrationTestDB(t *testing.T) {
	if testMigrationDB != nil {
		// Clean up test data - only delete test migration records
		testMigrationDB.Exec("DELETE FROM migrations WHERE name LIKE 'test_%'")

		// Close connection
		sqlDB, err := testMigrationDB.DB()
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

// TestMigrationRunner_GetMigrationStatus tests migration status retrieval
func TestMigrationRunner_GetMigrationStatus(t *testing.T) {
	setupMigrationTestDB(t)
	defer teardownMigrationTestDB(t)

	// Test initial status (no migrations applied)
	status, err := testMigrationRunner.GetMigrationStatus()
	if err != nil {
		t.Errorf("GetMigrationStatus() error = %v", err)
		return
	}

	// Note: CurrentVersion might not be 0 if migrations were already applied in previous tests
	if status.CurrentVersion < 0 {
		t.Errorf("Expected CurrentVersion >= 0, got %d", status.CurrentVersion)
	}

	if status.LatestVersion == 0 {
		t.Error("Expected LatestVersion > 0")
	}

	// If migrations are already applied, IsUpToDate might be true
	if status.CurrentVersion == status.LatestVersion && !status.IsUpToDate {
		t.Error("Expected IsUpToDate to be true when current equals latest version")
	}

	// If migrations are already applied, there might be no pending migrations
	if status.CurrentVersion < status.LatestVersion && len(status.PendingMigrations) == 0 {
		t.Error("Expected pending migrations when current version < latest version")
	}

	// Verify pending migrations contain expected files
	expectedMigrations := map[string]bool{
		"001_create_roles_table.sql":       true,
		"002_create_users_table.sql":       true,
		"003_update_users_for_zitadel.sql": true,
		"004_create_user_roles_table.sql":  true,
	}

	for _, migration := range status.PendingMigrations {
		if !expectedMigrations[migration] {
			t.Errorf("Unexpected pending migration: %s", migration)
		}
	}
}

// TestMigrationRunner_RunMigrations tests migration execution
func TestMigrationRunner_RunMigrations(t *testing.T) {
	setupMigrationTestDB(t)
	defer teardownMigrationTestDB(t)

	// Run migrations
	err := testMigrationRunner.RunMigrations()
	if err != nil {
		t.Errorf("RunMigrations() error = %v", err)
		return
	}

	// Verify migrations were applied
	status, err := testMigrationRunner.GetMigrationStatus()
	if err != nil {
		t.Errorf("GetMigrationStatus() after migrations error = %v", err)
		return
	}

	if !status.IsUpToDate {
		t.Error("Expected IsUpToDate to be true after running migrations")
	}

	if len(status.PendingMigrations) != 0 {
		t.Errorf("Expected 0 pending migrations, got %d", len(status.PendingMigrations))
	}

	if status.CurrentVersion != status.LatestVersion {
		t.Errorf("Expected CurrentVersion (%d) to equal LatestVersion (%d)",
			status.CurrentVersion, status.LatestVersion)
	}

	if status.LastMigrationAt.IsZero() {
		t.Error("Expected LastMigrationAt to be set")
	}

	// Verify migrations table was created and populated
	var migrationCount int64
	err = testMigrationDB.Model(&Migration{}).Count(&migrationCount).Error
	if err != nil {
		t.Errorf("Failed to count migrations: %v", err)
	}

	if migrationCount == 0 {
		t.Error("Expected migration records to be created")
	}
}

// TestMigrationRunner_IdempotentMigrations tests that running migrations multiple times is safe
func TestMigrationRunner_IdempotentMigrations(t *testing.T) {
	setupMigrationTestDB(t)
	defer teardownMigrationTestDB(t)

	// Run migrations first time
	err := testMigrationRunner.RunMigrations()
	if err != nil {
		t.Errorf("First RunMigrations() error = %v", err)
		return
	}

	// Get status after first run
	firstStatus, err := testMigrationRunner.GetMigrationStatus()
	if err != nil {
		t.Errorf("GetMigrationStatus() after first run error = %v", err)
		return
	}

	// Run migrations second time (should be idempotent)
	err = testMigrationRunner.RunMigrations()
	if err != nil {
		t.Errorf("Second RunMigrations() error = %v", err)
		return
	}

	// Get status after second run
	secondStatus, err := testMigrationRunner.GetMigrationStatus()
	if err != nil {
		t.Errorf("GetMigrationStatus() after second run error = %v", err)
		return
	}

	// Verify status is the same
	if firstStatus.CurrentVersion != secondStatus.CurrentVersion {
		t.Errorf("CurrentVersion changed: %d -> %d", firstStatus.CurrentVersion, secondStatus.CurrentVersion)
	}

	if firstStatus.IsUpToDate != secondStatus.IsUpToDate {
		t.Errorf("IsUpToDate changed: %v -> %v", firstStatus.IsUpToDate, secondStatus.IsUpToDate)
	}

	// Verify no duplicate migration records were created
	var migrationCount int64
	err = testMigrationDB.Model(&Migration{}).Count(&migrationCount).Error
	if err != nil {
		t.Errorf("Failed to count migrations: %v", err)
	}

	// Should have exactly 4 migrations (based on getAvailableMigrations)
	expectedCount := int64(4)
	if migrationCount != expectedCount {
		t.Errorf("Expected %d migration records, got %d", expectedCount, migrationCount)
	}
}

// TestMigrationRunner_ValidateSchema tests schema validation functionality
func TestMigrationRunner_ValidateSchema(t *testing.T) {
	setupMigrationTestDB(t)
	defer teardownMigrationTestDB(t)

	// Test validation before migrations (might pass if migrations were already run)
	err := testMigrationRunner.ValidateSchema()
	if err != nil {
		t.Logf("Schema validation failed (expected if migrations not yet run): %v", err)
	}

	// Run migrations
	err = testMigrationRunner.RunMigrations()
	if err != nil {
		t.Errorf("RunMigrations() error = %v", err)
		return
	}

	// Test validation after migrations (should pass)
	err = testMigrationRunner.ValidateSchema()
	if err != nil {
		t.Errorf("ValidateSchema() error after migrations = %v", err)
	}
}

// TestMigrationRunner_MigrationVersionParsing tests migration version parsing
func TestMigrationRunner_MigrationVersionParsing(t *testing.T) {
	setupMigrationTestDB(t)
	defer teardownMigrationTestDB(t)

	// Cast to concrete type to access private method for testing
	runner := testMigrationRunner.(*migrationRunner)

	tests := []struct {
		filename        string
		expectedVersion int
	}{
		{"001_create_roles_table.sql", 1},
		{"002_create_users_table.sql", 2},
		{"010_add_indexes.sql", 10},
		{"123_complex_migration.sql", 123},
		{"invalid_filename.sql", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			version := runner.parseMigrationVersion(tt.filename)
			if version != tt.expectedVersion {
				t.Errorf("parseMigrationVersion(%s) = %d, expected %d",
					tt.filename, version, tt.expectedVersion)
			}
		})
	}
}

// TestMigrationRunner_MigrationTableCreation tests migrations table creation
func TestMigrationRunner_MigrationTableCreation(t *testing.T) {
	setupMigrationTestDB(t)
	defer teardownMigrationTestDB(t)

	// Cast to concrete type to access private method for testing
	runner := testMigrationRunner.(*migrationRunner)

	// Ensure migrations table doesn't exist initially
	if testMigrationDB.Migrator().HasTable(&Migration{}) {
		testMigrationDB.Migrator().DropTable(&Migration{})
	}

	// Test table creation
	err := runner.ensureMigrationsTable()
	if err != nil {
		t.Errorf("ensureMigrationsTable() error = %v", err)
		return
	}

	// Verify table was created
	if !testMigrationDB.Migrator().HasTable(&Migration{}) {
		t.Error("Expected migrations table to be created")
	}

	// Test that calling it again doesn't cause errors (idempotent)
	err = runner.ensureMigrationsTable()
	if err != nil {
		t.Errorf("Second ensureMigrationsTable() error = %v", err)
	}
}

// TestMigrationRunner_GetAppliedMigrations tests applied migrations retrieval
func TestMigrationRunner_GetAppliedMigrations(t *testing.T) {
	setupMigrationTestDB(t)
	defer teardownMigrationTestDB(t)

	// Cast to concrete type to access private method for testing
	runner := testMigrationRunner.(*migrationRunner)

	// Initially should have no applied migrations
	migrations, err := runner.getAppliedMigrations()
	if err != nil {
		t.Errorf("getAppliedMigrations() error = %v", err)
		return
	}

	if len(migrations) != 0 {
		t.Errorf("Expected 0 applied migrations initially, got %d", len(migrations))
	}

	// Manually insert a migration record
	testMigration := Migration{
		Version:   1,
		Name:      "001_test_migration.sql",
		AppliedAt: time.Now(),
	}

	err = testMigrationDB.Create(&testMigration).Error
	if err != nil {
		t.Errorf("Failed to create test migration record: %v", err)
		return
	}

	// Now should have one applied migration
	migrations, err = runner.getAppliedMigrations()
	if err != nil {
		t.Errorf("getAppliedMigrations() after insert error = %v", err)
		return
	}

	if len(migrations) != 1 {
		t.Errorf("Expected 1 applied migration, got %d", len(migrations))
		return
	}

	if migrations[0].Version != 1 {
		t.Errorf("Expected migration version 1, got %d", migrations[0].Version)
	}

	if migrations[0].Name != "001_test_migration.sql" {
		t.Errorf("Expected migration name '001_test_migration.sql', got %s", migrations[0].Name)
	}
}

// TestMigrationRunner_GetAvailableMigrations tests available migrations retrieval
func TestMigrationRunner_GetAvailableMigrations(t *testing.T) {
	setupMigrationTestDB(t)
	defer teardownMigrationTestDB(t)

	// Cast to concrete type to access private method for testing
	runner := testMigrationRunner.(*migrationRunner)

	migrations, err := runner.getAvailableMigrations()
	if err != nil {
		t.Errorf("getAvailableMigrations() error = %v", err)
		return
	}

	if len(migrations) == 0 {
		t.Error("Expected available migrations to be returned")
	}

	// Verify migrations are sorted
	for i := 1; i < len(migrations); i++ {
		prevVersion := runner.parseMigrationVersion(migrations[i-1])
		currVersion := runner.parseMigrationVersion(migrations[i])
		if prevVersion > currVersion {
			t.Errorf("Migrations not sorted: %s (v%d) comes before %s (v%d)",
				migrations[i-1], prevVersion, migrations[i], currVersion)
		}
	}

	// Verify expected migrations are present
	expectedMigrations := []string{
		"001_create_roles_table.sql",
		"002_create_users_table.sql",
		"003_update_users_for_zitadel.sql",
		"004_create_user_roles_table.sql",
	}

	for _, expected := range expectedMigrations {
		found := false
		for _, migration := range migrations {
			if migration == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected migration %s not found in available migrations", expected)
		}
	}
}

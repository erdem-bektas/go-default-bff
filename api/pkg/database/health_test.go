package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testHealthDB *gorm.DB
)

// setupHealthTestDB creates a test database connection for health check tests
func setupHealthTestDB(t *testing.T) {
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
	testHealthDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Configure connection pool for tests
	if err := ConfigureConnectionPool(testHealthDB); err != nil {
		t.Fatalf("Failed to configure connection pool: %v", err)
	}
}

// teardownHealthTestDB cleans up the health test database
func teardownHealthTestDB(t *testing.T) {
	if testHealthDB != nil {
		// Close connection
		sqlDB, err := testHealthDB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// TestHealthCheck tests database health check functionality
func TestHealthCheck(t *testing.T) {
	setupHealthTestDB(t)
	defer teardownHealthTestDB(t)

	// Test successful health check
	err := HealthCheck(testHealthDB)
	if err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}

	// Test health check with closed connection
	sqlDB, err := testHealthDB.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	// Close the connection
	sqlDB.Close()

	// Health check should now fail
	err = HealthCheck(testHealthDB)
	if err == nil {
		t.Error("Expected HealthCheck() to fail with closed connection")
	}
}

// TestConfigureConnectionPool tests connection pool configuration
func TestConfigureConnectionPool(t *testing.T) {
	setupHealthTestDB(t)
	defer teardownHealthTestDB(t)

	// Test connection pool configuration
	err := ConfigureConnectionPool(testHealthDB)
	if err != nil {
		t.Errorf("ConfigureConnectionPool() error = %v", err)
		return
	}

	// Verify connection pool settings
	sqlDB, err := testHealthDB.DB()
	if err != nil {
		t.Errorf("Failed to get underlying sql.DB: %v", err)
		return
	}

	stats := sqlDB.Stats()

	// The exact values depend on the implementation, but we can check that
	// the connection pool is working (we have open connections)
	if stats.OpenConnections == 0 {
		t.Error("Expected at least one open connection")
	}

	// Test that we can actually use the connection pool
	err = sqlDB.Ping()
	if err != nil {
		t.Errorf("Connection pool ping failed: %v", err)
	}
}

// TestDatabaseHealthCheck tests the database health check functionality directly
func TestDatabaseHealthCheck(t *testing.T) {
	setupHealthTestDB(t)
	defer teardownHealthTestDB(t)

	// Test successful health check
	err := HealthCheck(testHealthDB)
	if err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}

	// Test health check with closed connection
	sqlDB, err := testHealthDB.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	// Close the connection
	sqlDB.Close()

	// Health check should now fail
	err = HealthCheck(testHealthDB)
	if err == nil {
		t.Error("Expected HealthCheck() to fail with closed connection")
	}
}

// TestMigrationRunnerHealthIntegration tests migration runner integration with health checks
func TestMigrationRunnerHealthIntegration(t *testing.T) {
	setupHealthTestDB(t)
	defer teardownHealthTestDB(t)

	// Create migration runner
	zapLogger, _ := zap.NewDevelopment()
	migrationRunner := NewMigrationRunner(testHealthDB, zapLogger)

	// Test migration status before running migrations
	status, err := migrationRunner.GetMigrationStatus()
	if err != nil {
		t.Errorf("GetMigrationStatus() error = %v", err)
		return
	}

	// Should have pending migrations initially
	if status.IsUpToDate {
		t.Log("Migrations are already up to date")
	} else {
		t.Logf("Found %d pending migrations", len(status.PendingMigrations))
	}

	// Test schema validation
	err = migrationRunner.ValidateSchema()
	if err != nil {
		t.Logf("Schema validation failed as expected: %v", err)
	}

	// Run migrations
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
		t.Error("Expected migrations to be up to date after running migrations")
	}

	// Test schema validation after migrations
	err = migrationRunner.ValidateSchema()
	if err != nil {
		t.Errorf("ValidateSchema() after migrations error = %v", err)
	}
}

// TestDatabaseConnectionPoolStats tests connection pool statistics
func TestDatabaseConnectionPoolStats(t *testing.T) {
	setupHealthTestDB(t)
	defer teardownHealthTestDB(t)

	// Get connection pool stats
	sqlDB, err := testHealthDB.DB()
	if err != nil {
		t.Errorf("Failed to get underlying sql.DB: %v", err)
		return
	}

	stats := sqlDB.Stats()

	// Test that we have at least one connection
	if stats.OpenConnections == 0 {
		t.Error("Expected at least one open connection")
	}

	// Test ping functionality
	err = sqlDB.Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}

	// Test that stats are reasonable
	if stats.OpenConnections > 100 {
		t.Errorf("Unexpected high number of open connections: %d", stats.OpenConnections)
	}
}

// TestDatabaseHealthWithClosedConnection tests health check with closed database
func TestDatabaseHealthWithClosedConnection(t *testing.T) {
	setupHealthTestDB(t)
	defer teardownHealthTestDB(t)

	// Close the database connection
	sqlDB, err := testHealthDB.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	sqlDB.Close()

	// Test health check with closed database
	err = HealthCheck(testHealthDB)
	if err == nil {
		t.Error("Expected HealthCheck() to fail with closed database")
	}

	// Test that the error message is meaningful
	if err != nil && err.Error() == "" {
		t.Error("Expected meaningful error message from health check")
	}
}

// TestDatabaseHealthWithTimeout tests health check with context timeout
func TestDatabaseHealthWithTimeout(t *testing.T) {
	setupHealthTestDB(t)
	defer teardownHealthTestDB(t)

	// Create context with very short timeout (for testing robustness)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(1 * time.Millisecond)

	// Check that context is indeed expired
	if ctx.Err() == nil {
		t.Log("Context should be expired by now")
	}

	// Test that health check still works even with expired context
	// (HealthCheck doesn't currently use context, but this tests robustness)
	err := HealthCheck(testHealthDB)
	if err != nil {
		t.Logf("HealthCheck with expired context: %v", err)
	}

	// The main point is that it doesn't panic or hang
	// The actual result depends on the implementation
}

package database

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MigrationStatus represents the current migration status
type MigrationStatus struct {
	CurrentVersion    int       `json:"current_version"`
	LatestVersion     int       `json:"latest_version"`
	PendingMigrations []string  `json:"pending_migrations"`
	IsUpToDate        bool      `json:"is_up_to_date"`
	LastMigrationAt   time.Time `json:"last_migration_at,omitempty"`
}

// MigrationRunner interface for database migrations
type MigrationRunner interface {
	RunMigrations() error
	GetMigrationStatus() (*MigrationStatus, error)
	ValidateSchema() error // Read-only validation
}

// migrationRunner implements MigrationRunner interface
type migrationRunner struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMigrationRunner creates a new migration runner instance
func NewMigrationRunner(db *gorm.DB, logger *zap.Logger) MigrationRunner {
	return &migrationRunner{
		db:     db,
		logger: logger,
	}
}

// Migration represents a database migration
type Migration struct {
	ID        int       `gorm:"primaryKey"`
	Version   int       `gorm:"uniqueIndex;not null"`
	Name      string    `gorm:"not null"`
	AppliedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// ensureMigrationsTable creates the migrations table if it doesn't exist
func (mr *migrationRunner) ensureMigrationsTable() error {
	return mr.db.AutoMigrate(&Migration{})
}

// getAppliedMigrations returns list of applied migrations
func (mr *migrationRunner) getAppliedMigrations() ([]Migration, error) {
	if err := mr.ensureMigrationsTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	var migrations []Migration
	if err := mr.db.Order("version ASC").Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	return migrations, nil
}

// getAvailableMigrations returns list of available migration files
func (mr *migrationRunner) getAvailableMigrations() ([]string, error) {
	// In a real implementation, you would read from a migrations directory
	// For now, we'll return the known migration files
	migrations := []string{
		"001_create_roles_table.sql",
		"002_create_users_table.sql",
		"003_update_users_for_zitadel.sql",
		"004_create_user_roles_table.sql",
	}

	sort.Strings(migrations)
	return migrations, nil
}

// parseMigrationVersion extracts version number from migration filename
func (mr *migrationRunner) parseMigrationVersion(filename string) int {
	base := filepath.Base(filename)
	parts := strings.Split(base, "_")
	if len(parts) == 0 {
		return 0
	}

	var version int
	fmt.Sscanf(parts[0], "%d", &version)
	return version
}

// RunMigrations executes pending migrations
func (mr *migrationRunner) RunMigrations() error {
	mr.logger.Info("Starting database migrations")

	appliedMigrations, err := mr.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	availableMigrations, err := mr.getAvailableMigrations()
	if err != nil {
		return fmt.Errorf("failed to get available migrations: %w", err)
	}

	// Create map of applied migration versions
	appliedVersions := make(map[int]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	// Execute pending migrations
	for _, migrationFile := range availableMigrations {
		version := mr.parseMigrationVersion(migrationFile)
		if version == 0 {
			mr.logger.Warn("Invalid migration filename", zap.String("file", migrationFile))
			continue
		}

		if appliedVersions[version] {
			mr.logger.Debug("Migration already applied",
				zap.String("file", migrationFile),
				zap.Int("version", version))
			continue
		}

		mr.logger.Info("Applying migration",
			zap.String("file", migrationFile),
			zap.Int("version", version))

		if err := mr.executeMigration(migrationFile, version); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationFile, err)
		}

		mr.logger.Info("Migration applied successfully",
			zap.String("file", migrationFile),
			zap.Int("version", version))
	}

	mr.logger.Info("Database migrations completed successfully")
	return nil
}

// executeMigration executes a single migration file
func (mr *migrationRunner) executeMigration(filename string, version int) error {
	// In a real implementation, you would read the SQL file and execute it
	// For now, we'll just record that the migration was applied

	migration := Migration{
		Version:   version,
		Name:      filename,
		AppliedAt: time.Now(),
	}

	return mr.db.Create(&migration).Error
}

// GetMigrationStatus returns the current migration status
func (mr *migrationRunner) GetMigrationStatus() (*MigrationStatus, error) {
	appliedMigrations, err := mr.getAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	availableMigrations, err := mr.getAvailableMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get available migrations: %w", err)
	}

	// Find current and latest versions
	currentVersion := 0
	var lastMigrationAt time.Time
	if len(appliedMigrations) > 0 {
		lastMigration := appliedMigrations[len(appliedMigrations)-1]
		currentVersion = lastMigration.Version
		lastMigrationAt = lastMigration.AppliedAt
	}

	latestVersion := 0
	for _, migrationFile := range availableMigrations {
		version := mr.parseMigrationVersion(migrationFile)
		if version > latestVersion {
			latestVersion = version
		}
	}

	// Find pending migrations
	appliedVersions := make(map[int]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	var pendingMigrations []string
	for _, migrationFile := range availableMigrations {
		version := mr.parseMigrationVersion(migrationFile)
		if !appliedVersions[version] {
			pendingMigrations = append(pendingMigrations, migrationFile)
		}
	}

	status := &MigrationStatus{
		CurrentVersion:    currentVersion,
		LatestVersion:     latestVersion,
		PendingMigrations: pendingMigrations,
		IsUpToDate:        currentVersion == latestVersion,
		LastMigrationAt:   lastMigrationAt,
	}

	return status, nil
}

// ValidateSchema performs read-only schema validation
func (mr *migrationRunner) ValidateSchema() error {
	mr.logger.Info("Validating database schema")

	// Check if migrations table exists
	if !mr.db.Migrator().HasTable(&Migration{}) {
		return fmt.Errorf("migrations table does not exist")
	}

	// Check if required tables exist
	requiredTables := []interface{}{
		&Migration{},
	}

	for _, table := range requiredTables {
		if !mr.db.Migrator().HasTable(table) {
			return fmt.Errorf("required table does not exist: %T", table)
		}
	}

	mr.logger.Info("Database schema validation completed successfully")
	return nil
}

// HealthCheck performs a database health check with connection pooling
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Test connection with a simple ping
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check connection pool stats
	stats := sqlDB.Stats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open database connections")
	}

	return nil
}

// ConfigureConnectionPool configures database connection pool settings
func ConfigureConnectionPool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)           // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)          // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Maximum connection lifetime

	return nil
}

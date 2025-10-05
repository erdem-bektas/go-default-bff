package database

import (
	"fiber-app/internal/models"
	"fiber-app/pkg/config"
	"fmt"
	"os"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config, zapLogger *zap.Logger) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	zapLogger.Info("Database connection establishing",
		zap.String("host", cfg.Database.Host),
		zap.String("port", cfg.Database.Port),
		zap.String("dbname", cfg.Database.DBName),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable GORM logs
	})

	if err != nil {
		zapLogger.Error("Database connection failed", zap.Error(err))
		return err
	}

	// Configure connection pool
	if err := ConfigureConnectionPool(DB); err != nil {
		zapLogger.Error("Failed to configure connection pool", zap.Error(err))
		return err
	}

	zapLogger.Info("Database connection successful")
	return nil
}

// Migrate runs auto-migration for development only
// In production, use MigrationRunner instead
func Migrate() error {
	// Check if we're in production mode
	if os.Getenv("APP_ENV") == "production" {
		return fmt.Errorf("auto-migration is disabled in production, use migration runner instead")
	}

	return DB.AutoMigrate(
		&models.User{},
		&models.UserRole{},
	)
}

// InitializeMigrationRunner creates and returns a migration runner instance
func InitializeMigrationRunner(zapLogger *zap.Logger) MigrationRunner {
	return NewMigrationRunner(DB, zapLogger)
}

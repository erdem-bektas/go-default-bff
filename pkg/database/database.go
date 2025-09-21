package database

import (
	"fiber-app/internal/models"
	"fiber-app/pkg/config"
	"fmt"

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

	zapLogger.Info("Database bağlantısı kuruluyor",
		zap.String("host", cfg.Database.Host),
		zap.String("port", cfg.Database.Port),
		zap.String("dbname", cfg.Database.DBName),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // GORM loglarını kapat
	})

	if err != nil {
		zapLogger.Error("Database bağlantısı başarısız", zap.Error(err))
		return err
	}

	zapLogger.Info("Database bağlantısı başarılı")
	return nil
}

func Migrate() error {
	return DB.AutoMigrate(
		&models.Role{},
		&models.User{},
	)
}

func SeedDefaultRoles() error {
	roles := []models.Role{
		{Name: "admin", Description: "System administrator with full access"},
		{Name: "user", Description: "Regular user with limited access"},
		{Name: "moderator", Description: "Moderator with content management access"},
	}

	for _, role := range roles {
		var existingRole models.Role
		if err := DB.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := DB.Create(&role).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}

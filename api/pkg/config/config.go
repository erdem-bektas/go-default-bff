package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port     string
	LogLevel string
	AppEnv   string
	Database DatabaseConfig
	Redis    RedisConfig
	Zitadel  ZitadelConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ZitadelConfig struct {
	Domain       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "3000"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		AppEnv:   getEnv("APP_ENV", "development"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "fiber_app"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Zitadel: ZitadelConfig{
			Domain:       getEnv("ZITADEL_DOMAIN", "http://localhost:8080"),
			ClientID:     getEnv("ZITADEL_CLIENT_ID", ""),
			ClientSecret: getEnv("ZITADEL_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("ZITADEL_REDIRECT_URL", "http://localhost:3003/auth/callback"),
			Scopes:       []string{"openid", "profile", "email", "urn:zitadel:iam:org:project:roles"},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

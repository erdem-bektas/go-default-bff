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
	Security SecurityConfig
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
	Domain        string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	Scopes        []string
	Issuer        string
	JWKSURL       string
	OrgID         string
	ProjectID     string
	RoleClaimName string
}

type SecurityConfig struct {
	SessionEncryptionKey string
	CSRFSecretKey        string
	CookieDomain         string
	CookieSecure         bool
	CookieSameSite       string
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
			Domain:        getEnv("ZITADEL_DOMAIN", "http://localhost:8080"),
			ClientID:      getEnv("ZITADEL_CLIENT_ID", ""),
			ClientSecret:  getEnv("ZITADEL_CLIENT_SECRET", ""),
			RedirectURL:   getEnv("ZITADEL_REDIRECT_URL", "http://localhost:3003/auth/callback"),
			Scopes:        []string{"openid", "profile", "email", "offline_access"},
			Issuer:        getEnv("ZITADEL_ISSUER", "http://localhost:8080"),
			JWKSURL:       getEnv("ZITADEL_JWKS_URL", ""),
			OrgID:         getEnv("ZITADEL_ORG_ID", ""),
			ProjectID:     getEnv("ZITADEL_PROJECT_ID", ""),
			RoleClaimName: getEnv("ZITADEL_ROLE_CLAIM_NAME", "urn:zitadel:iam:org:project:roles"),
		},
		Security: SecurityConfig{
			SessionEncryptionKey: getEnv("SESSION_ENCRYPTION_KEY", "dev-key-32-bytes-for-aes-256-gcm"),
			CSRFSecretKey:        getEnv("CSRF_SECRET_KEY", "dev-csrf-secret-32-bytes-for-hmac"),
			CookieDomain:         getEnv("COOKIE_DOMAIN", "localhost"),
			CookieSecure:         getEnvAsBool("COOKIE_SECURE", false),
			CookieSameSite:       getEnv("COOKIE_SAME_SITE", "Lax"),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

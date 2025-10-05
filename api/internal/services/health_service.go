package services

import (
	"context"
	"fiber-app/pkg/database"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HealthService provides health check functionality
type HealthService interface {
	LivenessCheck() *HealthCheckResult
	ReadinessCheck(ctx context.Context) *ReadinessCheckResult
}

// HealthCheckResult represents a simple health check result
type HealthCheckResult struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

// ReadinessCheckResult represents detailed readiness check results
type ReadinessCheckResult struct {
	Status     string                     `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentStatus `json:"components"`
}

// ComponentStatus represents the status of a system component
type ComponentStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// healthService implements HealthService interface
type healthService struct {
	db              *gorm.DB
	migrationRunner database.MigrationRunner
	logger          *zap.Logger
	startTime       time.Time
}

// NewHealthService creates a new health service instance
func NewHealthService(db *gorm.DB, migrationRunner database.MigrationRunner, logger *zap.Logger) HealthService {
	return &healthService{
		db:              db,
		migrationRunner: migrationRunner,
		logger:          logger,
		startTime:       time.Now(),
	}
}

// LivenessCheck performs a basic liveness check (fast response <100ms)
func (hs *healthService) LivenessCheck() *HealthCheckResult {
	uptime := time.Since(hs.startTime).String()

	return &HealthCheckResult{
		Status:    "ok",
		Timestamp: time.Now(),
		Uptime:    uptime,
	}
}

// ReadinessCheck performs detailed dependency validation
func (hs *healthService) ReadinessCheck(ctx context.Context) *ReadinessCheckResult {
	components := make(map[string]ComponentStatus)
	overallStatus := "ok"

	// Check PostgreSQL connection
	pgStatus := hs.checkPostgreSQL(ctx)
	components["postgresql"] = pgStatus
	if pgStatus.Status != "ok" {
		overallStatus = "degraded"
	}

	// Check migration status (tolerate schema version mismatches during deployment)
	migrationStatus := hs.checkMigrationStatus(ctx)
	components["migrations"] = migrationStatus
	// Note: Don't fail readiness check for migration mismatches during deployment

	// Check Redis connection (if Redis client is available)
	// This would be implemented when Redis is added
	// redisStatus := hs.checkRedis(ctx)
	// components["redis"] = redisStatus

	// Check JWKS endpoint reachability (not cache state)
	jwksStatus := hs.checkJWKSEndpoint(ctx)
	components["jwks"] = jwksStatus
	if jwksStatus.Status != "ok" {
		overallStatus = "degraded"
	}

	return &ReadinessCheckResult{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Components: components,
	}
}

// checkPostgreSQL checks PostgreSQL connection and basic functionality
func (hs *healthService) checkPostgreSQL(ctx context.Context) ComponentStatus {
	if err := database.HealthCheck(hs.db); err != nil {
		hs.logger.Error("PostgreSQL health check failed", zap.Error(err))
		return ComponentStatus{
			Status: "error",
			Error:  err.Error(),
		}
	}

	return ComponentStatus{
		Status:  "ok",
		Message: "PostgreSQL connection healthy",
	}
}

// checkMigrationStatus checks migration status (read-only validation)
func (hs *healthService) checkMigrationStatus(ctx context.Context) ComponentStatus {
	status, err := hs.migrationRunner.GetMigrationStatus()
	if err != nil {
		hs.logger.Error("Migration status check failed", zap.Error(err))
		return ComponentStatus{
			Status: "error",
			Error:  err.Error(),
		}
	}

	if !status.IsUpToDate {
		// In production, this is expected during deployment
		// Don't fail the readiness check, just report the status
		message := fmt.Sprintf("Migrations pending: %d (current: %d, latest: %d)",
			len(status.PendingMigrations), status.CurrentVersion, status.LatestVersion)

		return ComponentStatus{
			Status:  "warning",
			Message: message,
		}
	}

	return ComponentStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Migrations up to date (version: %d)", status.CurrentVersion),
	}
}

// checkJWKSEndpoint checks JWKS endpoint reachability (not cache state)
func (hs *healthService) checkJWKSEndpoint(ctx context.Context) ComponentStatus {
	// This is a placeholder - would be implemented when JWKS service is available
	// For now, we'll simulate a check

	// In a real implementation, you would:
	// 1. Make an HTTP request to the JWKS endpoint
	// 2. Check if it returns a valid response (not necessarily cached keys)
	// 3. Validate the response structure

	// Simulated check - replace with actual JWKS endpoint validation
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// This would use the actual JWKS URL from configuration
	jwksURL := "https://example.zitadel.cloud/.well-known/openid_configuration"

	req, err := http.NewRequestWithContext(ctx, "GET", jwksURL, nil)
	if err != nil {
		return ComponentStatus{
			Status: "error",
			Error:  fmt.Sprintf("Failed to create JWKS request: %v", err),
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ComponentStatus{
			Status: "error",
			Error:  fmt.Sprintf("JWKS endpoint unreachable: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ComponentStatus{
			Status: "error",
			Error:  fmt.Sprintf("JWKS endpoint returned status: %d", resp.StatusCode),
		}
	}

	return ComponentStatus{
		Status:  "ok",
		Message: "JWKS endpoint reachable",
	}
}

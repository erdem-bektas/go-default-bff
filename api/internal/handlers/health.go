package handlers

import (
	"fiber-app/internal/services"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var (
	startTime     = time.Now()
	healthService services.HealthService
)

// SetHealthService sets the health service instance
func SetHealthService(hs services.HealthService) {
	healthService = hs
}

// HealthCheck - General health check
// @Summary Health check
// @Description Application general health status
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/health [get]
func HealthCheck(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Health check endpoint called",
		zap.String("trace_id", traceID),
	)

	uptime := time.Since(startTime)

	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"uptime":    uptime.String(),
		"trace_id":  traceID,
		"service":   "fiber-app",
		"version":   "1.0.0",
	})
}

// ReadinessCheck - Production-ready readiness check with dependency validation
// @Summary Readiness check
// @Description Checks if service is ready to handle requests (PostgreSQL, Redis, JWKS)
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} services.ReadinessCheckResult
// @Failure 503 {object} services.ReadinessCheckResult
// @Router /readyz [get]
func ReadinessCheck(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Readiness check endpoint called",
		zap.String("trace_id", traceID),
	)

	if healthService == nil {
		zapLogger.Error("Health service not initialized")
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":    "error",
			"timestamp": time.Now().UTC(),
			"error":     "health service not initialized",
			"trace_id":  traceID,
		})
	}

	result := healthService.ReadinessCheck(c.Context())

	httpStatus := fiber.StatusOK
	if result.Status != "ok" {
		httpStatus = fiber.StatusServiceUnavailable
	}

	return c.Status(httpStatus).JSON(result)
}

// LivenessCheck - Fast liveness check (<100ms)
// @Summary Liveness check
// @Description Fast process health check for Kubernetes liveness probe
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} services.HealthCheckResult
// @Router /healthz [get]
func LivenessCheck(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Debug("Liveness check endpoint called",
		zap.String("trace_id", traceID),
	)

	if healthService == nil {
		// Fallback to basic check if health service not available
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"uptime":    time.Since(startTime).String(),
		})
	}

	result := healthService.LivenessCheck()
	return c.JSON(result)
}

// DetailedHealthCheck - Detailed health check with system metrics
// @Summary Detailed health check
// @Description Detailed health check with memory and goroutine information
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/health/detailed [get]
func DetailedHealthCheck(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Detailed health check endpoint called",
		zap.String("trace_id", traceID),
	)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return c.JSON(fiber.Map{
		"status":     "alive",
		"timestamp":  time.Now().UTC(),
		"uptime":     time.Since(startTime).String(),
		"goroutines": runtime.NumGoroutine(),
		"memory_mb":  m.Alloc / 1024 / 1024,
		"gc_cycles":  m.NumGC,
		"trace_id":   traceID,
	})
}

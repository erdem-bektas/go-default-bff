package handlers

import (
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var startTime = time.Now()

// HealthCheck - Genel sağlık kontrolü
func HealthCheck(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Health check endpoint çağrıldı",
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

// ReadinessCheck - Servisin hazır olup olmadığını kontrol eder
func ReadinessCheck(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Readiness check endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	// Burada database, redis vb. bağlantıları kontrol edilebilir
	checks := map[string]string{
		"database": "ok",
		"cache":    "ok",
		"storage":  "ok",
	}

	allHealthy := true
	for _, status := range checks {
		if status != "ok" {
			allHealthy = false
			break
		}
	}

	status := "ready"
	httpStatus := fiber.StatusOK
	if !allHealthy {
		status = "not_ready"
		httpStatus = fiber.StatusServiceUnavailable
	}

	return c.Status(httpStatus).JSON(fiber.Map{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"checks":    checks,
		"trace_id":  traceID,
	})
}

// LivenessCheck - Servisin yaşayıp yaşamadığını kontrol eder
func LivenessCheck(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Liveness check endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return c.JSON(fiber.Map{
		"status":     "alive",
		"timestamp":  time.Now().UTC(),
		"goroutines": runtime.NumGoroutine(),
		"memory_mb":  m.Alloc / 1024 / 1024,
		"gc_cycles":  m.NumGC,
		"trace_id":   traceID,
	})
}

package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var zapLogger *zap.Logger

// SetLogger - Handler'lar için logger'ı set eder
func SetLogger(l *zap.Logger) {
	zapLogger = l
}

// getTraceID - Context'ten trace_id'yi alır
func getTraceID(c *fiber.Ctx) string {
	if traceID := c.Locals("trace_id"); traceID != nil {
		return traceID.(string)
	}
	return "unknown"
}

// Home - Ana sayfa
// @Summary Ana sayfa
// @Description Uygulama ana sayfası ve endpoint listesi
// @Tags General
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func Home(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Ana sayfa ziyaret edildi",
		zap.String("trace_id", traceID),
	)

	return c.JSON(fiber.Map{
		"message":   "Merhaba Fiber! 🚀",
		"service":   "fiber-app",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"endpoints": fiber.Map{
			"health":  "/api/v1/health",
			"metrics": "/api/v1/metrics",
			"info":    "/api/v1/info",
			"users":   "/api/v1/users",
			"roles":   "/api/v1/roles",
			"cache":   "/api/v1/cache",
			"auth":    "/auth",
			"test":    "/api/v1/test",
		},
		"documentation": fiber.Map{
			"swagger_ui":   "/docs",
			"swagger_json": "/swagger.json",
		},
		"trace_id": traceID,
	})
}

// Ping - Basit ping endpoint
// @Summary Ping endpoint
// @Description Basit ping-pong testi
// @Tags General
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /ping [get]
func Ping(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Ping endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	return c.JSON(fiber.Map{
		"message":   "pong",
		"timestamp": time.Now().UTC(),
		"trace_id":  traceID,
	})
}

package handlers

import (
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// GetAppInfo - Uygulama bilgileri
func GetAppInfo(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("App info endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	appInfo := fiber.Map{
		"name":        "fiber-app",
		"version":     "1.0.0",
		"description": "Go Fiber app with hot reload, zap logger and trace_id support",
		"author":      "Developer",
		"license":     "MIT",
		"repository":  "https://github.com/example/fiber-app",
		"build_info": fiber.Map{
			"go_version": runtime.Version(),
			"goos":       runtime.GOOS,
			"goarch":     runtime.GOARCH,
			"compiler":   runtime.Compiler,
			"build_time": startTime.UTC(),
		},
		"features": []string{
			"hot_reload",
			"zap_logger",
			"trace_id",
			"health_checks",
			"metrics",
			"graceful_shutdown",
		},
		"endpoints": fiber.Map{
			"health":  "/api/v1/health",
			"metrics": "/api/v1/metrics",
			"info":    "/api/v1/info",
			"test":    "/api/v1/test",
		},
		"timestamp": time.Now().UTC(),
		"trace_id":  traceID,
	}

	return c.JSON(appInfo)
}

// GetVersion - Sadece versiyon bilgisi
func GetVersion(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Version endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	return c.JSON(fiber.Map{
		"version":    "1.0.0",
		"go_version": runtime.Version(),
		"build_time": startTime.UTC(),
		"trace_id":   traceID,
	})
}

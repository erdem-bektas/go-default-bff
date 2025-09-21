// @title Fiber App API
// @version 1.0.0
// @description Go Fiber app with PostgreSQL, GORM, Zap Logger, Trace ID and Role-based User Management
// @contact.name API Support
// @contact.email support@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:3003
// @BasePath /
// @schemes http
package main

import (
	"fiber-app/internal/handlers"
	"fiber-app/internal/middleware"
	"fiber-app/internal/services"
	"fiber-app/pkg/cache"
	"fiber-app/pkg/config"
	"fiber-app/pkg/database"
	"fiber-app/router"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var zapLogger *zap.Logger

func main() {
	// .env dosyasını yükle
	godotenv.Load()

	// Config yükle
	cfg := config.Load()

	// Zap logger'ı başlat
	var err error
	zapLogger, err = zap.NewProduction()
	if err != nil {
		log.Fatal("Zap logger başlatılamadı:", err)
	}
	defer zapLogger.Sync()

	// Database bağlantısı
	if err := database.Connect(cfg, zapLogger); err != nil {
		log.Fatal("Database bağlantısı başarısız:", err)
	}

	// Database migration
	if err := database.Migrate(); err != nil {
		zapLogger.Fatal("Database migration başarısız", zap.Error(err))
	}

	// Default rolleri oluştur
	if err := database.SeedDefaultRoles(); err != nil {
		zapLogger.Fatal("Default roles oluşturulamadı", zap.Error(err))
	}

	// Redis bağlantısı
	if err := cache.Connect(cfg, zapLogger); err != nil {
		zapLogger.Warn("Redis bağlantısı başarısız, cache devre dışı", zap.Error(err))
	} else {
		// Cache service'i başlat
		cacheService := services.NewCacheService(zapLogger)
		handlers.SetCacheService(cacheService)
		zapLogger.Info("Cache service başlatıldı")
	}

	// Auth service'i başlat
	if cfg.Zitadel.ClientID != "" && cfg.Zitadel.ClientSecret != "" {
		authService := services.NewAuthService(&cfg.Zitadel, zapLogger)
		handlers.SetAuthService(authService)

		// Auth middleware'i başlat
		authMiddleware := middleware.NewAuthMiddleware(authService, zapLogger)
		_ = authMiddleware // Şimdilik kullanılmıyor, route'larda kullanılacak

		zapLogger.Info("Auth service başlatıldı",
			zap.String("domain", cfg.Zitadel.Domain),
			zap.String("redirect_url", cfg.Zitadel.RedirectURL),
		)
	} else {
		zapLogger.Warn("Zitadel yapılandırılmamış, auth devre dışı")
	}

	// Fiber app oluştur
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})

	// Handler'lara logger'ı set et
	handlers.SetLogger(zapLogger)

	// Middleware'ler
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(traceIDMiddleware)

	// Routes
	router.SetupRoutes(app)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			zapLogger.Fatal("Server başlatılamadı", zap.Error(err))
		}
	}()

	zapLogger.Info("Server başlatıldı",
		zap.String("port", cfg.Port),
		zap.String("env", cfg.AppEnv),
	)

	<-c
	zapLogger.Info("Server kapatılıyor...")
	app.Shutdown()
}

// Trace ID middleware - her request için unique trace_id oluşturur
func traceIDMiddleware(c *fiber.Ctx) error {
	traceID := uuid.New().String()
	c.Locals("trace_id", traceID)
	c.Set("X-Trace-ID", traceID)

	zapLogger.Info("Request başladı",
		zap.String("trace_id", traceID),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("ip", c.IP()),
	)

	return c.Next()
}

// Error handler
func errorHandler(c *fiber.Ctx, err error) error {
	traceID := getTraceID(c)

	zapLogger.Error("Request hatası",
		zap.String("trace_id", traceID),
		zap.Error(err),
		zap.String("path", c.Path()),
	)

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":    "Internal Server Error",
		"trace_id": traceID,
	})
}

// Trace ID helper
func getTraceID(c *fiber.Ctx) string {
	if traceID := c.Locals("trace_id"); traceID != nil {
		return traceID.(string)
	}
	return "unknown"
}

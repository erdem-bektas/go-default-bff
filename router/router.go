package router

import (
	"fiber-app/router/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// API v1 group
	api := app.Group("/api/v1")

	// Health routes
	health := api.Group("/health")
	health.Get("/", handlers.HealthCheck)
	health.Get("/ready", handlers.ReadinessCheck)
	health.Get("/live", handlers.LivenessCheck)

	// Metrics routes
	metrics := api.Group("/metrics")
	metrics.Get("/", handlers.GetMetrics)
	metrics.Get("/system", handlers.GetSystemMetrics)

	// App info routes
	info := api.Group("/info")
	info.Get("/", handlers.GetAppInfo)
	info.Get("/version", handlers.GetVersion)

	// User routes
	users := api.Group("/users")
	users.Get("/", handlers.GetUsers)
	users.Get("/:id", handlers.GetUser)
	users.Post("/", handlers.CreateUser)
	users.Put("/:id", handlers.UpdateUser)
	users.Delete("/:id", handlers.DeleteUser)

	// Test routes
	test := api.Group("/test")
	test.Get("/", handlers.TestGet)
	test.Post("/", handlers.TestPost)
	test.Get("/error", handlers.TestError)

	// Root routes
	app.Get("/", handlers.Home)
	app.Get("/ping", handlers.Ping)
}

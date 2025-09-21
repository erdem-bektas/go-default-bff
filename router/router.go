package router

import (
	_ "fiber-app/docs"
	"fiber-app/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func SetupRoutes(app *fiber.App) {
	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

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

	// Role routes
	roles := api.Group("/roles")
	roles.Get("/", handlers.GetRoles)
	roles.Get("/:id", handlers.GetRole)
	roles.Post("/", handlers.CreateRole)
	roles.Put("/:id", handlers.UpdateRole)
	roles.Delete("/:id", handlers.DeleteRole)

	// Cache routes
	cache := api.Group("/cache")
	cache.Get("/stats", handlers.GetCacheStats)
	cache.Post("/flush", handlers.FlushCache)
	cache.Get("/keys", handlers.GetCacheKeys)
	cache.Delete("/keys/:key", handlers.DeleteCacheKey)

	// Test routes
	test := api.Group("/test")
	test.Get("/", handlers.TestGet)
	test.Post("/", handlers.TestPost)
	test.Get("/error", handlers.TestError)

	// Swagger JSON endpoint
	app.Get("/swagger.json", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"swagger": "2.0",
			"info": fiber.Map{
				"title":       "Fiber App API",
				"description": "Go Fiber app with PostgreSQL, GORM, Zap Logger, Trace ID and Role-based User Management",
				"version":     "1.0.0",
			},
			"host":     "localhost:3003",
			"basePath": "/",
			"schemes":  []string{"http"},
			"paths": fiber.Map{
				"/": fiber.Map{
					"get": fiber.Map{
						"summary":     "Ana sayfa",
						"description": "Uygulama ana sayfası ve endpoint listesi",
						"tags":        []string{"General"},
						"responses": fiber.Map{
							"200": fiber.Map{
								"description": "Başarılı",
							},
						},
					},
				},
				"/api/v1/users": fiber.Map{
					"get": fiber.Map{
						"summary":     "Kullanıcıları listele",
						"description": "Sayfalama ve arama desteği ile kullanıcıları listele",
						"tags":        []string{"Users"},
						"parameters": []fiber.Map{
							{
								"name":        "page",
								"in":          "query",
								"description": "Sayfa numarası",
								"type":        "integer",
								"default":     1,
							},
							{
								"name":        "limit",
								"in":          "query",
								"description": "Sayfa başına kayıt sayısı",
								"type":        "integer",
								"default":     10,
							},
							{
								"name":        "search",
								"in":          "query",
								"description": "Arama terimi",
								"type":        "string",
							},
						},
						"responses": fiber.Map{
							"200": fiber.Map{"description": "Başarılı"},
							"500": fiber.Map{"description": "Sunucu hatası"},
						},
					},
					"post": fiber.Map{
						"summary":     "Yeni kullanıcı oluştur",
						"description": "Yeni kullanıcı kaydı oluştur",
						"tags":        []string{"Users"},
						"responses": fiber.Map{
							"201": fiber.Map{"description": "Oluşturuldu"},
							"400": fiber.Map{"description": "Geçersiz istek"},
							"409": fiber.Map{"description": "Çakışma"},
							"500": fiber.Map{"description": "Sunucu hatası"},
						},
					},
				},
				"/api/v1/users/{id}": fiber.Map{
					"get": fiber.Map{
						"summary":     "Kullanıcı detayı",
						"description": "ID ile kullanıcı detayını getir",
						"tags":        []string{"Users"},
						"parameters": []fiber.Map{
							{
								"name":        "id",
								"in":          "path",
								"description": "User ID (UUID)",
								"required":    true,
								"type":        "string",
							},
						},
						"responses": fiber.Map{
							"200": fiber.Map{"description": "Başarılı"},
							"404": fiber.Map{"description": "Bulunamadı"},
							"500": fiber.Map{"description": "Sunucu hatası"},
						},
					},
					"put": fiber.Map{
						"summary":     "Kullanıcı güncelle",
						"description": "Mevcut kullanıcı bilgilerini güncelle",
						"tags":        []string{"Users"},
						"parameters": []fiber.Map{
							{
								"name":        "id",
								"in":          "path",
								"description": "User ID (UUID)",
								"required":    true,
								"type":        "string",
							},
						},
						"responses": fiber.Map{
							"200": fiber.Map{"description": "Başarılı"},
							"404": fiber.Map{"description": "Bulunamadı"},
							"500": fiber.Map{"description": "Sunucu hatası"},
						},
					},
					"delete": fiber.Map{
						"summary":     "Kullanıcı sil",
						"description": "Kullanıcıyı sistemden sil",
						"tags":        []string{"Users"},
						"parameters": []fiber.Map{
							{
								"name":        "id",
								"in":          "path",
								"description": "User ID (UUID)",
								"required":    true,
								"type":        "string",
							},
						},
						"responses": fiber.Map{
							"200": fiber.Map{"description": "Başarılı"},
							"404": fiber.Map{"description": "Bulunamadı"},
							"500": fiber.Map{"description": "Sunucu hatası"},
						},
					},
				},
				"/api/v1/roles": fiber.Map{
					"get": fiber.Map{
						"summary":     "Rolleri listele",
						"description": "Sayfalama desteği ile rolleri listele",
						"tags":        []string{"Roles"},
						"responses": fiber.Map{
							"200": fiber.Map{"description": "Başarılı"},
						},
					},
					"post": fiber.Map{
						"summary":     "Yeni rol oluştur",
						"description": "Yeni rol kaydı oluştur",
						"tags":        []string{"Roles"},
						"responses": fiber.Map{
							"201": fiber.Map{"description": "Oluşturuldu"},
						},
					},
				},
				"/api/v1/health": fiber.Map{
					"get": fiber.Map{
						"summary":     "Sağlık kontrolü",
						"description": "Uygulamanın genel sağlık durumu",
						"tags":        []string{"Health"},
						"responses": fiber.Map{
							"200": fiber.Map{"description": "Sağlıklı"},
						},
					},
				},
				"/api/v1/metrics": fiber.Map{
					"get": fiber.Map{
						"summary":     "Uygulama metrikleri",
						"description": "Temel uygulama performans metrikleri",
						"tags":        []string{"Metrics"},
						"responses": fiber.Map{
							"200": fiber.Map{"description": "Başarılı"},
						},
					},
				},
			},
		})
	})

	// Simple Swagger UI endpoint
	app.Get("/docs", func(c *fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/swagger.json',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>`
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	// Auth routes
	auth := app.Group("/auth")
	auth.Get("/login", handlers.Login)
	auth.Get("/login/redirect", handlers.LoginRedirect)
	auth.Get("/callback", handlers.Callback)
	auth.Post("/logout", handlers.Logout)
	auth.Get("/profile", handlers.Profile)

	// Root routes
	app.Get("/", handlers.Home)
	app.Get("/ping", handlers.Ping)
}

package middleware

import (
	"fiber-app/internal/services"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	authService *services.AuthService
	logger      *zap.Logger
}

func NewAuthMiddleware(authService *services.AuthService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth - Authentication gerekli
func (am *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := getTraceID(c)

		// Authorization header'ını kontrol et
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			am.logger.Warn("Missing authorization header",
				zap.String("trace_id", traceID),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":    "Authorization header gerekli",
				"trace_id": traceID,
			})
		}

		// Bearer token formatını kontrol et
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			am.logger.Warn("Invalid authorization header format",
				zap.String("trace_id", traceID),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":    "Geçersiz authorization header formatı",
				"trace_id": traceID,
			})
		}

		token := tokenParts[1]

		// Token'ı validate et
		claims, err := am.authService.ValidateToken(token)
		if err != nil {
			am.logger.Warn("Token validation failed",
				zap.String("trace_id", traceID),
				zap.Error(err),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":    "Geçersiz token",
				"trace_id": traceID,
			})
		}

		// User bilgilerini context'e ekle
		c.Locals("user_id", claims.Sub)
		c.Locals("user_name", claims.Name)
		c.Locals("user_email", claims.Email)
		c.Locals("user_roles", claims.Roles)

		am.logger.Debug("User authenticated",
			zap.String("trace_id", traceID),
			zap.String("user_id", claims.Sub),
			zap.String("email", claims.Email),
			zap.Strings("roles", claims.Roles),
		)

		return c.Next()
	}
}

// RequireRole - Belirli rol gerekli
func (am *AuthMiddleware) RequireRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := getTraceID(c)

		// Önce authentication kontrolü
		if err := am.RequireAuth()(c); err != nil {
			return err
		}

		// User rollerini al
		userRoles, ok := c.Locals("user_roles").([]string)
		if !ok {
			am.logger.Error("Failed to get user roles from context",
				zap.String("trace_id", traceID),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":    "Rol bilgisi alınamadı",
				"trace_id": traceID,
			})
		}

		// Gerekli rolü kontrol et
		hasRole := false
		for _, role := range userRoles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			am.logger.Warn("Insufficient permissions",
				zap.String("trace_id", traceID),
				zap.String("required_role", requiredRole),
				zap.Strings("user_roles", userRoles),
			)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":         "Yetersiz yetki",
				"required_role": requiredRole,
				"trace_id":      traceID,
			})
		}

		am.logger.Debug("Role check passed",
			zap.String("trace_id", traceID),
			zap.String("required_role", requiredRole),
		)

		return c.Next()
	}
}

// RequireAnyRole - Herhangi bir rolden birini gerekli
func (am *AuthMiddleware) RequireAnyRole(requiredRoles []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := getTraceID(c)

		// Önce authentication kontrolü
		if err := am.RequireAuth()(c); err != nil {
			return err
		}

		// User rollerini al
		userRoles, ok := c.Locals("user_roles").([]string)
		if !ok {
			am.logger.Error("Failed to get user roles from context",
				zap.String("trace_id", traceID),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":    "Rol bilgisi alınamadı",
				"trace_id": traceID,
			})
		}

		// Herhangi bir gerekli rolü kontrol et
		hasAnyRole := false
		for _, userRole := range userRoles {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole {
					hasAnyRole = true
					break
				}
			}
			if hasAnyRole {
				break
			}
		}

		if !hasAnyRole {
			am.logger.Warn("Insufficient permissions",
				zap.String("trace_id", traceID),
				zap.Strings("required_roles", requiredRoles),
				zap.Strings("user_roles", userRoles),
			)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":          "Yetersiz yetki",
				"required_roles": requiredRoles,
				"trace_id":       traceID,
			})
		}

		am.logger.Debug("Role check passed",
			zap.String("trace_id", traceID),
			zap.Strings("required_roles", requiredRoles),
		)

		return c.Next()
	}
}

// Optional auth - Token varsa validate et, yoksa devam et
func (am *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Next()
		}

		token := tokenParts[1]
		claims, err := am.authService.ValidateToken(token)
		if err != nil {
			return c.Next()
		}

		// User bilgilerini context'e ekle
		c.Locals("user_id", claims.Sub)
		c.Locals("user_name", claims.Name)
		c.Locals("user_email", claims.Email)
		c.Locals("user_roles", claims.Roles)

		return c.Next()
	}
}

// getTraceID - Context'ten trace_id'yi alır
func getTraceID(c *fiber.Ctx) string {
	if traceID := c.Locals("trace_id"); traceID != nil {
		return traceID.(string)
	}
	return "unknown"
}

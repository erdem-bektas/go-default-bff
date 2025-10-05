package handlers

import (
	"context"
	"fiber-app/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var authService *services.AuthService

// SetAuthService - Auth service'i set eder
func SetAuthService(as *services.AuthService) {
	authService = as
}

// Login - OAuth2 login başlat
// @Summary OAuth2 Login
// @Description Zitadel OAuth2 login işlemini başlatır (PKCE)
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/login [get]
func Login(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Login endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	if authService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Auth service yapılandırılmamış",
			"trace_id": traceID,
		})
	}

	// OAuth2 authorization URL oluştur (PKCE ile)
	authResponse, err := authService.GenerateAuthURL()
	if err != nil {
		zapLogger.Error("Auth URL oluşturulamadı",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Auth URL oluşturulamadı",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Auth URL oluşturuldu",
		zap.String("trace_id", traceID),
		zap.String("state", authResponse.State),
	)

	return c.JSON(fiber.Map{
		"auth_url": authResponse.URL,
		"state":    authResponse.State,
		"message":  "Bu URL'ye yönlendirilerek giriş yapabilirsiniz",
		"trace_id": traceID,
	})
}

// LoginRedirect - OAuth2 login'e yönlendir
// @Summary OAuth2 Login Redirect
// @Description Zitadel OAuth2 login sayfasına yönlendirir (PKCE)
// @Tags Auth
// @Accept json
// @Produce json
// @Router /auth/login/redirect [get]
func LoginRedirect(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	if authService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Auth service yapılandırılmamış",
			"trace_id": traceID,
		})
	}

	authResponse, err := authService.GenerateAuthURL()
	if err != nil {
		zapLogger.Error("Auth URL oluşturulamadı",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Auth URL oluşturulamadı",
			"trace_id": traceID,
		})
	}

	return c.Redirect(authResponse.URL)
}

// Callback - OAuth2 callback
// @Summary OAuth2 Callback
// @Description OAuth2 callback endpoint'i (PKCE)
// @Tags Auth
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/callback [get]
func Callback(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	code := c.Query("code")
	state := c.Query("state")

	zapLogger.Info("Auth callback çağrıldı",
		zap.String("trace_id", traceID),
		zap.String("state", state),
		zap.Bool("has_code", code != ""),
	)

	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Authorization code gerekli",
			"trace_id": traceID,
		})
	}

	if state == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "State parameter gerekli",
			"trace_id": traceID,
		})
	}

	if authService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Auth service yapılandırılmamış",
			"trace_id": traceID,
		})
	}

	ctx := context.Background()

	// Authorization code'u token ile değiştir (PKCE validation dahil)
	token, err := authService.ExchangeCodeForToken(ctx, code, state)
	if err != nil {
		zapLogger.Error("Token exchange başarısız",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Token exchange başarısız",
			"trace_id": traceID,
		})
	}

	// Kullanıcı bilgilerini al
	userInfo, err := authService.GetUserInfo(ctx, token)
	if err != nil {
		zapLogger.Error("User info alınamadı",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "User info alınamadı",
			"trace_id": traceID,
		})
	}

	// Session oluştur
	session, err := authService.CreateSession(userInfo)
	if err != nil {
		zapLogger.Error("Session oluşturulamadı",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Session oluşturulamadı",
			"trace_id": traceID,
		})
	}

	// HttpOnly cookie set et
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		Secure:   false,        // M0: HTTP için false, production'da true olmalı
		HTTPOnly: true,
		SameSite: "Lax",
	})

	zapLogger.Info("User başarıyla giriş yaptı",
		zap.String("trace_id", traceID),
		zap.String("user_id", userInfo.Sub),
		zap.String("email", userInfo.Email),
		zap.Strings("roles", userInfo.Roles),
		zap.String("session_id", session.ID),
	)

	return c.JSON(fiber.Map{
		"message":    "Giriş başarılı",
		"user_info":  userInfo,
		"session_id": session.ID,
		"expires_in": 24 * 60 * 60, // 24 saat (saniye)
		"trace_id":   traceID,
	})
}

// Logout - Çıkış yap
// @Summary Logout
// @Description Kullanıcı oturumunu sonlandır
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/logout [post]
func Logout(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	// Session ID'yi cookie'den al
	sessionID := c.Cookies("session_id")
	if sessionID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":    "Geçersiz oturum",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Logout endpoint çağrıldı",
		zap.String("trace_id", traceID),
		zap.String("session_id", sessionID),
	)

	if authService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Auth service yapılandırılmamış",
			"trace_id": traceID,
		})
	}

	// Session'ı sil
	if err := authService.DeleteSession(sessionID); err != nil {
		zapLogger.Warn("Session silinemedi",
			zap.String("trace_id", traceID),
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
	}

	// Cookie'yi temizle
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
	})

	zapLogger.Info("User başarıyla çıkış yaptı",
		zap.String("trace_id", traceID),
		zap.String("session_id", sessionID),
	)

	return c.JSON(fiber.Map{
		"message":  "Çıkış başarılı",
		"trace_id": traceID,
	})
}

// Profile - Kullanıcı profili
// @Summary User Profile
// @Description Oturum açmış kullanıcının profil bilgileri
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/profile [get]
func Profile(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	// Session ID'yi cookie'den al
	sessionID := c.Cookies("session_id")
	if sessionID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":    "Geçersiz oturum",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Profile endpoint çağrıldı",
		zap.String("trace_id", traceID),
		zap.String("session_id", sessionID),
	)

	if authService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Auth service yapılandırılmamış",
			"trace_id": traceID,
		})
	}

	// Session'ı validate et
	session, err := authService.ValidateSession(sessionID)
	if err != nil {
		zapLogger.Warn("Session validation başarısız",
			zap.String("trace_id", traceID),
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":    "Geçersiz veya süresi dolmuş oturum",
			"trace_id": traceID,
		})
	}

	profile := fiber.Map{
		"user_id":       session.UserID,
		"name":          session.Name,
		"email":         session.Email,
		"roles":         session.Roles,
		"login_time":    session.LoginTime,
		"last_activity": session.LastActivity,
		"session_id":    session.ID,
		"trace_id":      traceID,
	}

	return c.JSON(profile)
}

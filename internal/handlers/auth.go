package handlers

import (
	"context"
	"fiber-app/internal/services"
	"fiber-app/pkg/cache"
	"time"

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
// @Description Zitadel OAuth2 login işlemini başlatır
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

	// OAuth2 authorization URL oluştur
	authURL, state, err := authService.GenerateAuthURL()
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

	// State'i cache'e kaydet (CSRF koruması için)
	if err := cache.Set("auth_state:"+state, traceID, 10*time.Minute); err != nil {
		zapLogger.Warn("State cache'e kaydedilemedi",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
	}

	zapLogger.Info("Auth URL oluşturuldu",
		zap.String("trace_id", traceID),
		zap.String("state", state),
	)

	return c.JSON(fiber.Map{
		"auth_url": authURL,
		"state":    state,
		"message":  "Bu URL'ye yönlendirilerek giriş yapabilirsiniz",
		"trace_id": traceID,
	})
}

// LoginRedirect - OAuth2 login'e yönlendir
// @Summary OAuth2 Login Redirect
// @Description Zitadel OAuth2 login sayfasına yönlendirir
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

	authURL, state, err := authService.GenerateAuthURL()
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

	// State'i cache'e kaydet
	if err := cache.Set("auth_state:"+state, traceID, 10*time.Minute); err != nil {
		zapLogger.Warn("State cache'e kaydedilemedi",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
	}

	return c.Redirect(authURL)
}

// Callback - OAuth2 callback
// @Summary OAuth2 Callback
// @Description OAuth2 callback endpoint'i
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

	// State'i validate et (CSRF koruması)
	var cachedTraceID string
	if err := cache.Get("auth_state:"+state, &cachedTraceID); err != nil {
		zapLogger.Warn("State validation başarısız",
			zap.String("trace_id", traceID),
			zap.String("state", state),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Geçersiz state parameter",
			"trace_id": traceID,
		})
	}

	// State'i cache'den sil
	cache.Delete("auth_state:" + state)

	if authService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Auth service yapılandırılmamış",
			"trace_id": traceID,
		})
	}

	ctx := context.Background()

	// Authorization code'u token ile değiştir
	token, err := authService.ExchangeCodeForToken(ctx, code)
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

	// JWT token oluştur
	jwtToken, err := authService.CreateJWTToken(userInfo)
	if err != nil {
		zapLogger.Error("JWT token oluşturulamadı",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "JWT token oluşturulamadı",
			"trace_id": traceID,
		})
	}

	// User session'ını cache'e kaydet
	sessionKey := "session:" + userInfo.Sub
	sessionData := map[string]interface{}{
		"user_id":    userInfo.Sub,
		"name":       userInfo.Name,
		"email":      userInfo.Email,
		"roles":      userInfo.Roles,
		"login_time": time.Now(),
	}

	if err := cache.Set(sessionKey, sessionData, 24*time.Hour); err != nil {
		zapLogger.Warn("Session cache'e kaydedilemedi",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
	}

	zapLogger.Info("User başarıyla giriş yaptı",
		zap.String("trace_id", traceID),
		zap.String("user_id", userInfo.Sub),
		zap.String("email", userInfo.Email),
		zap.Strings("roles", userInfo.Roles),
	)

	return c.JSON(fiber.Map{
		"message":    "Giriş başarılı",
		"token":      jwtToken,
		"user_info":  userInfo,
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
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/logout [post]
func Logout(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	// User ID'yi context'ten al
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":    "Geçersiz oturum",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Logout endpoint çağrıldı",
		zap.String("trace_id", traceID),
		zap.String("user_id", userID),
	)

	// Session'ı cache'den sil
	sessionKey := "session:" + userID
	if err := cache.Delete(sessionKey); err != nil {
		zapLogger.Warn("Session cache'den silinemedi",
			zap.String("trace_id", traceID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}

	zapLogger.Info("User başarıyla çıkış yaptı",
		zap.String("trace_id", traceID),
		zap.String("user_id", userID),
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
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/profile [get]
func Profile(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	// User bilgilerini context'ten al
	userID, _ := c.Locals("user_id").(string)
	userName, _ := c.Locals("user_name").(string)
	userEmail, _ := c.Locals("user_email").(string)
	userRoles, _ := c.Locals("user_roles").([]string)

	zapLogger.Info("Profile endpoint çağrıldı",
		zap.String("trace_id", traceID),
		zap.String("user_id", userID),
	)

	// Session bilgilerini cache'den al
	sessionKey := "session:" + userID
	var sessionData map[string]interface{}
	if err := cache.Get(sessionKey, &sessionData); err != nil {
		zapLogger.Warn("Session cache'den alınamadı",
			zap.String("trace_id", traceID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}

	profile := fiber.Map{
		"user_id":  userID,
		"name":     userName,
		"email":    userEmail,
		"roles":    userRoles,
		"session":  sessionData,
		"trace_id": traceID,
	}

	return c.JSON(profile)
}

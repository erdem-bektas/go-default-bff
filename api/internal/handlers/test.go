package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// TestGet - GET test endpoint
// @Summary GET test endpoint
// @Description Test amaçlı GET endpoint
// @Tags Test
// @Accept json
// @Produce json
// @Param name query string false "İsim parametresi" default("anonymous")
// @Param delay query int false "Gecikme (ms)" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/test [get]
func TestGet(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Test GET endpoint çağrıldı",
		zap.String("trace_id", traceID),
		zap.String("query", c.OriginalURL()),
	)

	// Query parametrelerini al
	name := c.Query("name", "anonymous")
	delay := c.QueryInt("delay", 0)

	// Eğer delay parametresi varsa bekle
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	return c.JSON(fiber.Map{
		"message":   "Test GET başarılı",
		"name":      name,
		"delay_ms":  delay,
		"timestamp": time.Now().UTC(),
		"trace_id":  traceID,
		"headers":   c.GetReqHeaders(),
	})
}

// TestPost - POST test endpoint
// @Summary POST test endpoint
// @Description Test amaçlı POST endpoint
// @Tags Test
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Test verisi"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/test [post]
func TestPost(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		zapLogger.Error("Body parse hatası",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid JSON body",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Test POST endpoint çağrıldı",
		zap.String("trace_id", traceID),
		zap.Any("body", body),
	)

	return c.JSON(fiber.Map{
		"message":      "Test POST başarılı",
		"received":     body,
		"timestamp":    time.Now().UTC(),
		"trace_id":     traceID,
		"content_type": c.Get("Content-Type"),
	})
}

// TestError - Hata test endpoint
// @Summary Hata test endpoint
// @Description Test amaçlı hata endpoint
// @Tags Test
// @Accept json
// @Produce json
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/test/error [get]
func TestError(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Warn("Test error endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	// Intentional error for testing
	return errors.New("bu bir test hatasıdır")
}

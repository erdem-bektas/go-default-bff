package handlers

import (
	"fiber-app/internal/services"
	"fiber-app/pkg/cache"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var cacheService *services.CacheService

// SetCacheService - Cache service'i set eder
func SetCacheService(cs *services.CacheService) {
	cacheService = cs
}

// GetCacheStats - Cache istatistikleri
// @Summary Cache istatistikleri
// @Description Redis cache istatistikleri ve bilgileri
// @Tags Cache
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cache/stats [get]
func GetCacheStats(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Cache stats endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	// Cache service stats
	stats, err := cacheService.GetCacheStats()
	if err != nil {
		zapLogger.Error("Cache stats alınamadı",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Cache stats alınamadı",
			"trace_id": traceID,
		})
	}

	// Redis info
	info, err := cache.Info()
	if err != nil {
		zapLogger.Error("Redis info alınamadı",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
	}

	return c.JSON(fiber.Map{
		"cache_stats": stats,
		"redis_info":  info,
		"trace_id":    traceID,
	})
}

// FlushCache - Cache'i temizle
// @Summary Cache temizle
// @Description Tüm cache'i temizle
// @Tags Cache
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cache/flush [post]
func FlushCache(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Cache flush endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	err := cache.FlushDB()
	if err != nil {
		zapLogger.Error("Cache flush başarısız",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Cache flush başarısız",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Cache başarıyla temizlendi",
		zap.String("trace_id", traceID),
	)

	return c.JSON(fiber.Map{
		"message":  "Cache başarıyla temizlendi",
		"trace_id": traceID,
	})
}

// GetCacheKeys - Cache key'lerini listele
// @Summary Cache key'leri
// @Description Pattern ile cache key'lerini listele
// @Tags Cache
// @Accept json
// @Produce json
// @Param pattern query string false "Key pattern" default("*")
// @Param limit query int false "Limit" default(100)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cache/keys [get]
func GetCacheKeys(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	pattern := c.Query("pattern", "*")
	limit, _ := strconv.Atoi(c.Query("limit", "100"))

	zapLogger.Info("Cache keys endpoint çağrıldı",
		zap.String("trace_id", traceID),
		zap.String("pattern", pattern),
		zap.Int("limit", limit),
	)

	keys, err := cache.Keys(pattern)
	if err != nil {
		zapLogger.Error("Cache keys alınamadı",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Cache keys alınamadı",
			"trace_id": traceID,
		})
	}

	// Limit uygula
	if len(keys) > limit {
		keys = keys[:limit]
	}

	return c.JSON(fiber.Map{
		"keys":     keys,
		"count":    len(keys),
		"pattern":  pattern,
		"trace_id": traceID,
	})
}

// DeleteCacheKey - Cache key'ini sil
// @Summary Cache key sil
// @Description Belirtilen cache key'ini sil
// @Tags Cache
// @Accept json
// @Produce json
// @Param key path string true "Cache key"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cache/keys/{key} [delete]
func DeleteCacheKey(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Key parametresi gerekli",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Cache key siliniyor",
		zap.String("trace_id", traceID),
		zap.String("key", key),
	)

	err := cache.Delete(key)
	if err != nil {
		zapLogger.Error("Cache key silinemedi",
			zap.String("trace_id", traceID),
			zap.String("key", key),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Cache key silinemedi",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Cache key başarıyla silindi",
		zap.String("trace_id", traceID),
		zap.String("key", key),
	)

	return c.JSON(fiber.Map{
		"message":  "Cache key başarıyla silindi",
		"key":      key,
		"trace_id": traceID,
	})
}

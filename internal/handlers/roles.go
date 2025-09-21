package handlers

import (
	"errors"
	"fiber-app/internal/models"
	"fiber-app/pkg/database"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetRoles - Tüm rolleri listele
// @Summary Rolleri listele
// @Description Sayfalama desteği ile rolleri listele
// @Tags Roles
// @Accept json
// @Produce json
// @Param page query int false "Sayfa numarası" default(1)
// @Param limit query int false "Sayfa başına kayıt sayısı" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/roles [get]
func GetRoles(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	// Query parametreleri
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	zapLogger.Info("Roles listesi istendi",
		zap.String("trace_id", traceID),
		zap.Int("page", page),
		zap.Int("limit", limit),
	)

	var roles []models.Role
	var total int64

	// Toplam sayı
	if err := database.DB.Model(&models.Role{}).Count(&total).Error; err != nil {
		zapLogger.Error("Roles count hatası",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Sayfalama ile veri çek
	if err := database.DB.Offset(offset).Limit(limit).Order("created_at DESC").Find(&roles).Error; err != nil {
		zapLogger.Error("Roles listesi hatası",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	return c.JSON(fiber.Map{
		"roles": roles,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
		"trace_id": traceID,
	})
}

// GetRole - Tek rol getir
// @Summary Rol detayı
// @Description ID ile rol detayını getir
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/roles/{id} [get]
func GetRole(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	roleID := c.Params("id")
	if roleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Role ID gerekli",
			"trace_id": traceID,
		})
	}

	// UUID kontrolü
	id, err := uuid.Parse(roleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Geçersiz Role ID formatı",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Role detayı istendi",
		zap.String("trace_id", traceID),
		zap.String("role_id", roleID),
	)

	var role models.Role
	if err := database.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "Role bulunamadı",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("Role getirme hatası",
			zap.String("trace_id", traceID),
			zap.String("role_id", roleID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	return c.JSON(fiber.Map{
		"role":     role,
		"trace_id": traceID,
	})
}

// CreateRole - Yeni rol oluştur
// @Summary Yeni rol oluştur
// @Description Yeni rol kaydı oluştur
// @Tags Roles
// @Accept json
// @Produce json
// @Param role body models.CreateRoleRequest true "Role bilgileri"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/roles [post]
func CreateRole(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	var req models.CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		zapLogger.Error("Role create body parse hatası",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Geçersiz JSON formatı",
			"trace_id": traceID,
		})
	}

	// Basit validasyon
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Name alanı gerekli",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Yeni role oluşturuluyor",
		zap.String("trace_id", traceID),
		zap.String("name", req.Name),
	)

	role := models.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := database.DB.Create(&role).Error; err != nil {
		zapLogger.Error("Role oluşturma hatası",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)

		// Name unique constraint hatası
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "name") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":    "Bu role adı zaten kullanımda",
				"trace_id": traceID,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Role başarıyla oluşturuldu",
		zap.String("trace_id", traceID),
		zap.String("role_id", role.ID.String()),
	)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Role başarıyla oluşturuldu",
		"role":     role,
		"trace_id": traceID,
	})
}

// UpdateRole - Rol güncelle
// @Summary Rol güncelle
// @Description Mevcut rol bilgilerini güncelle
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Param role body models.UpdateRoleRequest true "Güncellenecek role bilgileri"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/roles/{id} [put]
func UpdateRole(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	roleID := c.Params("id")
	if roleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Role ID gerekli",
			"trace_id": traceID,
		})
	}

	// UUID kontrolü
	id, err := uuid.Parse(roleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Geçersiz Role ID formatı",
			"trace_id": traceID,
		})
	}

	var req models.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		zapLogger.Error("Role update body parse hatası",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Geçersiz JSON formatı",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Role güncelleniyor",
		zap.String("trace_id", traceID),
		zap.String("role_id", roleID),
	)

	// Önce role'ün var olup olmadığını kontrol et
	var role models.Role
	if err := database.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "Role bulunamadı",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("Role bulma hatası",
			zap.String("trace_id", traceID),
			zap.String("role_id", roleID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Güncelleme verilerini hazırla
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Güncellenecek alan bulunamadı",
			"trace_id": traceID,
		})
	}

	// Güncelle
	if err := database.DB.Model(&role).Updates(updates).Error; err != nil {
		zapLogger.Error("Role güncelleme hatası",
			zap.String("trace_id", traceID),
			zap.String("role_id", roleID),
			zap.Error(err),
		)

		// Name unique constraint hatası
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "name") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":    "Bu role adı zaten kullanımda",
				"trace_id": traceID,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Güncellenmiş role'ü getir
	if err := database.DB.First(&role, "id = ?", id).Error; err != nil {
		zapLogger.Error("Güncellenmiş role getirme hatası",
			zap.String("trace_id", traceID),
			zap.String("role_id", roleID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Role başarıyla güncellendi",
		zap.String("trace_id", traceID),
		zap.String("role_id", roleID),
	)

	return c.JSON(fiber.Map{
		"message":  "Role başarıyla güncellendi",
		"role":     role,
		"trace_id": traceID,
	})
}

// DeleteRole - Rol sil
// @Summary Rol sil
// @Description Rolü sistemden sil (kullanımda değilse)
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/roles/{id} [delete]
func DeleteRole(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	roleID := c.Params("id")
	if roleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Role ID gerekli",
			"trace_id": traceID,
		})
	}

	// UUID kontrolü
	id, err := uuid.Parse(roleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Geçersiz Role ID formatı",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Role siliniyor",
		zap.String("trace_id", traceID),
		zap.String("role_id", roleID),
	)

	// Önce role'ün var olup olmadığını kontrol et
	var role models.Role
	if err := database.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "Role bulunamadı",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("Role bulma hatası",
			zap.String("trace_id", traceID),
			zap.String("role_id", roleID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Bu role'ü kullanan user var mı kontrol et
	var userCount int64
	if err := database.DB.Model(&models.User{}).Where("role_id = ?", id).Count(&userCount).Error; err != nil {
		zapLogger.Error("User count kontrol hatası",
			zap.String("trace_id", traceID),
			zap.String("role_id", roleID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	if userCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":    "Bu role'ü kullanan kullanıcılar var, silinemez",
			"trace_id": traceID,
		})
	}

	// Sil
	if err := database.DB.Delete(&role).Error; err != nil {
		zapLogger.Error("Role silme hatası",
			zap.String("trace_id", traceID),
			zap.String("role_id", roleID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Role başarıyla silindi",
		zap.String("trace_id", traceID),
		zap.String("role_id", roleID),
	)

	return c.JSON(fiber.Map{
		"message":  "Role başarıyla silindi",
		"trace_id": traceID,
	})
}

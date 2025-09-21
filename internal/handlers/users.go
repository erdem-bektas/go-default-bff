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

// GetUsers - Tüm kullanıcıları listele
// @Summary Kullanıcıları listele
// @Description Sayfalama ve arama desteği ile kullanıcıları listele
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Sayfa numarası" default(1)
// @Param limit query int false "Sayfa başına kayıt sayısı" default(10)
// @Param search query string false "Arama terimi (isim veya email)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users [get]
func GetUsers(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	// Query parametreleri
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	zapLogger.Info("Users listesi istendi",
		zap.String("trace_id", traceID),
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("search", search),
	)

	var users []models.User
	var total int64

	query := database.DB.Model(&models.User{}).Preload("Role")

	// Arama filtresi
	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Toplam sayı
	if err := query.Count(&total).Error; err != nil {
		zapLogger.Error("Users count hatası",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Sayfalama ile veri çek
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		zapLogger.Error("Users listesi hatası",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
		"trace_id": traceID,
	})
}

// GetUser - Tek kullanıcı getir
// @Summary Kullanıcı detayı
// @Description ID ile kullanıcı detayını getir
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/{id} [get]
func GetUser(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "User ID gerekli",
			"trace_id": traceID,
		})
	}

	// UUID kontrolü
	id, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Geçersiz User ID formatı",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("User detayı istendi",
		zap.String("trace_id", traceID),
		zap.String("user_id", userID),
	)

	// Önce cache'den kontrol et
	if cacheService != nil {
		if cachedUser, err := cacheService.GetUser(id); err == nil {
			zapLogger.Info("User cache'den getirildi",
				zap.String("trace_id", traceID),
				zap.String("user_id", userID),
			)
			return c.JSON(fiber.Map{
				"user":     cachedUser,
				"trace_id": traceID,
				"cached":   true,
			})
		}
	}

	// Cache'de yoksa database'den getir
	var user models.User
	if err := database.DB.Preload("Role").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "User bulunamadı",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("User getirme hatası",
			zap.String("trace_id", traceID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Cache'e kaydet
	if cacheService != nil {
		if err := cacheService.SetUser(&user); err != nil {
			zapLogger.Warn("User cache'e kaydedilemedi",
				zap.String("trace_id", traceID),
				zap.String("user_id", userID),
				zap.Error(err),
			)
		}
	}

	return c.JSON(fiber.Map{
		"user":     user,
		"trace_id": traceID,
		"cached":   false,
	})
}

// CreateUser - Yeni kullanıcı oluştur
// @Summary Yeni kullanıcı oluştur
// @Description Yeni kullanıcı kaydı oluştur
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User bilgileri"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users [post]
func CreateUser(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		zapLogger.Error("User create body parse hatası",
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

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Email alanı gerekli",
			"trace_id": traceID,
		})
	}

	// Role kontrolü
	var role models.Role
	if err := database.DB.First(&role, "id = ?", req.RoleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":    "Geçersiz role ID",
				"trace_id": traceID,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("Yeni user oluşturuluyor",
		zap.String("trace_id", traceID),
		zap.String("name", req.Name),
		zap.String("email", req.Email),
		zap.String("role", role.Name),
	)

	user := models.User{
		Name:   req.Name,
		Email:  req.Email,
		Age:    req.Age,
		Active: true,
		RoleID: req.RoleID,
	}

	if req.Active != nil {
		user.Active = *req.Active
	}

	if err := database.DB.Create(&user).Error; err != nil {
		zapLogger.Error("User oluşturma hatası",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)

		// Email unique constraint hatası
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "email") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":    "Bu email adresi zaten kullanımda",
				"trace_id": traceID,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Role bilgisini yükle
	database.DB.Preload("Role").First(&user, user.ID)

	zapLogger.Info("User başarıyla oluşturuldu",
		zap.String("trace_id", traceID),
		zap.String("user_id", user.ID.String()),
	)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "User başarıyla oluşturuldu",
		"user":     user,
		"trace_id": traceID,
	})
}

// UpdateUser - Kullanıcı güncelle
// @Summary Kullanıcı güncelle
// @Description Mevcut kullanıcı bilgilerini güncelle
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param user body models.UpdateUserRequest true "Güncellenecek user bilgileri"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/{id} [put]
func UpdateUser(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "User ID gerekli",
			"trace_id": traceID,
		})
	}

	// UUID kontrolü
	id, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Geçersiz User ID formatı",
			"trace_id": traceID,
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		zapLogger.Error("User update body parse hatası",
			zap.String("trace_id", traceID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Geçersiz JSON formatı",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("User güncelleniyor",
		zap.String("trace_id", traceID),
		zap.String("user_id", userID),
	)

	// Önce user'ın var olup olmadığını kontrol et
	var user models.User
	if err := database.DB.Preload("Role").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "User bulunamadı",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("User bulma hatası",
			zap.String("trace_id", traceID),
			zap.String("user_id", userID),
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
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Age != nil {
		updates["age"] = *req.Age
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}
	if req.RoleID != nil {
		// Role kontrolü
		var role models.Role
		if err := database.DB.First(&role, "id = ?", *req.RoleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":    "Geçersiz role ID",
					"trace_id": traceID,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":    "Database hatası",
				"trace_id": traceID,
			})
		}
		updates["role_id"] = *req.RoleID
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Güncellenecek alan bulunamadı",
			"trace_id": traceID,
		})
	}

	// Güncelle
	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		zapLogger.Error("User güncelleme hatası",
			zap.String("trace_id", traceID),
			zap.String("user_id", userID),
			zap.Error(err),
		)

		// Email unique constraint hatası
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "email") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":    "Bu email adresi zaten kullanımda",
				"trace_id": traceID,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Güncellenmiş user'ı getir
	if err := database.DB.Preload("Role").First(&user, "id = ?", id).Error; err != nil {
		zapLogger.Error("Güncellenmiş user getirme hatası",
			zap.String("trace_id", traceID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Cache'i invalidate et
	if cacheService != nil {
		if err := cacheService.InvalidateUserCaches(id); err != nil {
			zapLogger.Warn("User cache invalidation başarısız",
				zap.String("trace_id", traceID),
				zap.String("user_id", userID),
				zap.Error(err),
			)
		}
	}

	zapLogger.Info("User başarıyla güncellendi",
		zap.String("trace_id", traceID),
		zap.String("user_id", userID),
	)

	return c.JSON(fiber.Map{
		"message":  "User başarıyla güncellendi",
		"user":     user,
		"trace_id": traceID,
	})
}

// DeleteUser - Kullanıcı sil
// @Summary Kullanıcı sil
// @Description Kullanıcıyı sistemden sil
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/{id} [delete]
func DeleteUser(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "User ID gerekli",
			"trace_id": traceID,
		})
	}

	// UUID kontrolü
	id, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Geçersiz User ID formatı",
			"trace_id": traceID,
		})
	}

	zapLogger.Info("User siliniyor",
		zap.String("trace_id", traceID),
		zap.String("user_id", userID),
	)

	// Önce user'ın var olup olmadığını kontrol et
	var user models.User
	if err := database.DB.Preload("Role").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":    "User bulunamadı",
				"trace_id": traceID,
			})
		}

		zapLogger.Error("User bulma hatası",
			zap.String("trace_id", traceID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Sil
	if err := database.DB.Delete(&user).Error; err != nil {
		zapLogger.Error("User silme hatası",
			zap.String("trace_id", traceID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Database hatası",
			"trace_id": traceID,
		})
	}

	// Cache'i invalidate et
	if cacheService != nil {
		if err := cacheService.InvalidateUserCaches(id); err != nil {
			zapLogger.Warn("User cache invalidation başarısız",
				zap.String("trace_id", traceID),
				zap.String("user_id", userID),
				zap.Error(err),
			)
		}
	}

	zapLogger.Info("User başarıyla silindi",
		zap.String("trace_id", traceID),
		zap.String("user_id", userID),
	)

	return c.JSON(fiber.Map{
		"message":  "User başarıyla silindi",
		"trace_id": traceID,
	})
}

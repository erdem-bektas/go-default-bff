package cache

import (
	"context"
	"encoding/json"
	"fiber-app/pkg/config"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	RedisClient *redis.Client
	ctx         = context.Background()
)

// Connect - Redis bağlantısı kur
func Connect(cfg *config.Config, zapLogger *zap.Logger) error {
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	zapLogger.Info("Redis bağlantısı kuruluyor",
		zap.String("host", cfg.Redis.Host),
		zap.String("port", cfg.Redis.Port),
		zap.Int("db", cfg.Redis.DB),
	)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Bağlantıyı test et
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		zapLogger.Error("Redis bağlantısı başarısız", zap.Error(err))
		return err
	}

	zapLogger.Info("Redis bağlantısı başarılı")
	return nil
}

// Set - Key-value çifti kaydet (TTL ile)
func Set(key string, value interface{}, ttl time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return RedisClient.Set(ctx, key, jsonValue, ttl).Err()
}

// Get - Key ile value al
func Get(key string, dest interface{}) error {
	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Delete - Key'i sil
func Delete(key string) error {
	return RedisClient.Del(ctx, key).Err()
}

// DeletePattern - Pattern'e uyan key'leri sil
func DeletePattern(pattern string) error {
	keys, err := RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return RedisClient.Del(ctx, keys...).Err()
	}

	return nil
}

// Exists - Key var mı kontrol et
func Exists(key string) bool {
	result, err := RedisClient.Exists(ctx, key).Result()
	return err == nil && result > 0
}

// DBSize - Database'deki key sayısı
func DBSize() (int64, error) {
	return RedisClient.DBSize(ctx).Result()
}

// FlushAll - Tüm key'leri sil
func FlushAll() error {
	return RedisClient.FlushAll(ctx).Err()
}

// FlushDB - Mevcut database'i temizle
func FlushDB() error {
	return RedisClient.FlushDB(ctx).Err()
}

// Keys - Pattern'e uyan key'leri listele
func Keys(pattern string) ([]string, error) {
	return RedisClient.Keys(ctx, pattern).Result()
}

// TTL - Key'in kalan yaşam süresi
func TTL(key string) (time.Duration, error) {
	return RedisClient.TTL(ctx, key).Result()
}

// Expire - Key'e TTL set et
func Expire(key string, ttl time.Duration) error {
	return RedisClient.Expire(ctx, key, ttl).Err()
}

// Info - Redis server bilgileri
func Info() (string, error) {
	return RedisClient.Info(ctx).Result()
}

package handlers

import (
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// GetMetrics - Uygulama metrikleri
// @Summary Uygulama metrikleri
// @Description Temel uygulama performans metrikleri
// @Tags Metrics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/metrics [get]
func GetMetrics(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("Metrics endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime)

	metrics := fiber.Map{
		"timestamp":      time.Now().UTC(),
		"uptime":         uptime.String(),
		"uptime_seconds": int64(uptime.Seconds()),
		"memory": fiber.Map{
			"alloc_mb":       m.Alloc / 1024 / 1024,
			"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
			"sys_mb":         m.Sys / 1024 / 1024,
			"heap_alloc_mb":  m.HeapAlloc / 1024 / 1024,
			"heap_sys_mb":    m.HeapSys / 1024 / 1024,
		},
		"gc": fiber.Map{
			"num_gc":         m.NumGC,
			"pause_total_ns": m.PauseTotalNs,
			"last_gc":        time.Unix(0, int64(m.LastGC)).UTC(),
		},
		"goroutines": runtime.NumGoroutine(),
		"cpu_count":  runtime.NumCPU(),
		"trace_id":   traceID,
	}

	return c.JSON(metrics)
}

// GetSystemMetrics - Sistem metrikleri
// @Summary Detaylı sistem metrikleri
// @Description Detaylı sistem ve runtime metrikleri
// @Tags Metrics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/metrics/system [get]
func GetSystemMetrics(c *fiber.Ctx) error {
	traceID := getTraceID(c)

	zapLogger.Info("System metrics endpoint çağrıldı",
		zap.String("trace_id", traceID),
	)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemMetrics := fiber.Map{
		"timestamp": time.Now().UTC(),
		"runtime": fiber.Map{
			"version":    runtime.Version(),
			"goos":       runtime.GOOS,
			"goarch":     runtime.GOARCH,
			"compiler":   runtime.Compiler,
			"num_cpu":    runtime.NumCPU(),
			"goroutines": runtime.NumGoroutine(),
			"cgo_calls":  runtime.NumCgoCall(),
		},
		"memory_detailed": fiber.Map{
			"alloc":         m.Alloc,
			"total_alloc":   m.TotalAlloc,
			"sys":           m.Sys,
			"lookups":       m.Lookups,
			"mallocs":       m.Mallocs,
			"frees":         m.Frees,
			"heap_alloc":    m.HeapAlloc,
			"heap_sys":      m.HeapSys,
			"heap_idle":     m.HeapIdle,
			"heap_inuse":    m.HeapInuse,
			"heap_released": m.HeapReleased,
			"heap_objects":  m.HeapObjects,
			"stack_inuse":   m.StackInuse,
			"stack_sys":     m.StackSys,
			"mspan_inuse":   m.MSpanInuse,
			"mspan_sys":     m.MSpanSys,
			"mcache_inuse":  m.MCacheInuse,
			"mcache_sys":    m.MCacheSys,
			"buck_hash_sys": m.BuckHashSys,
			"gc_sys":        m.GCSys,
			"other_sys":     m.OtherSys,
		},
		"gc_detailed": fiber.Map{
			"num_gc":          m.NumGC,
			"num_forced_gc":   m.NumForcedGC,
			"gc_cpu_fraction": m.GCCPUFraction,
			"pause_total_ns":  m.PauseTotalNs,
			"pause_ns":        m.PauseNs,
			"pause_end":       m.PauseEnd,
			"last_gc":         time.Unix(0, int64(m.LastGC)).UTC(),
			"next_gc":         m.NextGC,
		},
		"trace_id": traceID,
	}

	return c.JSON(systemMetrics)
}

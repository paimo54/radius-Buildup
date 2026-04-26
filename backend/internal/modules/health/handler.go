package health

import (
	"net/http"
	"runtime"
	"time"

	"radius-buildup/internal/database"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// RegisterRoutes registers health check routes
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/health", HealthCheck)
}

// HealthCheck returns the current server and database status
func HealthCheck(c *gin.Context) {
	// Check database
	dbStatus := "connected"
	dbError := ""

	sqlDB, err := database.DB.DB()
	if err != nil {
		dbStatus = "error"
		dbError = err.Error()
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = "disconnected"
		dbError = err.Error()
	}

	// Count some tables to verify DB access
	var userCount int64
	database.DB.Table("pppoe_users").Count(&userCount)

	var nasCount int64
	database.DB.Table("nas").Count(&nasCount)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Radius Buildup Go API is running",
		"server": gin.H{
			"uptime":     time.Since(startTime).String(),
			"goVersion":  runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
			"memoryMB":   memStats.Alloc / 1024 / 1024,
		},
		"database": gin.H{
			"status":    dbStatus,
			"error":     dbError,
			"pppoeUsers": userCount,
			"routers":   nasCount,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

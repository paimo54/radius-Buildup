package mikrotik

import (
	"net/http"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers MikroTik direct interaction routes
func RegisterRoutes(r *gin.RouterGroup) {
	mk := r.Group("/mikrotik")
	{
		mk.GET("/:id/ping", Ping)
		mk.GET("/:id/resource", GetResource)
		mk.POST("/:id/kick-pppoe", KickPppoe)
		mk.POST("/:id/kick-hotspot", KickHotspot)
	}
}

// Ping checks router connectivity
func Ping(c *gin.Context) {
	id := c.Param("id")
	
	var router models.Router
	if err := database.DB.First(&router, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router tidak ditemukan"})
		return
	}

	success, msg := PingRouter(&router)
	if !success {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "offline",
			"error":  msg,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "online",
		"message": "Koneksi berhasil",
	})
}

// GetResource fetches CPU/RAM stats
func GetResource(c *gin.Context) {
	id := c.Param("id")
	
	var router models.Router
	if err := database.DB.First(&router, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router tidak ditemukan"})
		return
	}

	res, err := GetSystemResource(&router)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil resource: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}

type KickRequest struct {
	Username string `json:"username" binding:"required"`
}

// KickPppoe kicks an active PPPoE user from the router
func KickPppoe(c *gin.Context) {
	id := c.Param("id")
	
	var req KickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	var router models.Router
	if err := database.DB.First(&router, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router tidak ditemukan"})
		return
	}

	if err := KickPppoeUser(&router, req.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memutus koneksi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User berhasil diputus koneksinya (kicked)"})
}

// KickHotspot kicks an active Hotspot user from the router
func KickHotspot(c *gin.Context) {
	id := c.Param("id")
	
	var req KickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	var router models.Router
	if err := database.DB.First(&router, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router tidak ditemukan"})
		return
	}

	if err := KickHotspotUser(&router, req.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memutus koneksi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User berhasil diputus koneksinya (kicked)"})
}

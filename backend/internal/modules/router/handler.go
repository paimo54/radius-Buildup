package router

import (
	"net/http"
	"strconv"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers Router/NAS routes
func RegisterRoutes(r *gin.RouterGroup) {
	router := r.Group("/routers")
	{
		router.GET("", ListRouters)
		router.GET("/:id", GetRouter)
		router.POST("", CreateRouter)
		router.PUT("/:id", UpdateRouter)
		router.DELETE("/:id", DeleteRouter)
	}
}

// ListRouters returns all routers
func ListRouters(c *gin.Context) {
	var routers []models.Router
	result := database.DB.Order("createdAt DESC").Find(&routers)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": routers})
}

// GetRouter returns a single router by ID
func GetRouter(c *gin.Context) {
	id := c.Param("id")

	var router models.Router
	if err := database.DB.First(&router, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": router})
}

// CreateRouterRequest is the payload for creating a router
type CreateRouterRequest struct {
	Name      string  `json:"name" binding:"required"`
	Nasname   string  `json:"nasname" binding:"required"`
	Shortname string  `json:"shortname" binding:"required"`
	IPAddress string  `json:"ipAddress" binding:"required"`
	Username  string  `json:"username" binding:"required"`
	Password  string  `json:"password" binding:"required"`
	Secret    string  `json:"secret"`
	Port      int     `json:"port"`
	ApiPort   int     `json:"apiPort"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// CreateRouter creates a new router
func CreateRouter(c *gin.Context) {
	var req CreateRouterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	// Defaults
	if req.Secret == "" {
		req.Secret = "secret123"
	}
	if req.Port == 0 {
		req.Port = 8728
	}
	if req.ApiPort == 0 {
		req.ApiPort = 8729
	}

	router := models.Router{
		ID:        generateID(),
		Name:      req.Name,
		Nasname:   req.Nasname,
		Shortname: req.Shortname,
		IPAddress: req.IPAddress,
		Username:  req.Username,
		Password:  req.Password,
		Secret:    req.Secret,
		Port:      req.Port,
		ApiPort:   req.ApiPort,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		IsActive:  true,
	}

	if err := database.DB.Create(&router).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat router: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Router berhasil ditambahkan",
		"data":    router,
	})
}

// UpdateRouterRequest is the payload for updating a router
type UpdateRouterRequest struct {
	Name      *string  `json:"name"`
	Nasname   *string  `json:"nasname"`
	Shortname *string  `json:"shortname"`
	IPAddress *string  `json:"ipAddress"`
	Username  *string  `json:"username"`
	Password  *string  `json:"password"`
	Secret    *string  `json:"secret"`
	Port      *int     `json:"port"`
	ApiPort   *int     `json:"apiPort"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	IsActive  *bool    `json:"isActive"`
}

// UpdateRouter updates an existing router
func UpdateRouter(c *gin.Context) {
	id := c.Param("id")

	var router models.Router
	if err := database.DB.First(&router, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router tidak ditemukan"})
		return
	}

	var req UpdateRouterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil { updates["name"] = *req.Name }
	if req.Nasname != nil { updates["nasname"] = *req.Nasname }
	if req.Shortname != nil { updates["shortname"] = *req.Shortname }
	if req.IPAddress != nil { updates["ipAddress"] = *req.IPAddress }
	if req.Username != nil { updates["username"] = *req.Username }
	if req.Password != nil { updates["password"] = *req.Password }
	if req.Secret != nil { updates["secret"] = *req.Secret }
	if req.Port != nil { updates["port"] = *req.Port }
	if req.ApiPort != nil { updates["apiPort"] = *req.ApiPort }
	if req.Latitude != nil { updates["latitude"] = *req.Latitude }
	if req.Longitude != nil { updates["longitude"] = *req.Longitude }
	if req.IsActive != nil { updates["isActive"] = *req.IsActive }

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada data yang diubah"})
		return
	}

	if err := database.DB.Model(&router).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update router: " + err.Error()})
		return
	}

	database.DB.First(&router, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{
		"message": "Router berhasil diupdate",
		"data":    router,
	})
}

// DeleteRouter deletes a router
func DeleteRouter(c *gin.Context) {
	id := c.Param("id")

	var router models.Router
	if err := database.DB.First(&router, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router tidak ditemukan"})
		return
	}

	if err := database.DB.Delete(&router).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus router: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Router berhasil dihapus"})
}

func generateID() string {
	now := time.Now()
	return "nas_" + strconv.FormatInt(now.UnixNano(), 36)
}

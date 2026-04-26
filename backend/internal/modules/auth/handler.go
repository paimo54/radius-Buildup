package auth

import (
	"net/http"

	"radius-buildup/internal/config"
	"radius-buildup/internal/database"
	"radius-buildup/internal/middleware"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var cfg *config.Config

// RegisterRoutes registers auth routes
func RegisterRoutes(r *gin.RouterGroup, appCfg *config.Config) {
	cfg = appCfg
	r.POST("/auth/login", Login)
}

// RegisterProtectedRoutes registers routes that need auth
func RegisterProtectedRoutes(r *gin.RouterGroup) {
	r.GET("/auth/me", Me)
}

// LoginRequest is the login payload
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles admin login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan password harus diisi"})
		return
	}

	// Try admin_users table first
	var admin models.AdminUser
	result := database.DB.Where("username = ? AND isActive = ?", req.Username, true).First(&admin)

	if result.Error == nil {
		// Found in admin_users - check password
		if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Password salah"})
			return
		}

		token, err := middleware.GenerateToken(cfg, admin.ID, admin.Username, admin.Name, admin.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
			return
		}

		// Update last login
		database.DB.Model(&admin).Update("lastLogin", database.DB.NowFunc())

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":       admin.ID,
				"username": admin.Username,
				"name":     admin.Name,
				"role":     admin.Role,
				"email":    admin.Email,
			},
		})
		return
	}

	// Fallback: try legacy users table
	var user models.User
	result = database.DB.Where("email = ?", req.Username).First(&user)

	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username tidak ditemukan"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password salah"})
		return
	}

	token, err := middleware.GenerateToken(cfg, user.ID, user.Email, user.Name, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Email,
			"name":     user.Name,
			"role":     user.Role,
		},
	})
}

// Me returns the currently authenticated user info
func Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":       c.GetString("userId"),
		"username": c.GetString("username"),
		"name":     c.GetString("name"),
		"role":     c.GetString("role"),
	})
}

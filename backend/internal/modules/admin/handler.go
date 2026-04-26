package admin

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRoutes registers admin user management routes
func RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin-users")
	{
		admin.GET("", ListAdmins)
		admin.GET("/:id", GetAdmin)
		admin.POST("", CreateAdmin)
		admin.PUT("/:id", UpdateAdmin)
		admin.DELETE("/:id", DeleteAdmin)
	}
}

// ListAdmins returns all admin users
func ListAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	query := database.DB.Model(&models.AdminUser{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("username LIKE ? OR name LIKE ? OR email LIKE ?", like, like, like)
	}

	var total int64
	query.Count(&total)

	var admins []models.AdminUser
	query.Order("createdAt DESC").Offset(offset).Limit(limit).Find(&admins)

	c.JSON(http.StatusOK, gin.H{
		"data": admins,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// GetAdmin returns a single admin
func GetAdmin(c *gin.Context) {
	id := c.Param("id")
	var admin models.AdminUser
	if err := database.DB.First(&admin, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": admin})
}

// CreateAdminRequest payload
type CreateAdminRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

// CreateAdmin creates a new admin user
func CreateAdmin(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	// Check duplicate username
	var exists int64
	database.DB.Model(&models.AdminUser{}).Where("username = ?", req.Username).Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username sudah digunakan"})
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal hash password"})
		return
	}

	role := req.Role
	if role == "" { role = "ADMIN" }

	admin := models.AdminUser{
		ID:       generateID(),
		Username: req.Username,
		Password: string(hashed),
		Name:     req.Name,
		Email:    &req.Email,
		Phone:    &req.Phone,
		Role:     role,
		IsActive: true,
	}

	if err := database.DB.Create(&admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat admin: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Admin berhasil dibuat", "data": admin})
}

// UpdateAdminRequest payload
type UpdateAdminRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Role     *string `json:"role"`
	Password *string `json:"password"`
	IsActive *bool   `json:"isActive"`
}

// UpdateAdmin updates an admin user
func UpdateAdmin(c *gin.Context) {
	id := c.Param("id")
	var admin models.AdminUser
	if err := database.DB.First(&admin, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin tidak ditemukan"})
		return
	}

	var req UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil { updates["name"] = *req.Name }
	if req.Email != nil { updates["email"] = *req.Email }
	if req.Phone != nil { updates["phone"] = *req.Phone }
	if req.Role != nil { updates["role"] = *req.Role }
	if req.IsActive != nil { updates["isActive"] = *req.IsActive }
	if req.Password != nil && *req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err == nil {
			updates["password"] = string(hashed)
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada data yang diubah"})
		return
	}

	database.DB.Model(&admin).Updates(updates)
	database.DB.First(&admin, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{"message": "Admin berhasil diupdate", "data": admin})
}

// DeleteAdmin deletes an admin (soft-check: can't delete yourself)
func DeleteAdmin(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetString("userId")

	if id == currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tidak bisa menghapus akun sendiri"})
		return
	}

	var admin models.AdminUser
	if err := database.DB.First(&admin, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin tidak ditemukan"})
		return
	}

	database.DB.Delete(&admin)
	c.JSON(http.StatusOK, gin.H{"message": "Admin berhasil dihapus"})
}

func generateID() string {
	return "admin_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

package pppoe

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers PPPoE routes (all protected)
func RegisterRoutes(r *gin.RouterGroup) {
	pppoe := r.Group("/pppoe")
	{
		// Users
		pppoe.GET("/users", ListUsers)
		pppoe.GET("/users/:id", GetUser)
		pppoe.POST("/users", CreateUser)
		pppoe.PUT("/users/:id", UpdateUser)
		pppoe.DELETE("/users/:id", DeleteUser)

		// Profiles
		pppoe.GET("/profiles", ListProfiles)

		// Areas
		pppoe.GET("/areas", ListAreas)
	}
}

// ============================================================================
// PPPoE Users
// ============================================================================

// ListUsers returns paginated PPPoE users with optional search/filter
func ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")
	profileID := c.Query("profileId")
	routerID := c.Query("routerId")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := database.DB.Model(&models.PppoeUser{})

	// Filters
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("username LIKE ? OR name LIKE ? OR phone LIKE ?", like, like, like)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if profileID != "" {
		query = query.Where("profileId = ?", profileID)
	}
	if routerID != "" {
		query = query.Where("routerId = ?", routerID)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch with relations
	var users []models.PppoeUser
	result := query.
		Preload("Profile").
		Preload("Area").
		Preload("Router").
		Order("createdAt DESC").
		Offset(offset).
		Limit(limit).
		Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// GetUser returns a single PPPoE user by ID
func GetUser(c *gin.Context) {
	id := c.Param("id")

	var user models.PppoeUser
	result := database.DB.
		Preload("Profile").
		Preload("Area").
		Preload("Router").
		Preload("PppoeCustomer").
		First(&user, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User PPPoE tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// CreateUserRequest is the payload for creating a PPPoE user
type CreateUserRequest struct {
	Username         string  `json:"username" binding:"required"`
	Password         string  `json:"password" binding:"required"`
	Name             string  `json:"name" binding:"required"`
	Phone            string  `json:"phone" binding:"required"`
	ProfileID        string  `json:"profileId" binding:"required"`
	RouterID         *string `json:"routerId"`
	AreaID           *string `json:"areaId"`
	Address          *string `json:"address"`
	Email            *string `json:"email"`
	SubscriptionType string  `json:"subscriptionType"`
	ConnectionType   string  `json:"connectionType"`
}

// CreateUser creates a new PPPoE user
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	// Check username uniqueness
	var count int64
	database.DB.Model(&models.PppoeUser{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username sudah digunakan"})
		return
	}

	// Defaults
	if req.SubscriptionType == "" {
		req.SubscriptionType = "POSTPAID"
	}
	if req.ConnectionType == "" {
		req.ConnectionType = "PPPOE"
	}

	user := models.PppoeUser{
		ID:               generateID(),
		Username:         req.Username,
		Password:         req.Password,
		Name:             req.Name,
		Phone:            req.Phone,
		ProfileID:        req.ProfileID,
		RouterID:         req.RouterID,
		AreaID:           req.AreaID,
		Address:          req.Address,
		Email:            req.Email,
		Status:           "active",
		SubscriptionType: req.SubscriptionType,
		ConnectionType:   req.ConnectionType,
	}

	result := database.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat user: " + result.Error.Error()})
		return
	}

	// RADIUS Sync (Background or Inline)
	// We'll do it inline for now to ensure consistency, though background is also possible.
	if err := SyncToRadius(&user); err != nil {
		LogSyncError(user.Username, err)
	}

	// Reload with relations
	database.DB.Preload("Profile").Preload("Area").Preload("Router").First(&user, "id = ?", user.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "User PPPoE berhasil dibuat",
		"data":    user,
	})
}

// UpdateUserRequest is the payload for updating a PPPoE user
type UpdateUserRequest struct {
	Password         *string `json:"password"`
	Name             *string `json:"name"`
	Phone            *string `json:"phone"`
	ProfileID        *string `json:"profileId"`
	RouterID         *string `json:"routerId"`
	AreaID           *string `json:"areaId"`
	Address          *string `json:"address"`
	Email            *string `json:"email"`
	Status           *string `json:"status"`
	SubscriptionType *string `json:"subscriptionType"`
	IPAddress        *string `json:"ipAddress"`
	MacAddress       *string `json:"macAddress"`
}

// UpdateUser updates an existing PPPoE user
func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var user models.PppoeUser
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User PPPoE tidak ditemukan"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	// Build updates map
	updates := map[string]interface{}{}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.ProfileID != nil {
		updates["profileId"] = *req.ProfileID
	}
	if req.RouterID != nil {
		updates["routerId"] = *req.RouterID
	}
	if req.AreaID != nil {
		updates["areaId"] = *req.AreaID
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.SubscriptionType != nil {
		updates["subscriptionType"] = *req.SubscriptionType
	}
	if req.IPAddress != nil {
		updates["ipAddress"] = *req.IPAddress
	}
	if req.MacAddress != nil {
		updates["macAddress"] = *req.MacAddress
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada data yang diubah"})
		return
	}

	database.DB.Model(&user).Updates(updates)

	// RADIUS Sync if critical fields changed
	// In the real world, we'd check if specific fields changed, 
	// but for simplicity and robustness, we re-sync if the user is active/isolated.
	if err := SyncToRadius(&user); err != nil {
		LogSyncError(user.Username, err)
	}

	// Reload
	database.DB.Preload("Profile").Preload("Area").Preload("Router").First(&user, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{
		"message": "User PPPoE berhasil diupdate",
		"data":    user,
	})
}

// DeleteUser deletes a PPPoE user
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	var user models.PppoeUser
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User PPPoE tidak ditemukan"})
		return
	}

	// RADIUS Cleanup
	if err := CleanupRadius(user.Username); err != nil {
		LogSyncError(user.Username, err)
	}

	database.DB.Delete(&user)

	c.JSON(http.StatusOK, gin.H{"message": "User PPPoE berhasil dihapus"})
}

// ============================================================================
// PPPoE Profiles
// ============================================================================

// ListProfiles returns all active PPPoE profiles
func ListProfiles(c *gin.Context) {
	var profiles []models.PppoeProfile
	query := database.DB.Order("name ASC")

	if c.Query("active") == "true" {
		query = query.Where("isActive = ?", true)
	}

	query.Find(&profiles)

	c.JSON(http.StatusOK, gin.H{"data": profiles})
}

// ============================================================================
// PPPoE Areas
// ============================================================================

// ListAreas returns all active PPPoE areas
func ListAreas(c *gin.Context) {
	var areas []models.PppoeArea
	query := database.DB.Order("name ASC")

	if c.Query("active") == "true" {
		query = query.Where("isActive = ?", true)
	}

	query.Find(&areas)

	c.JSON(http.StatusOK, gin.H{"data": areas})
}

// ============================================================================
// Helpers
// ============================================================================

func generateID() string {
	// Simple CUID-like ID using timestamp
	now := time.Now()
	return "go" + strconv.FormatInt(now.UnixNano(), 36)
}

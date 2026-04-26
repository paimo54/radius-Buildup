package hotspot

import (
	"crypto/rand"
	"encoding/hex"
	"math"
	"net/http"
	"strconv"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers hotspot routes
func RegisterRoutes(r *gin.RouterGroup) {
	// Profiles
	profiles := r.Group("/hotspot/profiles")
	{
		profiles.GET("", ListProfiles)
		profiles.GET("/:id", GetProfile)
		profiles.POST("", CreateProfile)
		profiles.PUT("/:id", UpdateProfile)
		profiles.DELETE("/:id", DeleteProfile)
	}

	// Vouchers
	vouchers := r.Group("/hotspot/vouchers")
	{
		vouchers.GET("", ListVouchers)
		vouchers.POST("/generate", GenerateVouchers)
		vouchers.DELETE("/:id", DeleteVoucher)
		vouchers.DELETE("/batch/:batchCode", DeleteBatch)
	}
}

// ============================================================================
// Profiles
// ============================================================================

func ListProfiles(c *gin.Context) {
	var profiles []models.HotspotProfile
	query := database.DB.Model(&models.HotspotProfile{})

	activeOnly := c.DefaultQuery("active", "")
	if activeOnly == "true" {
		query = query.Where("isActive = ?", true)
	}

	query.Order("createdAt DESC").Find(&profiles)

	// Count vouchers per profile
	type VCount struct {
		ProfileID string
		Total     int64
		Waiting   int64
		Active    int64
		Expired   int64
	}
	var vcounts []VCount
	database.DB.Model(&models.HotspotVoucher{}).
		Select("profileId as profile_id, COUNT(*) as total, "+
			"SUM(CASE WHEN status='WAITING' THEN 1 ELSE 0 END) as waiting, "+
			"SUM(CASE WHEN status='ACTIVE' THEN 1 ELSE 0 END) as active, "+
			"SUM(CASE WHEN status='EXPIRED' THEN 1 ELSE 0 END) as expired").
		Group("profileId").Find(&vcounts)

	vcountMap := make(map[string]VCount)
	for _, v := range vcounts {
		vcountMap[v.ProfileID] = v
	}

	type ProfileDTO struct {
		models.HotspotProfile
		VoucherStats gin.H `json:"voucherStats"`
	}

	result := make([]ProfileDTO, 0, len(profiles))
	for _, p := range profiles {
		vc := vcountMap[p.ID]
		result = append(result, ProfileDTO{
			HotspotProfile: p,
			VoucherStats: gin.H{
				"total":   vc.Total,
				"waiting": vc.Waiting,
				"active":  vc.Active,
				"expired": vc.Expired,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func GetProfile(c *gin.Context) {
	id := c.Param("id")
	var profile models.HotspotProfile
	if err := database.DB.First(&profile, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": profile})
}

type CreateProfileRequest struct {
	Name          string  `json:"name" binding:"required"`
	Speed         string  `json:"speed" binding:"required"`
	CostPrice     int     `json:"costPrice" binding:"required"`
	SellingPrice  int     `json:"sellingPrice" binding:"required"`
	ValidityUnit  string  `json:"validityUnit" binding:"required"`
	ValidityValue int     `json:"validityValue" binding:"required"`
	SharedUsers   int     `json:"sharedUsers"`
	GroupProfile  *string `json:"groupProfile"`
	ResellerFee   int     `json:"resellerFee"`
}

func CreateProfile(c *gin.Context) {
	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	if req.SharedUsers <= 0 {
		req.SharedUsers = 1
	}

	profile := models.HotspotProfile{
		ID:            generateID("hp"),
		Name:          req.Name,
		Speed:         req.Speed,
		CostPrice:     req.CostPrice,
		SellingPrice:  req.SellingPrice,
		ValidityUnit:  req.ValidityUnit,
		ValidityValue: req.ValidityValue,
		SharedUsers:   req.SharedUsers,
		GroupProfile:  req.GroupProfile,
		ResellerFee:   req.ResellerFee,
		IsActive:      true,
		AgentAccess:   true,
		EVoucherAccess: true,
	}

	if err := database.DB.Create(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Profile berhasil dibuat", "data": profile})
}

func UpdateProfile(c *gin.Context) {
	id := c.Param("id")
	var profile models.HotspotProfile
	if err := database.DB.First(&profile, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile tidak ditemukan"})
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	database.DB.Model(&profile).Updates(body)
	database.DB.First(&profile, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{"message": "Profile berhasil diupdate", "data": profile})
}

func DeleteProfile(c *gin.Context) {
	id := c.Param("id")
	var profile models.HotspotProfile
	if err := database.DB.First(&profile, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile tidak ditemukan"})
		return
	}

	// Check if vouchers exist
	var count int64
	database.DB.Model(&models.HotspotVoucher{}).Where("profileId = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Profile masih memiliki voucher, hapus voucher terlebih dahulu"})
		return
	}

	database.DB.Delete(&profile)
	c.JSON(http.StatusOK, gin.H{"message": "Profile berhasil dihapus"})
}

// ============================================================================
// Vouchers
// ============================================================================

func ListVouchers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	profileID := c.Query("profileId")
	status := c.Query("status")
	batchCode := c.Query("batchCode")
	search := c.Query("search")

	if page < 1 { page = 1 }
	if limit < 1 || limit > 200 { limit = 20 }
	offset := (page - 1) * limit

	query := database.DB.Model(&models.HotspotVoucher{})

	if profileID != "" { query = query.Where("profileId = ?", profileID) }
	if status != "" { query = query.Where("status = ?", status) }
	if batchCode != "" { query = query.Where("batchCode = ?", batchCode) }
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("code LIKE ?", like)
	}

	var total int64
	query.Count(&total)

	var vouchers []models.HotspotVoucher
	query.Preload("Profile").Order("createdAt DESC").Offset(offset).Limit(limit).Find(&vouchers)

	c.JSON(http.StatusOK, gin.H{
		"data": vouchers,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

type GenerateVouchersRequest struct {
	ProfileID string `json:"profileId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
	RouterID  string `json:"routerId"`
	CodeType  string `json:"codeType"`  // alphanumeric, numeric
	CodeLength int   `json:"codeLength"`
}

func GenerateVouchers(c *gin.Context) {
	var req GenerateVouchersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	if req.Quantity < 1 || req.Quantity > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity harus 1-500"})
		return
	}

	// Verify profile exists
	var profile models.HotspotProfile
	if err := database.DB.First(&profile, "id = ?", req.ProfileID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile tidak ditemukan"})
		return
	}

	codeLength := req.CodeLength
	if codeLength < 4 || codeLength > 16 { codeLength = 8 }

	batchCode := "batch_" + strconv.FormatInt(time.Now().UnixNano(), 36)

	var routerID *string
	if req.RouterID != "" { routerID = &req.RouterID }

	vouchers := make([]models.HotspotVoucher, 0, req.Quantity)
	for i := 0; i < req.Quantity; i++ {
		code := generateVoucherCode(codeLength, req.CodeType)
		vouchers = append(vouchers, models.HotspotVoucher{
			ID:          generateID("hv"),
			Code:        code,
			Password:    &code, // same as code by default
			ProfileID:   req.ProfileID,
			RouterID:    routerID,
			VoucherType: "same",
			CodeType:    req.CodeType,
			Status:      "WAITING",
			BatchCode:   &batchCode,
		})
	}

	if err := database.DB.CreateInBatches(&vouchers, 100).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate voucher: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Voucher berhasil digenerate",
		"batchCode": batchCode,
		"count":     len(vouchers),
		"data":      vouchers,
	})
}

func DeleteVoucher(c *gin.Context) {
	id := c.Param("id")
	var voucher models.HotspotVoucher
	if err := database.DB.First(&voucher, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Voucher tidak ditemukan"})
		return
	}
	database.DB.Delete(&voucher)
	c.JSON(http.StatusOK, gin.H{"message": "Voucher berhasil dihapus"})
}

func DeleteBatch(c *gin.Context) {
	batchCode := c.Param("batchCode")
	result := database.DB.Where("batchCode = ? AND status = ?", batchCode, "WAITING").Delete(&models.HotspotVoucher{})
	c.JSON(http.StatusOK, gin.H{
		"message": "Batch voucher berhasil dihapus",
		"deleted": result.RowsAffected,
	})
}

// ============================================================================
// Helpers
// ============================================================================

func generateID(prefix string) string {
	return prefix + "_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func generateVoucherCode(length int, codeType string) string {
	if codeType == "numeric" {
		b := make([]byte, length)
		rand.Read(b)
		code := ""
		for _, v := range b {
			code += strconv.Itoa(int(v) % 10)
		}
		return code[:length]
	}
	// alphanumeric
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

package olt

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers OLT endpoints
func RegisterRoutes(r *gin.RouterGroup) {
	olt := r.Group("/olt")
	{
		// Devices
		olt.GET("/devices", ListOlt)
		olt.GET("/devices/:id", GetOlt)
		olt.POST("/devices", CreateOlt)
		olt.PUT("/devices/:id", UpdateOlt)
		olt.DELETE("/devices/:id", DeleteOlt)
		olt.GET("/devices/:id/scan", ScanOlt)
		olt.GET("/devices/:id/onus", ListOnus) // Get cached ONU data from DB
		olt.POST("/devices/:id/reboot-onu", RebootOnuHandler)

		// OID Mappings
		olt.GET("/mappings", ListMappings)
		olt.POST("/mappings", CreateMapping)
		olt.PUT("/mappings/:id", UpdateMapping)
		olt.DELETE("/mappings/:id", DeleteMapping)
	}
}

// ============================================================================
// OLT Devices
// ============================================================================

func ListOlt(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	query := database.DB.Model(&models.Olt{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR host LIKE ? OR vendor LIKE ?", like, like, like)
	}

	var total int64
	query.Count(&total)

	var olts []models.Olt
	query.Order("createdAt DESC").Offset(offset).Limit(limit).Find(&olts)

	c.JSON(http.StatusOK, gin.H{
		"data": olts,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

func GetOlt(c *gin.Context) {
	id := c.Param("id")
	var olt models.Olt
	if err := database.DB.First(&olt, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OLT tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": olt})
}

type CreateOltRequest struct {
	Name         string  `json:"name" binding:"required"`
	Host         string  `json:"host" binding:"required"`
	Port         *int    `json:"port"`
	Community    *string `json:"community"`
	TelnetPort   *int    `json:"telnetPort"`
	TelnetUser   string  `json:"telnetUser"`
	TelnetPass   string  `json:"telnetPass"`
	TelnetEnable bool    `json:"telnetEnable"`
	Vendor       string  `json:"vendor" binding:"required"`
	Description  *string `json:"description"`
}

func CreateOlt(c *gin.Context) {
	var req CreateOltRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	port := 161
	if req.Port != nil { port = *req.Port }

	telnetPort := 23
	if req.TelnetPort != nil { telnetPort = *req.TelnetPort }

	community := "public"
	if req.Community != nil { community = *req.Community }

	olt := models.Olt{
		ID:           "olt_" + strconv.FormatInt(time.Now().UnixNano(), 36),
		Name:         req.Name,
		Host:         req.Host,
		Port:         port,
		Community:    community,
		TelnetPort:   telnetPort,
		TelnetUser:   req.TelnetUser,
		TelnetPass:   req.TelnetPass,
		TelnetEnable: req.TelnetEnable,
		Vendor:       req.Vendor,
		Description:  req.Description,
		Status:       "active",
	}

	if err := database.DB.Create(&olt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan OLT: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "OLT berhasil ditambahkan", "data": olt})
}

func UpdateOlt(c *gin.Context) {
	id := c.Param("id")
	var olt models.Olt
	if err := database.DB.First(&olt, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OLT tidak ditemukan"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	database.DB.Model(&olt).Updates(req)
	database.DB.First(&olt, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{"message": "OLT berhasil diupdate", "data": olt})
}

func DeleteOlt(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.Olt{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus OLT"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OLT berhasil dihapus"})
}

func ScanOlt(c *gin.Context) {
	id := c.Param("id")
	
	onus, err := ScanONU(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal melakukan scan OLT: " + err.Error()})
		return
	}

	// Save results to database
	SaveScanResults(id, onus)

	c.JSON(http.StatusOK, gin.H{
		"message": "Scan OLT berhasil",
		"total":   len(onus),
		"data":    onus,
	})
}

// ListOnus returns cached ONU data from the database (last scan results)
func ListOnus(c *gin.Context) {
	id := c.Param("id")
	
	var onus []models.OltOnu
	database.DB.Where("olt_id = ?", id).Order("onu_index ASC").Find(&onus)
	
	c.JSON(http.StatusOK, gin.H{
		"total": len(onus),
		"data":  onus,
	})
}

// SaveScanResults saves SNMP scan results to the database, replacing old data
func SaveScanResults(oltID string, onus []OnuData) {
	if len(onus) == 0 {
		return
	}

	// Delete old data for this OLT
	database.DB.Where("olt_id = ?", oltID).Delete(&models.OltOnu{})

	// Batch insert new data (in chunks of 100 for performance)
	batch := make([]models.OltOnu, 0, len(onus))
	for _, onu := range onus {
		batch = append(batch, models.OltOnu{
			OltID:    oltID,
			OnuIndex: onu.Index,
			Name:     onu.Name,
			Serial:   onu.Serial,
			Model:    onu.Model,
			Status:   onu.Status,
			Tx:       onu.Tx,
			Rx:       onu.Rx,
			Distance: onu.Distance,
		})
	}
	
	database.DB.CreateInBatches(&batch, 100)
	log.Printf("[OLT] Saved %d ONUs to database for OLT %s", len(onus), oltID)
}

// ScanAllOLTs scans all active OLTs and saves results (used by cron)
func ScanAllOLTs() {
	var olts []models.Olt
	database.DB.Where("status = ?", "active").Find(&olts)

	if len(olts) == 0 {
		log.Printf("[OLT CRON] No active OLTs to scan.")
		return
	}

	log.Printf("[OLT CRON] Starting auto-scan for %d OLTs...", len(olts))
	for _, o := range olts {
		log.Printf("[OLT CRON] Scanning OLT: %s (%s)", o.Name, o.Host)
		onus, err := ScanONU(o.ID)
		if err != nil {
			log.Printf("[OLT CRON] Failed to scan %s: %v", o.Name, err)
			continue
		}
		SaveScanResults(o.ID, onus)
		log.Printf("[OLT CRON] OLT %s: %d ONUs saved.", o.Name, len(onus))
	}
	log.Printf("[OLT CRON] Auto-scan complete.")
}

func RebootOnuHandler(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		PonPort string `json:"ponPort" binding:"required"`
		OnuID   string `json:"onuId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PON Port dan ONU ID diperlukan"})
		return
	}

	var olt models.Olt
	if err := database.DB.First(&olt, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OLT tidak ditemukan"})
		return
	}

	session, err := NewTelnetSession(&olt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal koneksi telnet: " + err.Error()})
		return
	}
	defer session.Close()

	if err := session.Login(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal login telnet: " + err.Error()})
		return
	}

	output, err := session.RebootOnu(req.PonPort, req.OnuID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal reboot ONU: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Perintah reboot ONU berhasil dikirim",
		"output":  output,
	})
}

// ============================================================================
// OID Mappings
// ============================================================================

func ListMappings(c *gin.Context) {
	var mappings []models.OidMapping
	database.DB.Find(&mappings)
	c.JSON(http.StatusOK, gin.H{"data": mappings})
}

type CreateMappingRequest struct {
	Vendor      string `json:"vendor" binding:"required"`
	OidName     string `json:"oid_name" binding:"required"`
	OidTx       string `json:"oid_tx" binding:"required"`
	OidRx       string `json:"oid_rx" binding:"required"`
	OidStatus   string `json:"oid_status" binding:"required"`
	OidSerial   string `json:"oid_serial"`
	OidModel    string `json:"oid_model"`
	OidDistance  string `json:"oid_distance"`
}

func CreateMapping(c *gin.Context) {
	var req CreateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	mapping := models.OidMapping{
		Vendor:      req.Vendor,
		OidName:     req.OidName,
		OidTx:       req.OidTx,
		OidRx:       req.OidRx,
		OidStatus:   req.OidStatus,
		OidSerial:   req.OidSerial,
		OidModel:    req.OidModel,
		OidDistance:  req.OidDistance,
	}

	if err := database.DB.Create(&mapping).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan Mapping: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Mapping berhasil ditambahkan", "data": mapping})
}

func UpdateMapping(c *gin.Context) {
	id := c.Param("id")
	var mapping models.OidMapping
	if err := database.DB.First(&mapping, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mapping tidak ditemukan"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	database.DB.Model(&mapping).Updates(req)
	database.DB.First(&mapping, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{"message": "Mapping berhasil diupdate", "data": mapping})
}

func DeleteMapping(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.OidMapping{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus Mapping"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Mapping berhasil dihapus"})
}


package whatsapp

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers WhatsApp endpoints
func RegisterRoutes(r *gin.RouterGroup) {
	wa := r.Group("/whatsapp")
	{
		// Providers
		wa.GET("/providers", ListProviders)
		wa.POST("/providers", CreateProvider)
		wa.PUT("/providers/:id", UpdateProvider)
		wa.DELETE("/providers/:id", DeleteProvider)

		// Templates
		wa.GET("/templates", ListTemplates)
		wa.POST("/templates", CreateTemplate)
		wa.PUT("/templates/:id", UpdateTemplate)
		wa.DELETE("/templates/:id", DeleteTemplate)

		// History
		wa.GET("/history", ListHistory)

		// Manual Sending
		wa.POST("/send", SendMessageManual)
	}
}

// ============================================================================
// Providers
// ============================================================================

func ListProviders(c *gin.Context) {
	var providers []models.WhatsAppProvider
	database.DB.Order("priority DESC, createdAt DESC").Find(&providers)
	c.JSON(http.StatusOK, gin.H{"data": providers})
}

type ProviderRequest struct {
	Name         string  `json:"name" binding:"required"`
	Type         string  `json:"type" binding:"required"`
	APIKey       string  `json:"apiKey" binding:"required"`
	APIUrl       string  `json:"apiUrl" binding:"required"`
	SenderNumber *string `json:"senderNumber"`
	IsActive     *bool   `json:"isActive"`
	Priority     *int    `json:"priority"`
	Description  *string `json:"description"`
}

func CreateProvider(c *gin.Context) {
	var req ProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	isActive := true
	if req.IsActive != nil { isActive = *req.IsActive }
	
	priority := 0
	if req.Priority != nil { priority = *req.Priority }

	provider := models.WhatsAppProvider{
		ID:           "wap_" + strconv.FormatInt(time.Now().UnixNano(), 36),
		Name:         req.Name,
		Type:         req.Type,
		APIKey:       req.APIKey,
		APIUrl:       req.APIUrl,
		SenderNumber: req.SenderNumber,
		IsActive:     isActive,
		Priority:     priority,
		Description:  req.Description,
	}

	if err := database.DB.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan provider: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Provider berhasil ditambahkan", "data": provider})
}

func UpdateProvider(c *gin.Context) {
	id := c.Param("id")
	var provider models.WhatsAppProvider
	if err := database.DB.First(&provider, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider tidak ditemukan"})
		return
	}

	var req ProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	updates := map[string]interface{}{
		"name":   req.Name,
		"type":   req.Type,
		"apiKey": req.APIKey,
		"apiUrl": req.APIUrl,
	}
	if req.SenderNumber != nil { updates["senderNumber"] = req.SenderNumber }
	if req.IsActive != nil { updates["isActive"] = req.IsActive }
	if req.Priority != nil { updates["priority"] = req.Priority }
	if req.Description != nil { updates["description"] = req.Description }

	database.DB.Model(&provider).Updates(updates)
	database.DB.First(&provider, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{"message": "Provider berhasil diupdate", "data": provider})
}

func DeleteProvider(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.WhatsAppProvider{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus provider"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Provider berhasil dihapus"})
}

// ============================================================================
// Templates
// ============================================================================

func ListTemplates(c *gin.Context) {
	var templates []models.WhatsAppTemplate
	database.DB.Order("type ASC").Find(&templates)
	c.JSON(http.StatusOK, gin.H{"data": templates})
}

type TemplateRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Message  string `json:"message" binding:"required"`
	IsActive *bool  `json:"isActive"`
}

func CreateTemplate(c *gin.Context) {
	var req TemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	var existing models.WhatsAppTemplate
	if err := database.DB.Where("type = ?", req.Type).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Tipe template sudah ada"})
		return
	}

	isActive := true
	if req.IsActive != nil { isActive = *req.IsActive }

	template := models.WhatsAppTemplate{
		ID:       "tpl_" + strconv.FormatInt(time.Now().UnixNano(), 36),
		Name:     req.Name,
		Type:     req.Type,
		Message:  req.Message,
		IsActive: isActive,
	}

	if err := database.DB.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Template berhasil dibuat", "data": template})
}

func UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	var template models.WhatsAppTemplate
	if err := database.DB.First(&template, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template tidak ditemukan"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	database.DB.Model(&template).Updates(req)
	database.DB.First(&template, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{"message": "Template berhasil diupdate", "data": template})
}

func DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.WhatsAppTemplate{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"message": "Template berhasil dihapus"})
}

// ============================================================================
// History
// ============================================================================

func ListHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")

	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	query := database.DB.Model(&models.WhatsAppHistory{})
	
	if status != "" { query = query.Where("status = ?", status) }
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("phone LIKE ? OR message LIKE ?", like, like)
	}

	var total int64
	query.Count(&total)

	var history []models.WhatsAppHistory
	query.Order("sentAt DESC").Offset(offset).Limit(limit).Find(&history)

	c.JSON(http.StatusOK, gin.H{
		"data": history,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// ============================================================================
// Manual Sending & Internal Helper
// ============================================================================

type SendRequest struct {
	Phone   string `json:"phone" binding:"required"`
	Message string `json:"message" binding:"required"`
}

func SendMessageManual(c *gin.Context) {
	var req SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	err := SendWhatsAppMessage(req.Phone, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim pesan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pesan berhasil dikirim (diproses background)"})
}

// SendWhatsAppMessage is an internal helper to send WA via active provider
func SendWhatsAppMessage(phone string, message string) error {
	// 1. Get highest priority active provider
	var provider models.WhatsAppProvider
	if err := database.DB.Where("isActive = ?", true).Order("priority DESC").First(&provider).Error; err != nil {
		// Log failed if no provider
		logHistory(phone, message, "FAILED", "No active WhatsApp provider found", nil, nil)
		return err
	}

	// Remove leading 0 and replace with 62
	if strings.HasPrefix(phone, "0") {
		phone = "62" + phone[1:]
	}

	// 2. Send via goroutine
	go func(p models.WhatsAppProvider, targetPhone, msg string) {
		var responseBody string
		var status string = "SENT"

		// Very basic Fonnte implementation example (can be expanded for Wablas, etc)
		if p.Type == "fonnte" {
			reqBody, _ := json.Marshal(map[string]string{
				"target": targetPhone,
				"message": msg,
			})
			req, _ := http.NewRequest("POST", p.APIUrl, bytes.NewBuffer(reqBody))
			req.Header.Set("Authorization", p.APIKey)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 10 * time.Second}
			res, err := client.Do(req)
			if err != nil {
				status = "FAILED"
				responseBody = err.Error()
			} else {
				defer res.Body.Close()
				body, _ := io.ReadAll(res.Body)
				responseBody = string(body)
				if res.StatusCode >= 400 {
					status = "FAILED"
				}
			}
		} else {
			// Other providers...
			status = "FAILED"
			responseBody = "Provider type not supported yet by Go backend: " + p.Type
		}

		logHistory(targetPhone, msg, status, responseBody, &p.Name, &p.Type)
	}(provider, phone, message)

	return nil
}

func logHistory(phone, message, status, response string, pName, pType *string) {
	var respPtr *string
	if response != "" { respPtr = &response }

	log := models.WhatsAppHistory{
		ID:           "wah_" + strconv.FormatInt(time.Now().UnixNano(), 36),
		Phone:        phone,
		Message:      message,
		Status:       status,
		Response:     respPtr,
		ProviderName: pName,
		ProviderType: pType,
	}
	database.DB.Create(&log)
}

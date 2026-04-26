package billing

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers billing routes
func RegisterRoutes(r *gin.RouterGroup) {
	billing := r.Group("/invoices")
	{
		billing.GET("", ListInvoices)
		billing.GET("/:id", GetInvoice)
		billing.POST("", CreateInvoice)
		billing.PUT("/:id/status", UpdateInvoiceStatus)
	}
}

// ListInvoices returns paginated invoices
func ListInvoices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")
	userID := c.Query("userId")

	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	query := database.DB.Model(&models.Invoice{})

	// Filters
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("invoiceNumber LIKE ? OR customerName LIKE ?", like, like)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if userID != "" {
		query = query.Where("userId = ?", userID)
	}

	var total int64
	query.Count(&total)

	var invoices []models.Invoice
	result := query.
		Preload("User").
		Order("createdAt DESC").
		Offset(offset).
		Limit(limit).
		Find(&invoices)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": invoices,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// GetInvoice returns a single invoice
func GetInvoice(c *gin.Context) {
	id := c.Param("id")

	var invoice models.Invoice
	if err := database.DB.Preload("User").First(&invoice, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoice})
}

// CreateInvoiceRequest is the payload for creating an invoice
type CreateInvoiceRequest struct {
	UserID     string `json:"userId" binding:"required"`
	Amount     int    `json:"amount" binding:"required"`
	BaseAmount *int   `json:"baseAmount"`
	Notes      string `json:"notes"`
	DaysDue    int    `json:"daysDue"` // How many days until due date
}

// CreateInvoice creates a new manual invoice
func CreateInvoice(c *gin.Context) {
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	// Get user to populate customer details
	var user models.PppoeUser
	if err := database.DB.First(&user, "id = ?", req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	now := time.Now()
	daysDue := req.DaysDue
	if daysDue <= 0 { daysDue = 7 }
	dueDate := now.AddDate(0, 0, daysDue)

	invoice := models.Invoice{
		ID:               generateID(),
		InvoiceNumber:    generateInvoiceNumber(),
		UserID:           &user.ID,
		Amount:           req.Amount,
		BaseAmount:       req.BaseAmount,
		Status:           "PENDING",
		DueDate:          dueDate,
		CustomerName:     &user.Name,
		CustomerPhone:    &user.Phone,
		CustomerEmail:    user.Email,
		CustomerUsername: &user.Username,
		Notes:            &req.Notes,
		InvoiceType:      "MANUAL",
	}

	if err := database.DB.Create(&invoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat invoice: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Invoice berhasil dibuat",
		"data":    invoice,
	})
}

// UpdateInvoiceStatusRequest payload
type UpdateInvoiceStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateInvoiceStatus updates status (e.g. to PAID)
func UpdateInvoiceStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateInvoiceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	var invoice models.Invoice
	if err := database.DB.First(&invoice, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice tidak ditemukan"})
		return
	}

	updates := map[string]interface{}{"status": req.Status}
	if req.Status == "PAID" && invoice.PaidAt == nil {
		now := time.Now()
		updates["paidAt"] = &now
	}

	if err := database.DB.Model(&invoice).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update status: " + err.Error()})
		return
	}

	// Note: in a real world, marking PAID should trigger auto-renewal logic and CoA if isolated
	database.DB.First(&invoice, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{
		"message": "Status invoice berhasil diupdate",
		"data":    invoice,
	})
}

// Helpers
func generateID() string {
	now := time.Now()
	return "inv_" + strconv.FormatInt(now.UnixNano(), 36)
}

func generateInvoiceNumber() string {
	now := time.Now()
	// Format: INV/YYYY/MM/DD/RANDOM
	return "INV/" + now.Format("2006/01/02/") + strconv.FormatInt(now.UnixNano()%10000, 10)
}

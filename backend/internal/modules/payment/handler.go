package payment

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers payment gateway routes
func RegisterRoutes(r *gin.RouterGroup) {
	gateways := r.Group("/payment-gateways")
	{
		gateways.GET("", ListGateways)
		gateways.GET("/:id", GetGateway)
		gateways.PUT("/:id", UpdateGateway)
		gateways.POST("/:id/toggle", ToggleGateway)
	}

	webhooks := r.Group("/webhooks")
	{
		webhooks.GET("/logs", ListWebhookLogs)
	}
}

// RegisterPublicRoutes registers unauthenticated webhook listeners
func RegisterPublicRoutes(r *gin.RouterGroup) {
	webhooks := r.Group("/api/webhooks")
	{
		webhooks.POST("/midtrans", MidtransWebhook)
		webhooks.POST("/tripay", TripayWebhook)
		webhooks.POST("/duitku", DuitkuWebhook)
		webhooks.POST("/xendit", XenditWebhook)
	}
}

// ============================================================================
// Gateway Management
// ============================================================================

func ListGateways(c *gin.Context) {
	var gateways []models.PaymentGateway
	database.DB.Order("provider ASC").Find(&gateways)
	c.JSON(http.StatusOK, gin.H{"data": gateways})
}

func GetGateway(c *gin.Context) {
	id := c.Param("id")
	var gateway models.PaymentGateway
	if err := database.DB.First(&gateway, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gateway tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gateway})
}

func UpdateGateway(c *gin.Context) {
	id := c.Param("id")
	var gateway models.PaymentGateway
	if err := database.DB.First(&gateway, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gateway tidak ditemukan"})
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	database.DB.Model(&gateway).Updates(body)
	database.DB.First(&gateway, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"message": "Gateway berhasil diupdate", "data": gateway})
}

func ToggleGateway(c *gin.Context) {
	id := c.Param("id")
	var gateway models.PaymentGateway
	if err := database.DB.First(&gateway, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gateway tidak ditemukan"})
		return
	}

	newState := !gateway.IsActive

	// If enabling, we might want to disable others of the same provider type if logic dictates
	database.DB.Model(&gateway).Update("isActive", newState)
	gateway.IsActive = newState

	c.JSON(http.StatusOK, gin.H{
		"message": "Status gateway berhasil diubah",
		"data":    gateway,
	})
}

// ============================================================================
// Webhook Logs
// ============================================================================

func ListWebhookLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	var total int64
	database.DB.Model(&models.WebhookLog{}).Count(&total)

	var logs []models.WebhookLog
	database.DB.Order("createdAt DESC").Offset(offset).Limit(limit).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"total": total,
	})
}

// ============================================================================
// Webhook Handlers
// ============================================================================

// SaveLog helper
func saveWebhookLog(gateway, orderId, status string, payload string, success bool, errMsg string) {
	log := models.WebhookLog{
		ID:           "wh_" + strconv.FormatInt(time.Now().UnixNano(), 36),
		Gateway:      gateway,
		OrderID:      orderId,
		Status:       status,
		Payload:      &payload,
		Success:      success,
	}
	if errMsg != "" {
		log.ErrorMessage = &errMsg
	}
	database.DB.Create(&log)
}

// processInvoicePayment marks invoice as paid
func processInvoicePayment(invoiceNumber string, amount int, method string) error {
	var invoice models.Invoice
	if err := database.DB.First(&invoice, "invoiceNumber = ?", invoiceNumber).Error; err != nil {
		return err // Or handle "not found"
	}

	if invoice.Status == "PAID" {
		return nil // Already paid
	}

	now := time.Now()
	database.DB.Model(&invoice).Updates(map[string]interface{}{
		"status": "PAID",
		"paidAt": now,
	})

	// Record payment
	payment := models.Payment{
		ID:        "pay_" + strconv.FormatInt(now.UnixNano(), 36),
		InvoiceID: invoice.ID,
		Amount:    amount,
		Method:    method,
		Status:    "SUCCESS",
	}
	database.DB.Create(&payment)

	return nil
}

// MidtransWebhook handler
func MidtransWebhook(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	payloadStr := string(body)

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	orderId, _ := payload["order_id"].(string)
	transactionStatus, _ := payload["transaction_status"].(string)
	grossAmount, _ := payload["gross_amount"].(string)
	
	amount := 0
	if grossAmount != "" {
		parsed, _ := strconv.ParseFloat(grossAmount, 64)
		amount = int(parsed)
	}

	if transactionStatus == "capture" || transactionStatus == "settlement" {
		err := processInvoicePayment(orderId, amount, "MIDTRANS")
		if err != nil {
			saveWebhookLog("midtrans", orderId, transactionStatus, payloadStr, false, err.Error())
		} else {
			saveWebhookLog("midtrans", orderId, transactionStatus, payloadStr, true, "")
		}
	} else {
		saveWebhookLog("midtrans", orderId, transactionStatus, payloadStr, true, "")
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// TripayWebhook handler
func TripayWebhook(c *gin.Context) {
	// Add Tripay signature validation here in real world
	body, _ := io.ReadAll(c.Request.Body)
	payloadStr := string(body)

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	merchantRef, _ := payload["merchant_ref"].(string)
	status, _ := payload["status"].(string)
	
	if status == "PAID" {
		err := processInvoicePayment(merchantRef, 0, "TRIPAY") // Need to parse amount properly
		if err != nil {
			saveWebhookLog("tripay", merchantRef, status, payloadStr, false, err.Error())
		} else {
			saveWebhookLog("tripay", merchantRef, status, payloadStr, true, "")
		}
	} else {
		saveWebhookLog("tripay", merchantRef, status, payloadStr, true, "")
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DuitkuWebhook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func XenditWebhook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

package customer

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers PPPoE customer routes
func RegisterRoutes(r *gin.RouterGroup) {
	cust := r.Group("/pppoe/customers")
	{
		cust.GET("", ListCustomers)
		cust.GET("/:id", GetCustomer)
		cust.POST("", CreateCustomer)
		cust.PUT("/:id", UpdateCustomer)
		cust.DELETE("/:id", DeleteCustomer)
	}
}

// ListCustomers returns paginated customers
func ListCustomers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	query := database.DB.Model(&models.PppoeCustomer{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR phone LIKE ? OR customerId LIKE ? OR email LIKE ?", like, like, like, like)
	}

	var total int64
	query.Count(&total)

	var customers []models.PppoeCustomer
	query.Preload("Area").Preload("PppoeUsers").
		Order("createdAt DESC").Offset(offset).Limit(limit).Find(&customers)

	c.JSON(http.StatusOK, gin.H{
		"data": customers,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// GetCustomer returns a single customer with their PPPoE users
func GetCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.PppoeCustomer
	if err := database.DB.Preload("Area").Preload("PppoeUsers").First(&customer, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": customer})
}

// CreateCustomerRequest payload
type CreateCustomerRequest struct {
	Name         string  `json:"name" binding:"required"`
	Phone        string  `json:"phone" binding:"required"`
	Email        *string `json:"email"`
	Address      *string `json:"address"`
	IDCardNumber *string `json:"idCardNumber"`
	IDCardPhoto  *string `json:"idCardPhoto"`
	AreaID       *string `json:"areaId"`
}

// CreateCustomer creates a new PPPoE customer
func CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	// Generate unique 8-digit customer ID
	var customerID string
	for {
		customerID = strconv.Itoa(10000000 + int(time.Now().UnixNano()%90000000))
		var count int64
		database.DB.Model(&models.PppoeCustomer{}).Where("customerId = ?", customerID).Count(&count)
		if count == 0 { break }
	}

	customer := models.PppoeCustomer{
		ID:           generateCustID(),
		CustomerID:   customerID,
		Name:         req.Name,
		Phone:        req.Phone,
		Email:        req.Email,
		Address:      req.Address,
		IDCardNumber: req.IDCardNumber,
		IDCardPhoto:  req.IDCardPhoto,
		AreaID:       req.AreaID,
		IsActive:     true,
	}

	if err := database.DB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat customer: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Customer berhasil dibuat", "data": customer})
}

// UpdateCustomer updates a customer
func UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.PppoeCustomer
	if err := database.DB.First(&customer, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer tidak ditemukan"})
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	// Prevent changing immutable fields
	delete(body, "id")
	delete(body, "customerId")

	database.DB.Model(&customer).Updates(body)
	database.DB.Preload("Area").First(&customer, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{"message": "Customer berhasil diupdate", "data": customer})
}

// DeleteCustomer deletes a customer (only if no PPPoE users linked)
func DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.PppoeCustomer
	if err := database.DB.First(&customer, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer tidak ditemukan"})
		return
	}

	// Check if PPPoE users are linked
	var count int64
	database.DB.Model(&models.PppoeUser{}).Where("pppoe_customer_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Customer masih memiliki user PPPoE, hapus user terlebih dahulu"})
		return
	}

	database.DB.Delete(&customer)
	c.JSON(http.StatusOK, gin.H{"message": "Customer berhasil dihapus"})
}

func generateCustID() string {
	return "cust_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

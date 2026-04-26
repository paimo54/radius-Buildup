package company

import (
	"net/http"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers company settings routes
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/company", GetCompany)
	r.PUT("/company", UpdateCompany)
}

// RegisterPublicRoutes registers public company info (for login page branding)
func RegisterPublicRoutes(r *gin.RouterGroup) {
	r.GET("/settings/company", GetPublicCompanyInfo)
}

// GetCompany returns full company settings (admin only)
func GetCompany(c *gin.Context) {
	var company models.Company
	result := database.DB.First(&company)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company settings belum dikonfigurasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": company})
}

// GetPublicCompanyInfo returns limited company info for login pages
func GetPublicCompanyInfo(c *gin.Context) {
	var company models.Company
	result := database.DB.First(&company)

	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"name":    "SALFANET RADIUS",
				"logo":    nil,
				"address": nil,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"name":            company.Name,
			"logo":            company.Logo,
			"address":         company.Address,
			"phone":           company.Phone,
			"poweredBy":       company.PoweredBy,
			"footerAdmin":     nil,
			"footerCustomer":  nil,
			"footerTechnician": nil,
			"footerAgent":     nil,
		},
	})
}

// UpdateCompanyRequest payload
type UpdateCompanyRequest struct {
	Name                *string `json:"name"`
	Address             *string `json:"address"`
	Phone               *string `json:"phone"`
	Email               *string `json:"email"`
	Logo                *string `json:"logo"`
	AdminPhone          *string `json:"adminPhone"`
	BaseURL             *string `json:"baseUrl"`
	Timezone            *string `json:"timezone"`
	PoweredBy           *string `json:"poweredBy"`
	CustomerIDPrefix    *string `json:"customerIdPrefix"`
	InvoiceGenerateDays *int    `json:"invoiceGenerateDays"`
	GracePeriodDays     *int    `json:"gracePeriodDays"`
	IsolationEnabled    *bool   `json:"isolationEnabled"`
	IsolationIPPool     *string `json:"isolationIpPool"`
	IsolationRateLimit  *string `json:"isolationRateLimit"`
}

// UpdateCompany updates company settings
func UpdateCompany(c *gin.Context) {
	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	var company models.Company
	result := database.DB.First(&company)

	updates := map[string]interface{}{}
	if req.Name != nil { updates["name"] = *req.Name }
	if req.Address != nil { updates["address"] = *req.Address }
	if req.Phone != nil { updates["phone"] = *req.Phone }
	if req.Email != nil { updates["email"] = *req.Email }
	if req.Logo != nil { updates["logo"] = *req.Logo }
	if req.AdminPhone != nil { updates["adminPhone"] = *req.AdminPhone }
	if req.BaseURL != nil { updates["baseUrl"] = *req.BaseURL }
	if req.Timezone != nil { updates["timezone"] = *req.Timezone }
	if req.PoweredBy != nil { updates["poweredBy"] = *req.PoweredBy }
	if req.CustomerIDPrefix != nil { updates["customerIdPrefix"] = *req.CustomerIDPrefix }
	if req.InvoiceGenerateDays != nil { updates["invoiceGenerateDays"] = *req.InvoiceGenerateDays }
	if req.GracePeriodDays != nil { updates["gracePeriodDays"] = *req.GracePeriodDays }
	if req.IsolationEnabled != nil { updates["isolationEnabled"] = *req.IsolationEnabled }
	if req.IsolationIPPool != nil { updates["isolationIpPool"] = *req.IsolationIPPool }
	if req.IsolationRateLimit != nil { updates["isolationRateLimit"] = *req.IsolationRateLimit }

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada data yang diubah"})
		return
	}

	if result.Error != nil {
		// Company doesn't exist yet — create with defaults
		name := "ISP Company"
		if req.Name != nil { name = *req.Name }
		company = models.Company{ID: "company_1", Name: name}
		database.DB.Create(&company)
	}

	database.DB.Model(&company).Updates(updates)
	database.DB.First(&company, "id = ?", company.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Pengaturan perusahaan berhasil diupdate",
		"data":    company,
	})
}

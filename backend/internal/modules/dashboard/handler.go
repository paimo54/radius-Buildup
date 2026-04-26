package dashboard

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers dashboard routes
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/dashboard/stats", GetStats)
}

// GetStats returns dashboard statistics
func GetStats(c *gin.Context) {
	now := time.Now()

	// Parse optional ?month=YYYY-MM
	monthParam := c.Query("month")
	var selectedYear, selectedMonth int
	if monthParam != "" && len(monthParam) == 7 {
		fmt.Sscanf(monthParam, "%d-%d", &selectedYear, &selectedMonth)
		selectedMonth-- // 0-indexed
	} else {
		selectedYear = now.Year()
		selectedMonth = int(now.Month()) - 1
	}

	monthNames := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	startOfMonth := time.Date(selectedYear, time.Month(selectedMonth+1), 1, 0, 0, 0, 0, now.Location())
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	periodLabel := fmt.Sprintf("%s %d", monthNames[selectedMonth], selectedYear)
	monthKey := fmt.Sprintf("%d-%02d", selectedYear, selectedMonth+1)
	isCurrentMonth := selectedYear == now.Year() && selectedMonth == int(now.Month())-1

	// ==================== 1. Total PPPoE Users ====================
	var totalPppoeUsers int64
	database.DB.Model(&models.PppoeUser{}).Count(&totalPppoeUsers)

	// ==================== 2. Active PPPoE Users ====================
	var activePppoeUsers int64
	database.DB.Model(&models.PppoeUser{}).Where("status IN ?", []string{"active", "ACTIVE"}).Count(&activePppoeUsers)

	// ==================== 3. Active Sessions (PPPoE) ====================
	var activeSessionsPPPoE int64
	database.DB.Model(&models.Radacct{}).Where("acctstoptime IS NULL").Count(&activeSessionsPPPoE)

	// ==================== 4. Isolated / Blocked ====================
	var isolatedCount int64
	database.DB.Model(&models.PppoeUser{}).Where("status IN ?", []string{"isolated", "ISOLATED", "blocked", "BLOCKED"}).Count(&isolatedCount)

	// ==================== 5. Suspended ====================
	var suspendedCount int64
	database.DB.Model(&models.PppoeUser{}).Where("status IN ?", []string{"suspended", "SUSPENDED"}).Count(&suspendedCount)

	// ==================== 6. Invoice Revenue (month + today) ====================
	type SumResult struct {
		Total float64
		Count int64
	}

	var invoiceMonth SumResult
	database.DB.Model(&models.Invoice{}).
		Select("COALESCE(SUM(amount), 0) as total, COUNT(*) as count").
		Where("status = ? AND paidAt >= ? AND paidAt < ?", "PAID", startOfMonth, startOfNextMonth).
		Scan(&invoiceMonth)

	var invoiceToday SumResult
	database.DB.Model(&models.Invoice{}).
		Select("COALESCE(SUM(amount), 0) as total, COUNT(*) as count").
		Where("status = ? AND paidAt >= ?", "PAID", startOfToday).
		Scan(&invoiceToday)

	var unpaidInvoicesCount int64
	database.DB.Model(&models.Invoice{}).Where("status IN ?", []string{"PENDING", "OVERDUE"}).Count(&unpaidInvoicesCount)

	var totalAllTime SumResult
	database.DB.Model(&models.Invoice{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("status = ?", "PAID").
		Scan(&totalAllTime)

	// ==================== 7. Upcoming / Overdue Invoices ====================
	sevenDaysFromNow := now.AddDate(0, 0, 7)
	var upcomingInvoices []struct {
		InvoiceNumber    string     `json:"invoiceNumber"`
		CustomerName     *string    `json:"customerName"`
		CustomerUsername *string    `json:"customerUsername"`
		Amount           int        `json:"amount"`
		DueDate          time.Time  `json:"dueDate"`
		Status           string     `json:"status"`
	}
	database.DB.Model(&models.Invoice{}).
		Select("invoiceNumber, customerName, customerUsername, amount, dueDate, status").
		Where("status IN ? AND dueDate <= ?", []string{"PENDING", "OVERDUE"}, sevenDaysFromNow).
		Order("dueDate ASC").
		Limit(20).
		Find(&upcomingInvoices)

	// Build upcoming list with daysUntilDue
	type UpcomingInvoiceDTO struct {
		InvoiceNumber    string  `json:"invoiceNumber"`
		CustomerName     string  `json:"customerName"`
		CustomerUsername string  `json:"customerUsername"`
		Amount           int     `json:"amount"`
		DueDate          string  `json:"dueDate"`
		Status           string  `json:"status"`
		DaysUntilDue     int     `json:"daysUntilDue"`
	}
	upcomingList := make([]UpcomingInvoiceDTO, 0)
	for _, inv := range upcomingInvoices {
		name := "-"
		if inv.CustomerName != nil { name = *inv.CustomerName }
		uname := "-"
		if inv.CustomerUsername != nil { uname = *inv.CustomerUsername }
		days := int(math.Ceil(float64(inv.DueDate.Sub(now)) / float64(24*time.Hour)))
		upcomingList = append(upcomingList, UpcomingInvoiceDTO{
			InvoiceNumber:    inv.InvoiceNumber,
			CustomerName:     name,
			CustomerUsername: uname,
			Amount:           inv.Amount,
			DueDate:          inv.DueDate.Format(time.RFC3339),
			Status:           inv.Status,
			DaysUntilDue:     days,
		})
	}

	// ==================== 8. RADIUS Auth Log ====================
	var radiusAuthLog []struct {
		Username string    `json:"username"`
		Reply    string    `json:"reply"`
		Authdate time.Time `json:"authdate"`
	}
	database.DB.Model(&models.Radpostauth{}).
		Select("username, reply, authdate").
		Order("authdate DESC").
		Limit(15).
		Find(&radiusAuthLog)

	var acceptToday, rejectToday int64
	database.DB.Model(&models.Radpostauth{}).
		Where("reply = ? AND authdate >= ?", "Access-Accept", startOfToday).
		Count(&acceptToday)
	database.DB.Model(&models.Radpostauth{}).
		Where("reply = ? AND authdate >= ?", "Access-Reject", startOfToday).
		Count(&rejectToday)

	// ==================== 9. System Status ====================
	radiusStatus := false
	oneHourAgo := now.Add(-1 * time.Hour)
	var recentAcct int64
	database.DB.Model(&models.Radacct{}).Where("acctstarttime >= ?", oneHourAgo).Count(&recentAcct)
	if recentAcct > 0 {
		radiusStatus = true
	}

	// ==================== Format Currency ====================
	formatCurrency := func(amount float64) string {
		return fmt.Sprintf("Rp%s", formatNumber(int64(amount)))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats": gin.H{
			"totalPppoeUsers":                totalPppoeUsers,
			"activePppoeUsers":               activePppoeUsers,
			"activeSessionsPPPoE":            activeSessionsPPPoE,
			"activeSessionsHotspot":          0, // TODO: implement hotspot model
			"unusedVouchers":                 0, // TODO: implement hotspot model
			"isolatedCount":                  isolatedCount,
			"suspendedCount":                 suspendedCount,
			"newRegistrations":               0, // TODO: implement registration model
			"upcomingInvoices":               upcomingList,
			"invoiceRevenue":                 int64(invoiceMonth.Total),
			"invoiceRevenueFormatted":        formatCurrency(invoiceMonth.Total),
			"invoiceRevenueToday":            int64(invoiceToday.Total),
			"invoiceRevenueTodayFormatted":   formatCurrency(invoiceToday.Total),
			"invoiceCountToday":              invoiceToday.Count,
			"invoiceCountMonth":              invoiceMonth.Count,
			"unpaidInvoicesCount":            unpaidInvoicesCount,
			"totalAllTimeRevenue":            int64(totalAllTime.Total),
			"totalAllTimeRevenueFormatted":   formatCurrency(totalAllTime.Total),
			"voucherRevenue":                 0, // TODO
			"voucherRevenueFormatted":        "Rp0",
			"voucherRevenueToday":            0,
			"voucherRevenueTodayFormatted":   "Rp0",
		},
		"activities":      []interface{}{}, // TODO: implement activity log
		"systemStatus": gin.H{
			"radius":   radiusStatus,
			"database": true,
			"api":      true,
		},
		"agentSales":      []interface{}{},
		"agentSalesTotal": gin.H{"count": 0, "revenue": 0},
		"radiusAuthLog":   radiusAuthLog,
		"radiusAuthStats": gin.H{
			"acceptToday": acceptToday,
			"rejectToday": rejectToday,
		},
		"periodLabel":    periodLabel,
		"monthKey":       monthKey,
		"isCurrentMonth": isCurrentMonth,
	})
}

// formatNumber adds thousands separators
func formatNumber(n int64) string {
	if n < 0 {
		return "-" + formatNumber(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return result
}

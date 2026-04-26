package activity

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers activity log routes
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/activity-logs", ListActivityLogs)
}

// ListActivityLogs returns paginated activity logs
func ListActivityLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	module := c.Query("module")
	search := c.Query("search")

	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	query := database.DB.Model(&models.ActivityLog{})

	if module != "" {
		query = query.Where("module = ?", module)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("username LIKE ? OR description LIKE ? OR action LIKE ?", like, like, like)
	}

	var total int64
	query.Count(&total)

	var logs []models.ActivityLog
	query.Order("createdAt DESC").Offset(offset).Limit(limit).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// LogActivity creates a new activity log entry (called from other modules)
func LogActivity(userID, username, userRole, action, description, module, status, ipAddress string, metadata interface{}) {
	var metadataStr *string
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			s := string(b)
			metadataStr = &s
		}
	}

	var userIDPtr, userRolePtr, ipPtr *string
	if userID != "" { userIDPtr = &userID }
	if userRole != "" { userRolePtr = &userRole }
	if ipAddress != "" { ipPtr = &ipAddress }

	log := models.ActivityLog{
		ID:          generateLogID(),
		UserID:      userIDPtr,
		Username:    username,
		UserRole:    userRolePtr,
		Action:      action,
		Description: description,
		Module:      module,
		Status:      status,
		IPAddress:   ipPtr,
		Metadata:    metadataStr,
	}

	// Insert in background — don't block the request
	go func() {
		database.DB.Create(&log)
	}()
}

func generateLogID() string {
	now := time.Now()
	return "log_" + strconv.FormatInt(now.UnixNano(), 36)
}

package sessions

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"radius-buildup/internal/database"
	"radius-buildup/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers session routes
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/sessions", ListSessions)
	r.GET("/sessions/active", ListActiveSessions)
}

// SessionDTO is the response format for a session
type SessionDTO struct {
	RadacctID       int64   `json:"radacctid"`
	Username        string  `json:"username"`
	CustomerName    string  `json:"customerName"`
	NASIPAddress    string  `json:"nasipaddress"`
	FramedIPAddress string  `json:"framedipaddress"`
	CallingStation  string  `json:"callingstationid"`
	StartTime       string  `json:"acctstarttime"`
	StopTime        *string `json:"acctstoptime"`
	SessionTime     int     `json:"acctsessiontime"`
	InputOctets     int64   `json:"acctinputoctets"`
	OutputOctets    int64   `json:"acctoutputoctets"`
	TerminateCause  string  `json:"acctterminatecause"`
	IsOnline        bool    `json:"isOnline"`
	Uptime          string  `json:"uptime"`
}

// ListSessions returns paginated sessions (all sessions, including stopped)
func ListSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status") // "active" or "stopped"

	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	query := database.DB.Model(&models.Radacct{})

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("username LIKE ? OR framedipaddress LIKE ? OR callingstationid LIKE ?", like, like, like)
	}
	if status == "active" {
		query = query.Where("acctstoptime IS NULL")
	} else if status == "stopped" {
		query = query.Where("acctstoptime IS NOT NULL")
	}

	var total int64
	query.Count(&total)

	var sessions []models.Radacct
	query.Order("acctstarttime DESC").Offset(offset).Limit(limit).Find(&sessions)

	// Batch lookup PPPoE usernames for customer names
	usernames := make([]string, 0, len(sessions))
	for _, s := range sessions {
		usernames = append(usernames, s.Username)
	}

	nameMap := make(map[string]string)
	if len(usernames) > 0 {
		var users []struct {
			Username string
			Name     string
		}
		database.DB.Model(&models.PppoeUser{}).
			Select("username, name").
			Where("username IN ?", usernames).
			Find(&users)
		for _, u := range users {
			nameMap[u.Username] = u.Name
		}
	}

	now := time.Now()
	dtos := make([]SessionDTO, 0, len(sessions))
	for _, s := range sessions {
		isOnline := s.AcctStopTime == nil

		// Calculate real-time uptime
		var sessionTime int
		if isOnline && s.AcctStartTime != nil {
			sessionTime = int(now.Sub(*s.AcctStartTime).Seconds())
		} else if s.AcctSessionTime != nil {
			sessionTime = int(*s.AcctSessionTime)
		}

		var stopStr *string
		if s.AcctStopTime != nil {
			t := s.AcctStopTime.Format(time.RFC3339)
			stopStr = &t
		}

		startStr := ""
		if s.AcctStartTime != nil {
			startStr = s.AcctStartTime.Format(time.RFC3339)
		}

		var inputOctets, outputOctets int64
		if s.AcctInputOctets != nil { inputOctets = *s.AcctInputOctets }
		if s.AcctOutputOctets != nil { outputOctets = *s.AcctOutputOctets }

		customerName := s.Username
		if name, ok := nameMap[s.Username]; ok {
			customerName = name
		}

		dtos = append(dtos, SessionDTO{
			RadacctID:       s.RadacctID,
			Username:        s.Username,
			CustomerName:    customerName,
			NASIPAddress:    s.NASIPAddress,
			FramedIPAddress: s.FramedIPAddress,
			CallingStation:  s.CallingStationID,
			StartTime:       startStr,
			StopTime:        stopStr,
			SessionTime:     sessionTime,
			InputOctets:     inputOctets,
			OutputOctets:    outputOctets,
			TerminateCause:  s.AcctTerminateCause,
			IsOnline:        isOnline,
			Uptime:          formatDuration(sessionTime),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dtos,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// ListActiveSessions returns only currently active sessions
func ListActiveSessions(c *gin.Context) {
	c.Set("_forceStatus", "active")

	var sessions []models.Radacct
	database.DB.Where("acctstoptime IS NULL").Order("acctstarttime DESC").Find(&sessions)

	// Batch lookup
	usernames := make([]string, 0, len(sessions))
	for _, s := range sessions {
		usernames = append(usernames, s.Username)
	}

	nameMap := make(map[string]string)
	if len(usernames) > 0 {
		var users []struct {
			Username string
			Name     string
		}
		database.DB.Model(&models.PppoeUser{}).
			Select("username, name").
			Where("username IN ?", usernames).
			Find(&users)
		for _, u := range users {
			nameMap[u.Username] = u.Name
		}
	}

	now := time.Now()
	dtos := make([]SessionDTO, 0, len(sessions))
	for _, s := range sessions {
		var sessionTime int
		if s.AcctStartTime != nil {
			sessionTime = int(now.Sub(*s.AcctStartTime).Seconds())
		}

		startStr := ""
		if s.AcctStartTime != nil {
			startStr = s.AcctStartTime.Format(time.RFC3339)
		}

		var inputOctets, outputOctets int64
		if s.AcctInputOctets != nil { inputOctets = *s.AcctInputOctets }
		if s.AcctOutputOctets != nil { outputOctets = *s.AcctOutputOctets }

		customerName := s.Username
		if name, ok := nameMap[s.Username]; ok {
			customerName = name
		}

		dtos = append(dtos, SessionDTO{
			RadacctID:       s.RadacctID,
			Username:        s.Username,
			CustomerName:    customerName,
			NASIPAddress:    s.NASIPAddress,
			FramedIPAddress: s.FramedIPAddress,
			CallingStation:  s.CallingStationID,
			StartTime:       startStr,
			SessionTime:     sessionTime,
			InputOctets:     inputOctets,
			OutputOctets:    outputOctets,
			IsOnline:        true,
			Uptime:          formatDuration(sessionTime),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  dtos,
		"total": len(dtos),
	})
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "0s"
	}
	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60

	if d > 0 {
		return strconv.Itoa(d) + "d " + strconv.Itoa(h) + "h " + strconv.Itoa(m) + "m"
	}
	if h > 0 {
		return strconv.Itoa(h) + "h " + strconv.Itoa(m) + "m " + strconv.Itoa(s) + "s"
	}
	if m > 0 {
		return strconv.Itoa(m) + "m " + strconv.Itoa(s) + "s"
	}
	return strconv.Itoa(s) + "s"
}

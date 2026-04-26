package models

import "time"

// ============================================================================
// Router / NAS Model
// ============================================================================

// Router - MikroTik router / NAS (Network Access Server)
type Router struct {
	ID          string    `gorm:"primaryKey;size:191" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Nasname     string    `gorm:"not null;index" json:"nasname"`
	Shortname   string    `gorm:"not null;index" json:"shortname"`
	Type        string    `gorm:"default:mikrotik" json:"type"`
	IPAddress   string    `gorm:"column:ipAddress;not null" json:"ipAddress"`
	Username    string    `gorm:"not null" json:"username"`
	Password    string    `gorm:"not null" json:"password"`
	Port        int       `gorm:"default:8728" json:"port"`
	ApiPort     int       `gorm:"column:apiPort;default:8729" json:"apiPort"`
	Secret      string    `gorm:"default:secret123" json:"secret"`
	Ports       int       `gorm:"default:1812" json:"ports"`
	Server      *string   `json:"server"`
	Community   *string   `json:"community"`
	Description *string   `json:"description"`
	Latitude    *float64  `json:"latitude"`
	Longitude   *float64  `json:"longitude"`
	VpnClientID *string   `gorm:"column:vpnClientId;index" json:"vpnClientId"`
	IsActive    bool      `gorm:"column:isActive;default:true" json:"isActive"`
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (Router) TableName() string { return "nas" }

// ============================================================================
// Auth Models
// ============================================================================

// AdminUser - Admin/staff yang login ke panel admin
type AdminUser struct {
	ID               string     `gorm:"primaryKey;size:191" json:"id"`
	Username         string     `gorm:"uniqueIndex;not null" json:"username"`
	Email            *string    `gorm:"uniqueIndex" json:"email"`
	Password         string     `gorm:"not null" json:"-"`
	Name             string     `gorm:"not null" json:"name"`
	Role             string     `gorm:"default:CUSTOMER_SERVICE" json:"role"`
	IsActive         bool       `gorm:"column:isActive;default:true" json:"isActive"`
	Phone            *string    `json:"phone"`
	TwoFactorEnabled bool      `gorm:"column:twoFactorEnabled;default:false" json:"twoFactorEnabled"`
	TwoFactorSecret  *string    `gorm:"column:twoFactorSecret;size:255" json:"-"`
	LastLogin        *time.Time `gorm:"column:lastLogin" json:"lastLogin"`
	CreatedAt        time.Time  `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (AdminUser) TableName() string { return "admin_users" }

// User - Legacy user model (NextAuth)
type User struct {
	ID        string    `gorm:"primaryKey;size:191" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Name      string    `gorm:"not null" json:"name"`
	Role      string    `gorm:"default:ADMIN" json:"role"`
	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (User) TableName() string { return "users" }

// ============================================================================
// Company / Settings
// ============================================================================

// Company - Pengaturan perusahaan/ISP
type Company struct {
	ID                  string    `gorm:"primaryKey;size:191" json:"id"`
	Name                string    `gorm:"not null" json:"name"`
	Address             *string   `json:"address"`
	Phone               *string   `json:"phone"`
	Email               *string   `json:"email"`
	Logo                *string   `json:"logo"`
	AdminPhone          *string   `gorm:"column:adminPhone" json:"adminPhone"`
	BaseURL             *string   `gorm:"column:baseUrl;default:http://localhost:3000" json:"baseUrl"`
	Timezone            *string   `gorm:"default:Asia/Jakarta" json:"timezone"`
	PoweredBy           *string   `gorm:"column:poweredBy;default:SALFANET RADIUS" json:"poweredBy"`
	CustomerIDPrefix    *string   `gorm:"column:customerIdPrefix;size:10" json:"customerIdPrefix"`
	InvoiceGenerateDays *int      `gorm:"column:invoiceGenerateDays;default:7" json:"invoiceGenerateDays"`
	GracePeriodDays     *int      `gorm:"column:gracePeriodDays;default:0" json:"gracePeriodDays"`
	IsolationEnabled    *bool     `gorm:"column:isolationEnabled;default:true" json:"isolationEnabled"`
	IsolationIPPool     *string   `gorm:"column:isolationIpPool;size:100;default:192.168.200.0/24" json:"isolationIpPool"`
	IsolationRateLimit  *string   `gorm:"column:isolationRateLimit;size:50;default:64k/64k" json:"isolationRateLimit"`
	CreatedAt           time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (Company) TableName() string { return "companies" }

// ============================================================================
// Invoice & Payment
// ============================================================================

// Invoice - Tagihan pelanggan
type Invoice struct {
	ID               string     `gorm:"primaryKey;size:191" json:"id"`
	InvoiceNumber    string     `gorm:"column:invoiceNumber;uniqueIndex;not null" json:"invoiceNumber"`
	UserID           *string    `gorm:"column:userId;index" json:"userId"`
	Amount           int        `gorm:"not null" json:"amount"`
	Status           string     `gorm:"default:PENDING;index" json:"status"`
	DueDate          time.Time  `gorm:"column:dueDate;index" json:"dueDate"`
	PaidAt           *time.Time `gorm:"column:paidAt" json:"paidAt"`
	PaymentLink      *string    `gorm:"column:paymentLink" json:"paymentLink"`
	PaymentToken     *string    `gorm:"column:paymentToken;uniqueIndex" json:"paymentToken"`
	CustomerName     *string    `gorm:"column:customerName" json:"customerName"`
	CustomerPhone    *string    `gorm:"column:customerPhone" json:"customerPhone"`
	CustomerEmail    *string    `gorm:"column:customerEmail" json:"customerEmail"`
	CustomerUsername *string    `gorm:"column:customerUsername" json:"customerUsername"`
	Notes            *string    `gorm:"type:text" json:"notes"`
	InvoiceType      string     `gorm:"column:invoiceType;default:MONTHLY;index" json:"invoiceType"`
	BaseAmount       *int       `gorm:"column:baseAmount" json:"baseAmount"`
	CreatedAt        time.Time  `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	// Relations
	User *PppoeUser `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Invoice) TableName() string { return "invoices" }

// ============================================================================
// Session
// ============================================================================

// Session - Active PPPoE sessions
type Session struct {
	ID            string     `gorm:"primaryKey;size:191" json:"id"`
	Username      string     `gorm:"not null" json:"username"`
	UserID        *string    `gorm:"column:userId;index" json:"userId"`
	NasIPAddress  string     `gorm:"column:nasIpAddress;not null" json:"nasIpAddress"`
	SessionID     string     `gorm:"column:sessionId;uniqueIndex" json:"sessionId"`
	StartTime     time.Time  `gorm:"column:startTime;autoCreateTime" json:"startTime"`
	StopTime      *time.Time `gorm:"column:stopTime" json:"stopTime"`
	UploadBytes   int64      `gorm:"column:uploadBytes;default:0" json:"uploadBytes"`
	DownloadBytes int64      `gorm:"column:downloadBytes;default:0" json:"downloadBytes"`
	CreatedAt     time.Time  `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

func (Session) TableName() string { return "sessions" }

// ============================================================================
// Activity Log
// ============================================================================

// ActivityLog - Audit trail for admin actions
type ActivityLog struct {
	ID          string    `gorm:"primaryKey;size:191" json:"id"`
	UserID      *string   `gorm:"column:userId" json:"userId"`
	Username    string    `gorm:"not null" json:"username"`
	UserRole    *string   `gorm:"column:userRole" json:"userRole"`
	Action      string    `gorm:"not null" json:"action"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Module      string    `gorm:"not null;index" json:"module"`
	Status      string    `gorm:"default:success" json:"status"`
	IPAddress   *string   `gorm:"column:ipAddress" json:"ipAddress"`
	Metadata    *string   `gorm:"type:text" json:"metadata"`
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

func (ActivityLog) TableName() string { return "activity_logs" }

package models

import "time"

// ============================================================================
// PPPoE Models
// ============================================================================

// PppoeProfile - Paket layanan PPPoE (speed, harga, validity)
type PppoeProfile struct {
	ID                  string     `gorm:"primaryKey;size:191" json:"id"`
	Name                string     `gorm:"uniqueIndex;not null" json:"name"`
	Price               int        `gorm:"not null" json:"price"`
	DownloadSpeed       int        `gorm:"not null" json:"downloadSpeed"`
	UploadSpeed         int        `gorm:"not null" json:"uploadSpeed"`
	RateLimit           *string    `json:"rateLimit"`
	GroupName           string     `gorm:"not null;index" json:"groupName"`
	MikrotikProfileName *string   `json:"mikrotikProfileName"`
	IPPoolName          *string    `gorm:"column:ipPoolName" json:"ipPoolName"`
	IPPoolRange         *string    `gorm:"column:ipPoolRange" json:"ipPoolRange"`
	LocalAddress        *string    `json:"localAddress"`
	LastRouterID        *string    `gorm:"column:lastRouterId" json:"lastRouterId"`
	Description         *string    `json:"description"`
	Hpp                 *int       `json:"hpp"`
	PpnActive           bool       `gorm:"default:false" json:"ppnActive"`
	PpnRate             int        `gorm:"default:11" json:"ppnRate"`
	IsActive            bool       `gorm:"default:true" json:"isActive"`
	LastSyncAt          *time.Time `json:"lastSyncAt"`
	SyncedToRadius      bool       `gorm:"default:false" json:"syncedToRadius"`
	ValidityUnit        string     `gorm:"default:MONTHS" json:"validityUnit"`
	ValidityValue       int        `gorm:"default:1" json:"validityValue"`
	SharedUser          bool       `gorm:"default:true" json:"sharedUser"`
	CreatedAt           time.Time  `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	// Relations
	Users []PppoeUser `gorm:"foreignKey:ProfileID" json:"users,omitempty"`
}

func (PppoeProfile) TableName() string { return "pppoe_profiles" }

// PppoeArea - Wilayah/area pelanggan
type PppoeArea struct {
	ID          string    `gorm:"primaryKey;size:191" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description *string   `json:"description"`
	IsActive    bool      `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (PppoeArea) TableName() string { return "pppoe_areas" }

// PppoeCustomer - Data pelanggan (bisa punya banyak PPPoE user)
type PppoeCustomer struct {
	ID           string    `gorm:"primaryKey;size:191" json:"id"`
	CustomerID   string    `gorm:"column:customerId;uniqueIndex;size:10" json:"customerId"`
	Name         string    `gorm:"not null" json:"name"`
	Phone        string    `gorm:"not null;index" json:"phone"`
	Email        *string   `json:"email"`
	Address      *string   `gorm:"type:text" json:"address"`
	IDCardNumber *string   `gorm:"column:idCardNumber;size:50" json:"idCardNumber"`
	IDCardPhoto  *string   `gorm:"column:idCardPhoto;size:500" json:"idCardPhoto"`
	IsActive     bool      `gorm:"default:true;index" json:"isActive"`
	AreaID       *string   `gorm:"column:areaId;index" json:"areaId"`
	CreatedAt    time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	// Relations
	Area       *PppoeArea  `gorm:"foreignKey:AreaID" json:"area,omitempty"`
	PppoeUsers []PppoeUser `gorm:"foreignKey:PppoeCustomerID" json:"pppoeUsers,omitempty"`
}

func (PppoeCustomer) TableName() string { return "pppoe_customers" }

// PppoeUser - Akun PPPoE (username/password untuk koneksi internet)
type PppoeUser struct {
	ID                   string     `gorm:"primaryKey;size:191" json:"id"`
	Username             string     `gorm:"uniqueIndex;not null" json:"username"`
	CustomerID           *string    `gorm:"column:customer_id;uniqueIndex;size:20" json:"customerId"`
	PppoeCustomerID      *string    `gorm:"column:pppoe_customer_id;index" json:"pppoeCustomerId"`
	Password             string     `gorm:"not null" json:"password"`
	ProfileID            string     `gorm:"column:profileId;not null;index" json:"profileId"`
	AreaID               *string    `gorm:"column:areaId;index" json:"areaId"`
	Status               string     `gorm:"default:active;index" json:"status"`
	IPAddress            *string    `gorm:"column:ipAddress" json:"ipAddress"`
	MacAddress           *string    `gorm:"column:macAddress" json:"macAddress"`
	Comment              *string    `json:"comment"`
	ExpiredAt            *time.Time `gorm:"column:expiredAt;index" json:"expiredAt"`
	Address              *string    `json:"address"`
	Latitude             *float64   `json:"latitude"`
	Longitude            *float64   `json:"longitude"`
	Email                *string    `json:"email"`
	LastSyncAt           *time.Time `gorm:"column:lastSyncAt" json:"lastSyncAt"`
	Name                 string     `gorm:"not null" json:"name"`
	Phone                string     `gorm:"not null;index" json:"phone"`
	SyncedToRadius       bool       `gorm:"column:syncedToRadius;default:false" json:"syncedToRadius"`
	RouterID             *string    `gorm:"column:routerId;index" json:"routerId"`
	SubscriptionType     string     `gorm:"column:subscriptionType;default:POSTPAID;index" json:"subscriptionType"`
	LastPaymentDate      *time.Time `gorm:"column:lastPaymentDate" json:"lastPaymentDate"`
	BillingDay           *int       `gorm:"column:billingDay;default:1;index" json:"billingDay"`
	AutoIsolationEnabled bool       `gorm:"column:autoIsolationEnabled;default:true" json:"autoIsolationEnabled"`
	Balance              int        `gorm:"default:0" json:"balance"`
	AutoRenewal          bool       `gorm:"column:autoRenewal;default:false;index" json:"autoRenewal"`
	ConnectionType       string     `gorm:"column:connectionType;default:PPPOE" json:"connectionType"`
	IDCardNumber         *string    `gorm:"column:idCardNumber;size:50" json:"idCardNumber"`
	IDCardPhoto          *string    `gorm:"column:idCardPhoto;size:500" json:"idCardPhoto"`
	InstallationPhotos   *string    `gorm:"column:installationPhotos;type:json" json:"installationPhotos"`
	FollowRoad           bool       `gorm:"column:followRoad;default:false" json:"followRoad"`
	ReferralCode         *string    `gorm:"column:referralCode;uniqueIndex;size:10" json:"referralCode"`
	ReferredByID         *string    `gorm:"column:referred_by_id;index" json:"referredById"`
	CreatedAt            time.Time  `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt            time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	// Relations
	Profile       *PppoeProfile  `gorm:"foreignKey:ProfileID" json:"profile,omitempty"`
	Area          *PppoeArea     `gorm:"foreignKey:AreaID" json:"area,omitempty"`
	Router        *Router        `gorm:"foreignKey:RouterID" json:"router,omitempty"`
	PppoeCustomer *PppoeCustomer `gorm:"foreignKey:PppoeCustomerID" json:"pppoeCustomer,omitempty"`
}

func (PppoeUser) TableName() string { return "pppoe_users" }

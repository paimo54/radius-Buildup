package models

import "time"

// ============================================================================
// Hotspot Profile
// ============================================================================

// HotspotProfile - Paket layanan Hotspot (voucher)
type HotspotProfile struct {
	ID                string    `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"uniqueIndex;not null" json:"name"`
	Speed             string    `gorm:"not null" json:"speed"`
	CostPrice         int       `gorm:"not null" json:"costPrice"`
	SellingPrice      int       `gorm:"not null" json:"sellingPrice"`
	SharedUsers       int       `gorm:"default:1" json:"sharedUsers"`
	ValidityUnit      string    `gorm:"not null" json:"validityUnit"` // MINUTES, HOURS, DAYS, MONTHS
	ValidityValue     int       `gorm:"not null" json:"validityValue"`
	GroupProfile      *string   `json:"groupProfile"`
	IsActive          bool      `gorm:"default:true" json:"isActive"`
	AgentAccess       bool      `gorm:"default:true" json:"agentAccess"`
	EVoucherAccess    bool      `gorm:"column:eVoucherAccess;default:true" json:"eVoucherAccess"`
	ResellerFee       int       `gorm:"default:0" json:"resellerFee"`
	UsageQuota        *int64    `gorm:"column:usageQuota" json:"usageQuota"`
	UsageDuration     *int      `gorm:"column:usageDuration" json:"usageDuration"`
	UsageDurationUnit string    `gorm:"column:usageDurationUnit;default:HOURS" json:"usageDurationUnit"`
	MikrotikProfileName *string `gorm:"column:mikrotikProfileName" json:"mikrotikProfileName"`
	IpPoolName        *string   `gorm:"column:ipPoolName" json:"ipPoolName"`
	CreatedAt         time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	// Relations
	Vouchers []HotspotVoucher `gorm:"foreignKey:ProfileID" json:"vouchers,omitempty"`
}

func (HotspotProfile) TableName() string { return "hotspot_profiles" }

// ============================================================================
// Hotspot Voucher
// ============================================================================

// HotspotVoucher - Voucher hotspot individual
type HotspotVoucher struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	Code         string     `gorm:"uniqueIndex;not null" json:"code"`
	Password     *string    `json:"password"`
	ProfileID    string     `gorm:"column:profileId;index" json:"profileId"`
	RouterID     *string    `gorm:"column:routerId;index" json:"routerId"`
	AgentID      *string    `gorm:"column:agentId;index" json:"agentId"`
	VoucherType  string     `gorm:"column:voucherType;default:same" json:"voucherType"`
	CodeType     string     `gorm:"column:codeType;default:alphanumeric" json:"codeType"`
	Status       string     `gorm:"default:WAITING;index" json:"status"`
	BatchCode    *string    `gorm:"column:batchCode;index" json:"batchCode"`
	ExpiresAt    *time.Time `gorm:"column:expiresAt" json:"expiresAt"`
	FirstLoginAt *time.Time `gorm:"column:firstLoginAt" json:"firstLoginAt"`
	LastUsedBy   *string    `gorm:"column:lastUsedBy" json:"lastUsedBy"`
	OrderID      *string    `gorm:"column:orderId;index" json:"orderId"`
	CreatedAt    time.Time  `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	// Relations
	Profile *HotspotProfile `gorm:"foreignKey:ProfileID" json:"profile,omitempty"`
	Router  *Router         `gorm:"foreignKey:RouterID" json:"router,omitempty"`
}

func (HotspotVoucher) TableName() string { return "hotspot_vouchers" }

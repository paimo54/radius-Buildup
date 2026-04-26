package models

import "time"

// ============================================================================
// OLT Devices
// ============================================================================

type Olt struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	Port           int       `gorm:"default:161" json:"port"`
	Community      string    `gorm:"default:public" json:"community"`
	TelnetPort     int       `gorm:"default:23" json:"telnetPort"`
	TelnetUser     string    `json:"telnetUser"`
	TelnetPass     string    `json:"telnetPass"`
	TelnetEnable   string    `json:"telnetEnable"`
	Vendor         string    `json:"vendor"`
	Description    *string   `gorm:"type:text" json:"description"`
	IsActive       bool      `gorm:"column:isActive;default:true" json:"isActive"`
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (Olt) TableName() string { return "olts" }

// ============================================================================
// OID Mappings (SNMP)
// ============================================================================

type OidMapping struct {
	ID        int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Vendor    string `gorm:"uniqueIndex;size:50" json:"vendor"`
	OidName   string `gorm:"column:oid_name" json:"oid_name"`
	OidTx     string `gorm:"column:oid_tx" json:"oid_tx"`
	OidRx     string `gorm:"column:oid_rx" json:"oid_rx"`
	OidStatus string `gorm:"column:oid_status" json:"oid_status"`
}

func (OidMapping) TableName() string { return "oid_mappings" }

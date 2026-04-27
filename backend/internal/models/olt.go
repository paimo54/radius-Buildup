package models

import "time"

// ============================================================================
// OLT Devices
// ============================================================================

type Olt struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Host          string    `gorm:"column:ipAddress" json:"host"`
	Port          int       `gorm:"column:snmpPort;default:161" json:"port"`
	Community     string    `gorm:"column:snmpCommunity;default:public" json:"community"`
	TelnetPort    int       `gorm:"column:telnetPort;default:23" json:"telnetPort"`
	TelnetUser    string    `gorm:"column:username" json:"telnetUser"`
	TelnetPass    string    `gorm:"column:password" json:"telnetPass"`
	TelnetEnable  bool      `gorm:"column:telnetEnabled" json:"telnetEnable"`
	SSHPort       int       `gorm:"column:sshPort;default:22" json:"sshPort"`
	Vendor        string    `json:"vendor"`
	Description *string   `gorm:"type:text" json:"description"`
	Status      string    `gorm:"column:status;default:active" json:"status"` // Next.js 'status' is a string 'active'/'inactive'
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (Olt) TableName() string { return "network_olts" }

// ============================================================================
// OID Mappings (SNMP)
// ============================================================================

type OidMapping struct {
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Vendor      string `gorm:"uniqueIndex;size:50" json:"vendor"`
	OidName     string `gorm:"column:oid_name" json:"oid_name"`
	OidTx       string `gorm:"column:oid_tx" json:"oid_tx"`
	OidRx       string `gorm:"column:oid_rx" json:"oid_rx"`
	OidStatus   string `gorm:"column:oid_status" json:"oid_status"`
	OidSerial   string `gorm:"column:oid_serial" json:"oid_serial"`
	OidModel    string `gorm:"column:oid_model" json:"oid_model"`
	OidDistance  string `gorm:"column:oid_distance" json:"oid_distance"`
}

func (OidMapping) TableName() string { return "oid_mappings" }

// ============================================================================
// OLT ONUs (Cached SNMP scan results)
// ============================================================================

type OltOnu struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OltID     string    `gorm:"index;size:255" json:"oltId"`
	OnuIndex  string    `gorm:"size:100" json:"index"`
	Name      string    `gorm:"size:255" json:"name"`
	Serial    string    `gorm:"size:100" json:"serial"`
	Model     string    `gorm:"size:100" json:"model"`
	Status    string    `gorm:"size:50;default:Unknown" json:"status"`
	Tx        float64   `gorm:"type:decimal(10,2)" json:"tx"`
	Rx        float64   `gorm:"type:decimal(10,2)" json:"rx"`
	Distance  float64   `gorm:"type:decimal(10,2)" json:"distance"`
	ScannedAt time.Time `gorm:"autoCreateTime" json:"scannedAt"`
}

func (OltOnu) TableName() string { return "olt_onus" }

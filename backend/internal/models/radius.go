package models

import "time"

// ============================================================================
// RADIUS Tables (FreeRADIUS standard tables)
// ============================================================================

// Radcheck - FreeRADIUS check attributes (username/password)
type Radcheck struct {
	ID        int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string `gorm:"size:64;not null;uniqueIndex:idx_username_attribute" json:"username"`
	Attribute string `gorm:"size:64;not null;uniqueIndex:idx_username_attribute" json:"attribute"`
	Op        string `gorm:"size:2;default::=" json:"op"`
	Value     string `gorm:"size:253;not null" json:"value"`
}

func (Radcheck) TableName() string { return "radcheck" }

// Radreply - FreeRADIUS reply attributes
type Radreply struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string `gorm:"size:64;not null;index" json:"username"`
	Attribute string `gorm:"size:64;not null" json:"attribute"`
	Op        string `gorm:"size:2;default:=" json:"op"`
	Value     string `gorm:"size:253;not null" json:"value"`
}

func (Radreply) TableName() string { return "radreply" }

// Radusergroup - FreeRADIUS user-group mapping
type Radusergroup struct {
	ID        int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string `gorm:"size:64;not null;uniqueIndex:idx_username_groupname;index" json:"username"`
	Groupname string `gorm:"size:64;not null;uniqueIndex:idx_username_groupname;index" json:"groupname"`
	Priority  int    `gorm:"default:1" json:"priority"`
}

func (Radusergroup) TableName() string { return "radusergroup" }

// Radgroupcheck - FreeRADIUS group check attributes
type Radgroupcheck struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Groupname string `gorm:"size:64;not null;index" json:"groupname"`
	Attribute string `gorm:"size:64;not null" json:"attribute"`
	Op        string `gorm:"size:2;default:==" json:"op"`
	Value     string `gorm:"size:253;not null" json:"value"`
}

func (Radgroupcheck) TableName() string { return "radgroupcheck" }

// Radgroupreply - FreeRADIUS group reply attributes (rate-limit, etc)
type Radgroupreply struct {
	ID        int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Groupname string `gorm:"size:64;not null;index" json:"groupname"`
	Attribute string `gorm:"size:64;not null" json:"attribute"`
	Op        string `gorm:"size:2;default::=" json:"op"`
	Value     string `gorm:"size:253;not null" json:"value"`
}

func (Radgroupreply) TableName() string { return "radgroupreply" }

// Radpostauth - FreeRADIUS post-auth log
type Radpostauth struct {
	ID       int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string    `gorm:"size:64;not null;index" json:"username"`
	Pass     string    `gorm:"size:64;not null" json:"pass"`
	Reply    string    `gorm:"size:32;not null" json:"reply"`
	Authdate time.Time `gorm:"autoCreateTime" json:"authdate"`
}

func (Radpostauth) TableName() string { return "radpostauth" }

// Radacct - FreeRADIUS accounting (sessions, traffic)
type Radacct struct {
	RadacctID           int64      `gorm:"primaryKey;autoIncrement;column:radacctid" json:"radacctid"`
	AcctSessionID       string     `gorm:"column:acctsessionid;size:64;not null;index" json:"acctsessionid"`
	AcctUniqueID        string     `gorm:"column:acctuniqueid;size:32;uniqueIndex" json:"acctuniqueid"`
	Username            string     `gorm:"size:64;not null;index" json:"username"`
	Groupname           string     `gorm:"size:64;default:''" json:"groupname"`
	Realm               *string    `gorm:"size:64" json:"realm"`
	NASIPAddress        string     `gorm:"column:nasipaddress;size:15;not null;index" json:"nasipaddress"`
	NASPortID           *string    `gorm:"column:nasportid;size:32" json:"nasportid"`
	NASPortType         *string    `gorm:"column:nasporttype;size:32" json:"nasporttype"`
	AcctStartTime       *time.Time `gorm:"column:acctstarttime;index" json:"acctstarttime"`
	AcctUpdateTime      *time.Time `gorm:"column:acctupdatetime" json:"acctupdatetime"`
	AcctStopTime        *time.Time `gorm:"column:acctstoptime;index" json:"acctstoptime"`
	AcctInterval        *int       `gorm:"column:acctinterval;index" json:"acctinterval"`
	AcctSessionTime     *uint      `gorm:"column:acctsessiontime;index" json:"acctsessiontime"`
	AcctAuthentic       *string    `gorm:"column:acctauthentic;size:32" json:"acctauthentic"`
	ConnectInfoStart    *string    `gorm:"column:connectinfo_start;size:50" json:"connectinfo_start"`
	ConnectInfoStop     *string    `gorm:"column:connectinfo_stop;size:50" json:"connectinfo_stop"`
	AcctInputOctets     *int64     `gorm:"column:acctinputoctets" json:"acctinputoctets"`
	AcctOutputOctets    *int64     `gorm:"column:acctoutputoctets" json:"acctoutputoctets"`
	CalledStationID     string     `gorm:"column:calledstationid;size:50;not null" json:"calledstationid"`
	CallingStationID    string     `gorm:"column:callingstationid;size:50;not null" json:"callingstationid"`
	AcctTerminateCause  string     `gorm:"column:acctterminatecause;size:32;not null" json:"acctterminatecause"`
	ServiceType         *string    `gorm:"column:servicetype;size:32" json:"servicetype"`
	FramedProtocol      *string    `gorm:"column:framedprotocol;size:32" json:"framedprotocol"`
	FramedIPAddress     string     `gorm:"column:framedipaddress;size:15;not null;index" json:"framedipaddress"`
	FramedIPv6Address   string     `gorm:"column:framedipv6address;size:45;default:''" json:"framedipv6address"`
	FramedIPv6Prefix    string     `gorm:"column:framedipv6prefix;size:45;default:''" json:"framedipv6prefix"`
	FramedInterfaceID   string     `gorm:"column:framedinterfaceid;size:44;default:''" json:"framedinterfaceid"`
	DelegatedIPv6Prefix string     `gorm:"column:delegatedipv6prefix;size:45;default:''" json:"delegatedipv6prefix"`
}

func (Radacct) TableName() string { return "radacct" }

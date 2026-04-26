package models

import "time"

// ============================================================================
// WhatsApp Providers
// ============================================================================

// WhatsAppProvider stores API configs for WA Gateway providers (Wablas, Fonnte, etc)
type WhatsAppProvider struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // e.g. 'wablas', 'fonnte'
	APIKey       string    `gorm:"column:apiKey" json:"apiKey"`
	APIUrl       string    `gorm:"column:apiUrl" json:"apiUrl"`
	SenderNumber *string   `gorm:"column:senderNumber" json:"senderNumber"`
	IsActive     bool      `gorm:"column:isActive;default:true" json:"isActive"`
	Priority     int       `gorm:"default:0" json:"priority"`
	Description  *string   `gorm:"type:text" json:"description"`
	CreatedAt    time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (WhatsAppProvider) TableName() string { return "whatsapp_providers" }

// ============================================================================
// WhatsApp Templates
// ============================================================================

// WhatsAppTemplate stores message templates for different events
type WhatsAppTemplate struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Type      string    `gorm:"uniqueIndex" json:"type"` // e.g. 'invoice_created', 'payment_success'
	Message   string    `gorm:"type:text" json:"message"`
	IsActive  bool      `gorm:"column:isActive;default:true" json:"isActive"`
	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (WhatsAppTemplate) TableName() string { return "whatsapp_templates" }

// ============================================================================
// WhatsApp History
// ============================================================================

// WhatsAppHistory logs sent messages
type WhatsAppHistory struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Phone        string    `json:"phone"`
	Message      string    `gorm:"type:text" json:"message"`
	Status       string    `json:"status"` // 'PENDING', 'SENT', 'FAILED'
	Response     *string   `gorm:"type:text" json:"response"`
	ProviderName *string   `gorm:"column:providerName" json:"providerName"`
	ProviderType *string   `gorm:"column:providerType" json:"providerType"`
	SentAt       time.Time `gorm:"column:sentAt;autoCreateTime" json:"sentAt"`
}

func (WhatsAppHistory) TableName() string { return "whatsapp_history" }

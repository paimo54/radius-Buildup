package models

import "time"

// ============================================================================
// Payment Gateway
// ============================================================================

// PaymentGateway configuration (Midtrans, Xendit, Tripay, Duitku)
type PaymentGateway struct {
	ID                  string    `gorm:"primaryKey" json:"id"`
	Name                string    `json:"name"`
	Provider            string    `gorm:"uniqueIndex" json:"provider"` // midtrans, xendit, tripay, duitku
	IsActive            bool      `gorm:"default:false" json:"isActive"`
	
	// Midtrans
	MidtransEnvironment string    `gorm:"column:midtransEnvironment;default:sandbox" json:"midtransEnvironment"`
	MidtransServerKey   *string   `gorm:"column:midtransServerKey" json:"midtransServerKey"`
	MidtransClientKey   *string   `gorm:"column:midtransClientKey" json:"midtransClientKey"`
	
	// Xendit
	XenditEnvironment   string    `gorm:"column:xenditEnvironment;default:sandbox" json:"xenditEnvironment"`
	XenditApiKey        *string   `gorm:"column:xenditApiKey" json:"xenditApiKey"`
	XenditWebhookToken  *string   `gorm:"column:xenditWebhookToken" json:"xenditWebhookToken"`
	
	// Duitku
	DuitkuEnvironment   string    `gorm:"column:duitkuEnvironment;default:sandbox" json:"duitkuEnvironment"`
	DuitkuMerchantCode  *string   `gorm:"column:duitkuMerchantCode" json:"duitkuMerchantCode"`
	DuitkuApiKey        *string   `gorm:"column:duitkuApiKey" json:"duitkuApiKey"`
	
	// Tripay
	TripayEnvironment   string    `gorm:"column:tripayEnvironment;default:sandbox" json:"tripayEnvironment"`
	TripayMerchantCode  *string   `gorm:"column:tripayMerchantCode" json:"tripayMerchantCode"`
	TripayApiKey        *string   `gorm:"column:tripayApiKey" json:"tripayApiKey"`
	TripayPrivateKey    *string   `gorm:"column:tripayPrivateKey" json:"tripayPrivateKey"`

	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (PaymentGateway) TableName() string { return "payment_gateways" }

// ============================================================================
// Webhook Logs
// ============================================================================

type WebhookLog struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	Gateway       string    `json:"gateway"`
	GatewayID     *string   `gorm:"column:gatewayId" json:"gatewayId"`
	OrderID       string    `gorm:"column:orderId;index" json:"orderId"` // e.g. InvoiceNumber
	Status        string    `json:"status"`
	TransactionID *string   `gorm:"column:transactionId" json:"transactionId"`
	Amount        *int      `json:"amount"`
	Payload       *string   `gorm:"type:text" json:"payload"`
	Response      *string   `gorm:"type:text" json:"response"`
	Success       bool      `gorm:"default:true" json:"success"`
	ErrorMessage  *string   `gorm:"column:errorMessage;type:text" json:"errorMessage"`
	CreatedAt     time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

func (WebhookLog) TableName() string { return "webhook_logs" }

// ============================================================================
// Payment
// ============================================================================

type Payment struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	InvoiceID string    `gorm:"column:invoiceId;index" json:"invoiceId"`
	Amount    int       `json:"amount"`
	Method    string    `json:"method"`
	GatewayID *string   `gorm:"column:gatewayId" json:"gatewayId"`
	Status    string    `json:"status"`
	Notes     *string   `gorm:"type:text" json:"notes"`
	PaidAt    time.Time `gorm:"column:paidAt;autoCreateTime" json:"paidAt"`
	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`

	// Relations
	Invoice *Invoice `gorm:"foreignKey:InvoiceID" json:"invoice,omitempty"`
}

func (Payment) TableName() string { return "payments" }

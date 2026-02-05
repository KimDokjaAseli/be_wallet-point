package wallet

import "time"

type PaymentToken struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Token        string    `json:"token" gorm:"uniqueIndex;not null"`
	QRCodeBase64 string    `json:"qr_code_base64" gorm:"type:text"`
	Amount       int       `json:"amount" gorm:"not null"`
	Merchant     string    `json:"merchant" gorm:"size:255"` // Added for UX
	Expiry       time.Time `json:"expiry" gorm:"not null"`
	WalletID     uint      `json:"wallet_id" gorm:"not null"` // Creator
	RecipientID  uint      `json:"recipient_id"`              // Who gets the money
	Status       string    `json:"status" gorm:"type:enum('active','consumed','expired');default:'active'"`
	Type         string    `json:"type" gorm:"size:50"` // "purchase" or "transfer"
	ProductID    uint      `json:"product_id"`          // For marketplace purchases
	CreatedAt    time.Time `json:"created_at"`
}

func (PaymentToken) TableName() string {
	return "payment_tokens"
}

type PaymentTokenRequest struct {
	Amount      int    `json:"amount" binding:"required,gt=0"`
	Merchant    string `json:"merchant"` // Optional display name
	Type        string `json:"type" binding:"required,oneof=purchase transfer"`
	RecipientID uint   `json:"recipient_id"`
	ProductID   uint   `json:"product_id"`
}

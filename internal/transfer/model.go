package transfer

import (
	"time"
)

// Transfer represents a point transfer between users, mapped to 'transfers' table
type Transfer struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SenderID    uint      `json:"sender_id" gorm:"not null;index"`
	ReceiverID  uint      `json:"receiver_id" gorm:"not null;index"`
	Amount      int       `json:"amount" gorm:"not null"`
	Description string    `json:"description" gorm:"size:255"`
	Status      string    `json:"status" gorm:"type:enum('success','failed');default:'success'"`
	CreatedAt   time.Time `json:"created_at"`

	// Virtual fields for response
	SenderName   string `json:"sender_name,omitempty" gorm:"-"`
	ReceiverName string `json:"receiver_name,omitempty" gorm:"-"`
	SenderNIM    string `json:"sender_nim,omitempty" gorm:"-"`
	ReceiverNIM  string `json:"receiver_nim,omitempty" gorm:"-"`
}

func (Transfer) TableName() string {
	return "transfers"
}

// TransferInfo is kept as a DTO for compatibility if needed, or simply use Transfer
type TransferInfo = Transfer

// TransferRequest represents the request body for creating a transfer
type TransferRequest struct {
	ReceiverUserID uint   `json:"receiver_user_id" binding:"required"`
	Amount         int    `json:"amount" binding:"required,gt=0"`
	Description    string `json:"description" binding:"max=255"`
	PIN            string `json:"pin"`
}

// RecipientSummary represents simplified user info for transfer verification
type RecipientSummary struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	NIM      string `json:"nim,omitempty"`
}

package transfer

import (
	"time"
)

// TransferResponseDTO is a DTO for transfer info, not mapped to DB table anymore
type TransferInfo struct {
	ID               uint      `json:"id"`
	SenderWalletID   uint      `json:"sender_wallet_id"`
	ReceiverWalletID uint      `json:"receiver_wallet_id"`
	Amount           int       `json:"amount"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"created_at"`

	// Virtual fields for response
	SenderName   string `json:"sender_name,omitempty"`
	ReceiverName string `json:"receiver_name,omitempty"`
	SenderNIM    string `json:"sender_nim,omitempty"`
	ReceiverNIM  string `json:"receiver_nim,omitempty"`
}

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

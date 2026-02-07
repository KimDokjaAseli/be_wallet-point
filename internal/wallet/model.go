package wallet

import (
	"time"
)

type Wallet struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	Balance   int       `json:"balance" gorm:"default:0;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Wallet) TableName() string {
	return "wallets"
}

type WalletTransaction struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	WalletID    uint      `json:"wallet_id" gorm:"not null"`
	Type        string    `json:"type" gorm:"type:enum('mission','transfer_in','transfer_out','marketplace','adjustment','topup');not null"`
	Amount      int       `json:"amount" gorm:"not null"`
	Direction   string    `json:"direction" gorm:"type:enum('credit','debit');not null"`
	ReferenceID *uint     `json:"reference_id"`
	Status      string    `json:"status" gorm:"type:enum('success','failed','pending');default:'success'"`
	Description string    `json:"description" gorm:"size:500"`
	CreatedBy   string    `json:"created_by" gorm:"type:enum('system','admin','dosen');default:'system'"`
	CreatedAt   time.Time `json:"created_at"`
}

func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}

type WalletWithUser struct {
	WalletID uint   `json:"wallet_id"`
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	NimNip   string `json:"nim_nip"`
	Role     string `json:"role"`
	Balance  int    `json:"balance"`
}

type TransactionWithDetails struct {
	ID          uint      `json:"id"`
	WalletID    uint      `json:"wallet_id"`
	UserEmail   string    `json:"user_email"`
	UserName    string    `json:"user_name"`
	NimNip      string    `json:"nim_nip"`
	Type        string    `json:"type"`
	Amount      int       `json:"amount"`
	Direction   string    `json:"direction"`
	ReferenceID *uint     `json:"reference_id"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type AdjustmentRequest struct {
	WalletID    uint   `json:"wallet_id" binding:"required"`
	Amount      int    `json:"amount" binding:"required,gt=0"`
	Direction   string `json:"direction" binding:"required,oneof=credit debit"`
	Description string `json:"description" binding:"required"`
}

type ResetWalletRequest struct {
	WalletID   uint   `json:"wallet_id" binding:"required"`
	NewBalance int    `json:"new_balance" binding:"gte=0"`
	Reason     string `json:"reason" binding:"required"`
}

type TransactionListParams struct {
	Type      string
	Status    string
	Direction string
	FromDate  string
	ToDate    string
	Page      int
	Limit     int
}

type TransactionListResponse struct {
	Transactions []TransactionWithDetails `json:"transactions"`
	Total        int64                    `json:"total"`
	Page         int                      `json:"page"`
	Limit        int                      `json:"limit"`
	TotalPages   int                      `json:"total_pages"`
}

type AdminStats struct {
	TotalUsers        int64 `json:"total_users"`
	ActiveUsers       int64 `json:"active_users"`
	CirculationPoints int64 `json:"circulation_points"`
	TodayTransactions int64 `json:"today_transactions"`
	TodayCredits      int64 `json:"today_credits"`
	TodayDebits       int64 `json:"today_debits"`

	// Additional Monitoring Fields
	TotalProducts     int64 `json:"total_products"`
	TotalMissions     int64 `json:"total_missions"`
	PendingSubmission int64 `json:"pending_submissions"`
}

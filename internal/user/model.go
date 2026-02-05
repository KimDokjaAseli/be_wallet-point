package user

import (
	"time"
)

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;not null"`
	FullName     string    `json:"full_name" gorm:"not null"`
	NimNip       string    `json:"nim_nip" gorm:"uniqueIndex;not null"`
	Role         string    `json:"role" gorm:"type:enum('admin','dosen','mahasiswa');not null"`
	Status       string    `json:"status" gorm:"type:enum('active','inactive','suspended');default:'active'"`
	PinHash      string    `json:"-" gorm:"column:pin_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

type Wallet struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id" gorm:"uniqueIndex;not null"`
	Balance    int        `json:"balance" gorm:"default:0;not null"`
	LastSyncAt *time.Time `json:"last_sync_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (Wallet) TableName() string {
	return "wallets"
}

type UserWithWallet struct {
	ID         uint       `json:"id"`
	WalletID   uint       `json:"wallet_id"`
	Email      string     `json:"email"`
	FullName   string     `json:"full_name"`
	NimNip     string     `json:"nim_nip"`
	Role       string     `json:"role"`
	Status     string     `json:"status"`
	Balance    int        `json:"balance"`
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name,omitempty"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	Status   string `json:"status,omitempty" binding:"omitempty,oneof=active inactive suspended"`
	Role     string `json:"role,omitempty" binding:"omitempty,oneof=admin dosen mahasiswa"`
}

type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UserListParams struct {
	Role   string
	Status string
	Page   int
	Limit  int
}

type UserListResponse struct {
	Users      []UserWithWallet `json:"users"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

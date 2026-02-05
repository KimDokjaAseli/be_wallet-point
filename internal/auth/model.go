package auth

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

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	NimNip   string `json:"nim_nip" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin dosen mahasiswa"`
}

type PublicRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	NimNip   string `json:"nim_nip" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  UserSummary `json:"user"`
}

type UserSummary struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	NimNip   string `json:"nim_nip"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Balance  int    `json:"balance,omitempty"`
}
type UpdateProfileRequest struct {
	FullName string `json:"full_name" binding:"required"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UpdatePinRequest struct {
	OldPin string `json:"old_pin"`
	NewPin string `json:"new_pin" binding:"required,len=6"`
}

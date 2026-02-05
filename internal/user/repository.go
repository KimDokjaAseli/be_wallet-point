package user

import (
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetAllWithWallets gets all users with their wallet information
func (r *UserRepository) GetAllWithWallets(params UserListParams) ([]UserWithWallet, int64, error) {
	var users []UserWithWallet
	var total int64

	query := r.db.Table("users").
		Select("users.*, wallets.id as wallet_id, COALESCE(wallets.balance, 0) as balance, wallets.last_sync_at").
		Joins("LEFT JOIN wallets ON users.id = wallets.user_id")

	// Apply filters
	if params.Role != "" {
		query = query.Where("users.role = ?", params.Role)
	}
	if params.Status != "" {
		query = query.Where("users.status = ?", params.Status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	query = query.Limit(params.Limit).Offset(offset).Order("users.created_at DESC")

	if err := query.Scan(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// FindByID finds user by ID
func (r *UserRepository) FindByID(userID uint) (*User, error) {
	var user User
	err := r.db.First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByIDWithWallet finds user by ID with wallet
func (r *UserRepository) FindByIDWithWallet(userID uint) (*UserWithWallet, error) {
	var user UserWithWallet
	err := r.db.Table("users").
		Select("users.*, wallets.id as wallet_id, COALESCE(wallets.balance, 0) as balance, wallets.last_sync_at").
		Joins("LEFT JOIN wallets ON users.id = wallets.user_id").
		Where("users.id = ?", userID).
		Scan(&user).Error

	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

// Update updates user information
func (r *UserRepository) Update(userID uint, updates map[string]interface{}) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Updates(updates).Error
}

// Delete soft deletes user by setting status to inactive
func (r *UserRepository) Delete(userID uint) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Update("status", "inactive").Error
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(userID uint, hashedPassword string) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Update("password_hash", hashedPassword).Error
}

// CheckEmailExists checks if email exists (excluding current user)
func (r *UserRepository) CheckEmailExists(email string, excludeUserID uint) (bool, error) {
	var count int64
	query := r.db.Model(&User{}).Where("email = ?", email)
	if excludeUserID > 0 {
		query = query.Where("id != ?", excludeUserID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

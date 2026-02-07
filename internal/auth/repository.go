package auth

import (
	"errors"

	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// FindByEmail finds user by email
func (r *AuthRepository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByID finds user by ID
func (r *AuthRepository) FindByID(userID uint) (*User, error) {
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

// Create creates a new user
func (r *AuthRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// CheckEmailExists checks if email already exists
func (r *AuthRepository) CheckEmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// CheckNimNipExists checks if NIM/NIP already exists
func (r *AuthRepository) CheckNimNipExists(nimNip string) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).Where("nim_nip = ?", nimNip).Count(&count).Error
	return count > 0, err
}
func (r *AuthRepository) Update(userID uint, updates map[string]interface{}) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *AuthRepository) UpdatePassword(userID uint, newPassword string) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Update("password_hash", newPassword).Error
}

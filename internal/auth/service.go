package auth

import (
	"errors"
	"fmt"
	"wallet-point/utils"
)

type AuthService struct {
	repo      *AuthRepository
	jwtExpiry int
}

func NewAuthService(repo *AuthRepository, jwtExpiry int) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtExpiry: jwtExpiry,
	}
}

// Login authenticates user and returns JWT token
func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if user.Status != "active" {
		return nil, errors.New("account is inactive or suspended")
	}

	// Verify password
	err = utils.VerifyPassword(user.PasswordHash, password)
	if err != nil {
		// Fallback for legacy plain text passwords during migration/testing
		if user.PasswordHash == password {
			fmt.Printf("WARNING: User %s still using plain text password. Please update for security.\n", email)
		} else {
			return nil, errors.New("invalid email or password")
		}
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role, s.jwtExpiry)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &LoginResponse{
		Token: token,
		User: UserSummary{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			NimNip:   user.NimNip,
			Role:     user.Role,
			Status:   user.Status,
		},
	}, nil
}

// Register creates a new user (admin only)
func (s *AuthService) Register(req *RegisterRequest) (*User, error) {
	return s.createUser(req.Email, req.Password, req.FullName, req.NimNip, req.Role)
}

// PublicRegister creates a new mahasiswa user (public)
func (s *AuthService) PublicRegister(req *PublicRegisterRequest) (*User, error) {
	return s.createUser(req.Email, req.Password, req.FullName, req.NimNip, "mahasiswa")
}

// Internal helper to create user
func (s *AuthService) createUser(email, password, fullName, nimNip, role string) (*User, error) {
	// Check if email already exists
	exists, err := s.repo.CheckEmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// Check if NIM/NIP already exists
	exists, err = s.repo.CheckNimNipExists(nimNip)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("NIM/NIP already registered")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("failed to secure password")
	}

	// Create user
	user := &User{
		Email:        email,
		PasswordHash: hashedPassword,
		FullName:     fullName,
		NimNip:       nimNip,
		Role:         role,
		Status:       "active",
	}

	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

// GetUserByID gets user by ID
func (s *AuthService) GetUserByID(userID uint) (*User, error) {
	return s.repo.FindByID(userID)
}

// UpdateProfile updates basic user information
func (s *AuthService) UpdateProfile(userID uint, req *UpdateProfileRequest) (*User, error) {
	updates := map[string]interface{}{
		"full_name": req.FullName,
	}
	if err := s.repo.Update(userID, updates); err != nil {
		return nil, err
	}
	return s.repo.FindByID(userID)
}

// UpdatePassword updates user password after verifying old password
func (s *AuthService) UpdatePassword(userID uint, req *UpdatePasswordRequest) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}

	// Verify old password
	err = utils.VerifyPassword(user.PasswordHash, req.OldPassword)
	if err != nil {
		// Fallback for legacy plain text
		if user.PasswordHash != req.OldPassword {
			return errors.New("current password incorrect")
		}
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to secure new password")
	}

	return s.repo.UpdatePassword(userID, hashedPassword)
}

// UpdatePIN updates user PIN
func (s *AuthService) UpdatePIN(userID uint, req *UpdatePinRequest) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}

	// If user already has a PIN, verify old one
	if user.PinHash != "" && req.OldPin != "" {
		err = utils.VerifyPassword(user.PinHash, req.OldPin)
		if err != nil && user.PinHash != req.OldPin {
			return errors.New("current PIN incorrect")
		}
	} else if user.PinHash != "" && req.OldPin == "" {
		return errors.New("current PIN is required to change to a new one")
	}

	// Hash new PIN
	hashedPin, err := utils.HashPassword(req.NewPin)
	if err != nil {
		return errors.New("failed to secure new PIN")
	}

	return s.repo.Update(userID, map[string]interface{}{"pin_hash": hashedPin})
}

// VerifyPIN checks if the provided PIN is correct
func (s *AuthService) VerifyPIN(userID uint, pin string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}

	if user.PinHash == "" {
		return errors.New("transaction PIN has not been set. Please set your PIN in Security settings first.")
	}

	err = utils.VerifyPassword(user.PinHash, pin)
	if err != nil && user.PinHash != pin {
		return errors.New("invalid transaction PIN code")
	}

	return nil
}

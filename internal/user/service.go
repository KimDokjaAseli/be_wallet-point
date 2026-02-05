package user

import (
	"errors"
	"math"
	"wallet-point/utils"
)

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetAllUsers gets all users with pagination and filters
func (s *UserService) GetAllUsers(params UserListParams) (*UserListResponse, error) {
	// Default pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 20
	}

	users, total, err := s.repo.GetAllWithWallets(params)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &UserListResponse{
		Users:      users,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetUserByID gets user by ID with wallet
func (s *UserService) GetUserByID(userID uint) (*UserWithWallet, error) {
	return s.repo.FindByIDWithWallet(userID)
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(userID uint, req *UpdateUserRequest) (*User, error) {
	// Check if user exists
	_, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Prepare updates
	updates := make(map[string]interface{})

	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Email != "" {
		// Check if email already exists
		exists, err := s.repo.CheckEmailExists(req.Email, userID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("email already exists")
		}
		updates["email"] = req.Email
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}

	// Update user
	if len(updates) > 0 {
		if err := s.repo.Update(userID, updates); err != nil {
			return nil, err
		}
	}

	// Return updated user
	return s.repo.FindByID(userID)
}

// DeactivateUser deactivates user account
func (s *UserService) DeactivateUser(userID uint) error {
	// Check if user exists
	_, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}

	return s.repo.Delete(userID)
}

// ChangeUserPassword changes user password (admin function)
func (s *UserService) ChangeUserPassword(userID uint, newPassword string) error {
	// Check if user exists
	_, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to secure new password")
	}

	return s.repo.UpdatePassword(userID, hashedPassword)
}

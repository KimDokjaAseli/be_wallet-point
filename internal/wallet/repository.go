package wallet

import (
	"errors"

	"gorm.io/gorm"
)

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// FindByID finds wallet by ID
func (r *WalletRepository) FindByID(walletID uint) (*Wallet, error) {
	var wallet Wallet
	err := r.db.First(&wallet, walletID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}
	return &wallet, nil
}

// FindByUserID finds wallet by user ID
func (r *WalletRepository) FindByUserID(userID uint) (*Wallet, error) {
	var wallet Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}
	return &wallet, nil
}

// GetAllWithUsers gets all wallets with user information
func (r *WalletRepository) GetAllWithUsers() ([]WalletWithUser, error) {
	var wallets []WalletWithUser
	err := r.db.Table("wallets").
		Select("wallets.id as wallet_id, users.id as user_id, users.email, users.full_name, users.nim_nip, users.role, wallets.balance, wallets.last_sync_at").
		Joins("INNER JOIN users ON wallets.user_id = users.id").
		Order("wallets.balance DESC").
		Scan(&wallets).Error
	return wallets, err
}

// CreateTransaction creates a new wallet transaction
func (r *WalletRepository) CreateTransaction(tx *gorm.DB, transaction *WalletTransaction) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(transaction).Error
}

// UpdateBalance updates wallet balance
func (r *WalletRepository) UpdateBalance(tx *gorm.DB, walletID uint, delta int) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&Wallet{}).
		Where("id = ?", walletID).
		Update("balance", gorm.Expr("balance + ?", delta)).
		Error
}

// SetBalance sets wallet balance to specific value (for reset)
func (r *WalletRepository) SetBalance(tx *gorm.DB, walletID uint, newBalance int) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&Wallet{}).
		Where("id = ?", walletID).
		Update("balance", newBalance).
		Error
}

// GetTransactions gets transactions with filters and pagination
func (r *WalletRepository) GetTransactions(params TransactionListParams) ([]TransactionWithDetails, int64, error) {
	var transactions []TransactionWithDetails
	var total int64

	query := r.db.Table("wallet_transactions").
		Select("wallet_transactions.*, users.email as user_email, users.full_name as user_name, users.nim_nip").
		Joins("INNER JOIN wallets ON wallet_transactions.wallet_id = wallets.id").
		Joins("INNER JOIN users ON wallets.user_id = users.id")

	// Apply filters
	if params.Type != "" {
		query = query.Where("wallet_transactions.type = ?", params.Type)
	}
	if params.Status != "" {
		query = query.Where("wallet_transactions.status = ?", params.Status)
	}
	if params.Direction != "" {
		query = query.Where("wallet_transactions.direction = ?", params.Direction)
	}
	if params.FromDate != "" {
		query = query.Where("wallet_transactions.created_at >= ?", params.FromDate)
	}
	if params.ToDate != "" {
		query = query.Where("wallet_transactions.created_at <= ?", params.ToDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	query = query.Limit(params.Limit).Offset(offset).Order("wallet_transactions.created_at DESC")

	if err := query.Scan(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetWalletTransactions gets transactions for specific wallet
func (r *WalletRepository) GetWalletTransactions(walletID uint, limit int) ([]WalletTransaction, error) {
	var transactions []WalletTransaction
	err := r.db.Where("wallet_id = ?", walletID).
		Order("created_at DESC").
		Limit(limit).
		Find(&transactions).Error
	return transactions, err
}

// GetLeaderboard retrieves top wallets by balance
func (r *WalletRepository) GetLeaderboard(limit int) ([]WalletWithUser, error) {
	var results []WalletWithUser
	// Only fetch mahasiswa role for leaderboard
	err := r.db.Table("wallets").
		Select("users.full_name, users.nim_nip, wallets.balance").
		Joins("INNER JOIN users ON wallets.user_id = users.id").
		Where("users.role = 'mahasiswa'").
		Order("wallets.balance DESC").
		Limit(limit).
		Scan(&results).Error
	return results, err
}

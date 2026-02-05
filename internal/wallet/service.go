package wallet

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"wallet-point/internal/auth"

	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

type WalletService struct {
	repo        *WalletRepository
	db          *gorm.DB
	authService *auth.AuthService
}

func (s *WalletService) SetAuthService(authService *auth.AuthService) {
	s.authService = authService
}

func NewWalletService(repo *WalletRepository, db *gorm.DB) *WalletService {
	return &WalletService{
		repo: repo,
		db:   db,
	}
}

// GetWalletByUserID retrieves a user's wallet
func (s *WalletService) GetWalletByUserID(userID uint) (*Wallet, error) {
	return s.repo.FindByUserID(userID)
}

// GetWalletByID finds wallet by ID
func (s *WalletService) GetWalletByID(walletID uint) (*Wallet, error) {
	return s.repo.FindByID(walletID)
}

// AdjustPoints adds or subtracts points from a wallet
func (s *WalletService) AdjustPoints(req *AdjustmentRequest, adminID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		delta := req.Amount
		if req.Direction == "debit" {
			delta = -req.Amount
		}

		// Check balance for debit
		if req.Direction == "debit" {
			wallet, err := s.repo.FindByID(req.WalletID)
			if err != nil {
				return err
			}
			if wallet.Balance < req.Amount {
				return errors.New("insufficient balance")
			}
		}

		err := s.repo.UpdateBalance(tx, req.WalletID, delta)
		if err != nil {
			return err
		}

		txn := &WalletTransaction{
			WalletID:    req.WalletID,
			Type:        "adjustment",
			Amount:      req.Amount,
			Direction:   req.Direction,
			Status:      "success",
			Description: req.Description,
			CreatedBy:   "admin",
		}

		return s.repo.CreateTransaction(tx, txn)
	})
}

// ResetWallet resets a wallet to a specific balance
func (s *WalletService) ResetWallet(req *ResetWalletRequest, adminID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		wallet, err := s.repo.FindByID(req.WalletID)
		if err != nil {
			return err
		}

		err = s.repo.SetBalance(tx, req.WalletID, req.NewBalance)
		if err != nil {
			return err
		}

		txn := &WalletTransaction{
			WalletID:    req.WalletID,
			Type:        "adjustment",
			Amount:      int(math.Abs(float64(req.NewBalance - wallet.Balance))),
			Direction:   "debit", // default direction, could be more specific
			Status:      "success",
			Description: "Reset Wallet: " + req.Reason,
			CreatedBy:   "admin",
		}
		if req.NewBalance > wallet.Balance {
			txn.Direction = "credit"
		}

		return s.repo.CreateTransaction(tx, txn)
	})
}

func (s *WalletService) GetTransactions(params TransactionListParams) ([]TransactionWithDetails, int64, error) {
	return s.repo.GetTransactions(params)
}

// GetAllTransactions is an alias for GetTransactions with default params or specifically for admin
func (s *WalletService) GetAllTransactions(params TransactionListParams) ([]TransactionWithDetails, int64, error) {
	return s.repo.GetTransactions(params)
}

func (s *WalletService) GetWalletTransactions(walletID uint, limit int) ([]WalletTransaction, error) {
	return s.repo.GetWalletTransactions(walletID, limit)
}

func (s *WalletService) GetLeaderboard(limit int) ([]WalletWithUser, error) {
	return s.repo.GetLeaderboard(limit)
}

func (s *WalletService) GetAllWallets() ([]WalletWithUser, error) {
	return s.repo.GetAllWithUsers()
}

// GeneratePaymentToken creates a secure token for QR payment
func (s *WalletService) GeneratePaymentToken(req PaymentTokenRequest, userID uint, recipientID uint) (*PaymentToken, error) {
	// 1. Get creator's wallet
	wallet, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("wallet not found")
	}

	// Note: We don't check balance here because in most flows (Bill QR),
	// the creator is the recipient. Balance is checked during consumption
	// from the scanner's wallet.

	// 2. Generate secure random token
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	tokenCode := hex.EncodeToString(b)

	// 3. Generate QR Code Image & Payload
	qrPayload := fmt.Sprintf("WPT:%s:%d:%s", tokenCode, req.Amount, req.Merchant)
	qrCode, err := qrcode.Encode(qrPayload, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	// 4. Create record
	token := &PaymentToken{
		Token:        tokenCode,
		QRCodeBase64: base64.StdEncoding.EncodeToString(qrCode),
		Amount:       req.Amount,
		Merchant:     req.Merchant,
		Expiry:       time.Now().Add(10 * time.Minute),
		WalletID:     wallet.ID,
		RecipientID:  recipientID,
		Status:       "active",
		Type:         req.Type,
		ProductID:    req.ProductID,
	}

	if err := s.db.Create(token).Error; err != nil {
		return nil, err
	}

	return token, nil
}

// ValidateAndConsumeToken verifies if a token is valid (legacy support for some modules)
func (s *WalletService) ValidateAndConsumeToken(tokenCode string, userID uint, amount int) error {
	var token PaymentToken
	err := s.db.Where("token = ? AND status = ?", tokenCode, "active").First(&token).Error
	if err != nil {
		return errors.New("invalid or expired QR token")
	}

	if time.Now().After(token.Expiry) {
		s.db.Model(&token).Update("status", "expired")
		return errors.New("QR token has expired")
	}

	wallet, err := s.repo.FindByUserID(userID)
	if err != nil || wallet.ID != token.WalletID {
		return errors.New("token does not belong to this user")
	}

	if token.Amount != amount {
		return fmt.Errorf("token amount mismatch. Expected: %d, Found: %d", token.Amount, amount)
	}

	return s.db.Model(&token).Update("status", "consumed").Error
}

// GetTokenDetails returns full token info regardless of status (active/consumed/expired)
func (s *WalletService) GetTokenDetails(tokenCode string) (*PaymentToken, error) {
	var token PaymentToken
	err := s.db.Where("token = ?", tokenCode).First(&token).Error
	if err != nil {
		return nil, errors.New("token tidak ditemukan")
	}

	// Dynamic check for expiry if still marked as active
	if token.Status == "active" && time.Now().After(token.Expiry) {
		token.Status = "expired"
		s.db.Model(&token).Update("status", "expired")
	}

	return &token, nil
}

// StudentPayToken executes a payment from a student scanning a bill
func (s *WalletService) StudentPayToken(tokenCode string, scannerUserID uint, pin string) error {
	// 1. Verify PIN
	if err := s.authService.VerifyPIN(scannerUserID, pin); err != nil {
		return err
	}

	var token PaymentToken
	if err := s.db.Where("token = ? AND status = ?", tokenCode, "active").First(&token).Error; err != nil {
		return errors.New("token tidak valid")
	}

	if time.Now().After(token.Expiry) {
		return errors.New("token kadaluarsa")
	}

	scannerWallet, err := s.repo.FindByUserID(scannerUserID)
	if err != nil {
		return errors.New("wallet pembayar tidak ditemukan")
	}

	if scannerWallet.Balance < token.Amount {
		return errors.New("saldo tidak mencukupi")
	}

	// Recipient logic
	var recipientID uint = token.RecipientID
	if recipientID == 0 {
		// Fallback to finding the first Admin
		var adminUser struct {
			ID       uint
			FullName string
		}
		err := s.db.Table("users").Where("role = ?", "admin").Select("id, full_name").Order("id asc").First(&adminUser).Error
		if err != nil {
			log.Printf("[StudentPayToken] No admin user found in database: %v", err)
			return errors.New("sistem gagal menemukan admin sebagai penerima")
		}
		recipientID = adminUser.ID
	}

	recipientWallet, err := s.repo.FindByUserID(recipientID)
	if err != nil {
		// If wallet doesn't exist, create it (every user should have one)
		log.Printf("[StudentPayToken] Recipient (User %d) has no wallet, creating one...", recipientID)
		newWallet := &Wallet{
			UserID:  recipientID,
			Balance: 0,
		}
		if err := s.db.Create(newWallet).Error; err != nil {
			return errors.New("gagal menyiapkan wallet penerima")
		}
		recipientWallet = newWallet
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Deduct from scanner
		if err := s.repo.UpdateBalance(tx, scannerWallet.ID, -token.Amount); err != nil {
			return err
		}

		// 2. Credit to recipient
		if err := s.repo.UpdateBalance(tx, recipientWallet.ID, token.Amount); err != nil {
			return err
		}

		// 3. Mark token as consumed
		if err := tx.Model(&token).Update("status", "consumed").Error; err != nil {
			return err
		}

		// 4. Handle Descriptions and Types based on Token Type
		txnType := "marketplace"
		payerDesc := "Pembayaran QR"
		recipientDesc := fmt.Sprintf("Terima Pembayaran QR dari User #%d", scannerUserID)
		var refID *uint

		if token.Type == "transfer" {
			txnType = "transfer" // Note: repository/model might use transfer_in/transfer_out
			payerDesc = fmt.Sprintf("Transfer QR ke User #%d", recipientID)
			recipientDesc = fmt.Sprintf("Terima Transfer QR dari User #%d", scannerUserID)
		} else if token.Type == "purchase" {
			txnType = "marketplace"
			if token.ProductID != 0 {
				refID = &token.ProductID
				var prodName string
				tx.Table("products").Where("id = ?", token.ProductID).Select("name").Scan(&prodName)
				if prodName != "" {
					payerDesc = "Beli: " + prodName
					recipientDesc = fmt.Sprintf("Penjualan: %s ke User #%d", prodName, scannerUserID)
				}

				// Reduce Stock
				if err := tx.Table("products").Where("id = ?", token.ProductID).Update("stock", gorm.Expr("stock - ?", 1)).Error; err != nil {
					return fmt.Errorf("gagal memperbarui stok: %v", err)
				}
			}
		}

		// 5. Create Transactions
		// Determine specific types if enum requires transfer_in/transfer_out
		payerType := txnType
		recipientType := txnType
		if txnType == "transfer" {
			payerType = "transfer_out"
			recipientType = "transfer_in"
		}

		tx.Create(&WalletTransaction{
			WalletID:    scannerWallet.ID,
			Type:        payerType,
			Amount:      token.Amount,
			Direction:   "debit",
			Status:      "success",
			Description: payerDesc,
			ReferenceID: refID,
		})

		tx.Create(&WalletTransaction{
			WalletID:    recipientWallet.ID,
			Type:        recipientType,
			Amount:      token.Amount,
			Direction:   "credit",
			Status:      "success",
			Description: recipientDesc,
			ReferenceID: refID,
		})

		return nil
	})
}

// DebitWithTransaction handles point deduction within an existing transaction
func (s *WalletService) DebitWithTransaction(tx *gorm.DB, walletID uint, amount int, txnType string, description string) error {
	// 1. Check balance
	wallet, err := s.repo.FindByID(walletID)
	if err != nil {
		return err
	}
	if wallet.Balance < amount {
		return errors.New("insufficient balance")
	}

	// 2. Update balance
	if err := s.repo.UpdateBalance(tx, walletID, -amount); err != nil {
		return err
	}

	// 3. Create transaction record
	txn := &WalletTransaction{
		WalletID:    walletID,
		Type:        txnType,
		Amount:      amount,
		Direction:   "debit",
		Status:      "success",
		Description: description,
	}

	return s.repo.CreateTransaction(tx, txn)
}

// CreditWithTransaction handles point addition within an existing transaction
func (s *WalletService) CreditWithTransaction(tx *gorm.DB, walletID uint, amount int, txnType string, description string) error {
	// 1. Update balance
	if err := s.repo.UpdateBalance(tx, walletID, amount); err != nil {
		return err
	}

	// 2. Create transaction record
	txn := &WalletTransaction{
		WalletID:    walletID,
		Type:        txnType,
		Amount:      amount,
		Direction:   "credit",
		Status:      "success",
		Description: description,
	}

	return s.repo.CreateTransaction(tx, txn)
}

// ProcessMissionRewardWithTx handles mission rewards within a transaction
func (s *WalletService) ProcessMissionRewardWithTx(tx *gorm.DB, userID uint, amount int, missionTitle string, missionID uint, reviewerID uint) error {
	wallet, err := s.repo.FindByUserID(userID)
	if err != nil {
		return err
	}

	// Create transaction record
	txn := &WalletTransaction{
		WalletID:    wallet.ID,
		Type:        "mission",
		Amount:      amount,
		Direction:   "credit",
		Status:      "success",
		Description: "Reward for mission: " + missionTitle,
		ReferenceID: &missionID,
		CreatedBy:   "dosen",
	}

	// Update balance
	if err := s.repo.UpdateBalance(tx, wallet.ID, amount); err != nil {
		return err
	}

	return s.repo.CreateTransaction(tx, txn)
}

func (s *WalletService) GetAdminStats() (*AdminStats, error) {
	var stats AdminStats

	// 1. User Stats
	s.db.Table("users").Count(&stats.TotalUsers)
	s.db.Table("users").Where("status = ?", "active").Count(&stats.ActiveUsers)

	// 2. Circulation Points
	s.db.Table("wallets").Select("COALESCE(SUM(balance), 0)").Scan(&stats.CirculationPoints)

	// 3. Today Stats
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		location = time.Local
	}
	now := time.Now().In(location)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)

	s.db.Model(&WalletTransaction{}).Where("created_at >= ?", startOfDay).Count(&stats.TodayTransactions)

	s.db.Model(&WalletTransaction{}).
		Where("created_at >= ? AND direction = ?", startOfDay, "credit").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TodayCredits)

	s.db.Model(&WalletTransaction{}).
		Where("created_at >= ? AND direction = ?", startOfDay, "debit").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TodayDebits)

	// 4. Additional System Monitoring
	s.db.Table("products").Where("status = ?", "active").Count(&stats.TotalProducts)
	s.db.Table("missions").Count(&stats.TotalMissions)
	s.db.Table("mission_submissions").Where("status = ?", "pending").Count(&stats.PendingSubmission)

	return &stats, nil
}

package transfer

import (
	"errors"
	"fmt"
	"wallet-point/internal/auth"
	"wallet-point/internal/wallet"

	"gorm.io/gorm"
)

type Service struct {
	walletRepo    *wallet.WalletRepository
	walletService *wallet.WalletService
	authService   *auth.AuthService
	db            *gorm.DB
}

func NewService(walletRepo *wallet.WalletRepository, walletService *wallet.WalletService, authService *auth.AuthService, db *gorm.DB) *Service {
	return &Service{
		walletRepo:    walletRepo,
		walletService: walletService,
		authService:   authService,
		db:            db,
	}
}

func (s *Service) CreateTransfer(senderUserID, receiverUserID uint, amount int, description string, pin string) (*TransferInfo, error) {
	// 1. Verify PIN
	if err := s.authService.VerifyPIN(senderUserID, pin); err != nil {
		return nil, err
	}

	if senderUserID == receiverUserID {
		return nil, errors.New("cannot transfer points to yourself")
	}

	senderWallet, err := s.walletService.GetWalletByUserID(senderUserID)
	if err != nil {
		return nil, errors.New("sender wallet not found")
	}

	receiverWallet, err := s.walletService.GetWalletByUserID(receiverUserID)
	if err != nil {
		return nil, errors.New("receiver wallet not found: check if user exists and has a wallet")
	}

	if senderWallet.Balance < amount {
		return nil, errors.New("insufficient balance")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Deduct from sender
		if err := s.walletService.DebitWithTransaction(tx, senderWallet.ID, amount, "transfer_out", fmt.Sprintf("Transfer to user %d: %s", receiverUserID, description)); err != nil {
			return err
		}

		// 2. Credit to receiver
		if err := s.walletService.CreditWithTransaction(tx, receiverWallet.ID, amount, "transfer_in", fmt.Sprintf("Transfer from user %d: %s", senderUserID, description)); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Return a virtual TransferInfo for the response
	return &TransferInfo{
		SenderWalletID:   senderWallet.ID,
		ReceiverWalletID: receiverWallet.ID,
		Amount:           amount,
		Description:      description,
	}, nil
}

func (s *Service) GetUserTransfers(userID uint, limit, page int) ([]wallet.TransactionWithDetails, int64, error) {
	walletData, err := s.walletService.GetWalletByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// We query wallet_transactions for types transfer_in/transfer_out
	// Since GetTransactions only takes one type, we might need to modify it or do a manual query here.
	// For now, let's just query everything and filter in our head, or do a custom query.

	var txs []wallet.TransactionWithDetails
	var total int64

	baseQuery := s.db.Table("wallet_transactions").
		Select("wallet_transactions.*, users.email as user_email, users.full_name as user_name, users.nim_nip").
		Joins("INNER JOIN wallets ON wallet_transactions.wallet_id = wallets.id").
		Joins("INNER JOIN users ON wallets.user_id = users.id").
		Where("wallet_transactions.wallet_id = ? AND wallet_transactions.type IN ('transfer_in', 'transfer_out', 'marketplace')", walletData.ID)

	baseQuery.Count(&total)

	err = baseQuery.Order("created_at DESC").Limit(limit).Offset((page - 1) * limit).Scan(&txs).Error
	return txs, total, err
}

func (s *Service) FindRecipient(userID uint) (*RecipientSummary, error) {
	// Check if user has a wallet
	w, err := s.walletService.GetWalletByUserID(userID)
	if err != nil {
		return nil, errors.New("user not found or has no wallet")
	}

	var recipient RecipientSummary
	err = s.db.Table("users").
		Select("id, full_name, role, nim_nip as nim").
		Where("id = ?", w.UserID).
		Scan(&recipient).Error

	if err != nil {
		return nil, err
	}
	if recipient.ID == 0 {
		return nil, errors.New("user not found")
	}

	return &recipient, nil
}

func (s *Service) GetAllTransfers(limit, page int) ([]TransferInfo, int64, error) {
	params := wallet.TransactionListParams{
		Page:  page,
		Limit: limit,
	}

	// Fetch all transactions of type transfer_out (the sender's perspective) to list unique transfers
	params.Type = "transfer_out"

	txns, total, err := s.walletService.GetTransactions(params)
	if err != nil {
		return nil, 0, err
	}

	var transfers []TransferInfo
	for _, t := range txns {
		transfers = append(transfers, TransferInfo{
			ID:               t.ID,
			SenderWalletID:   t.WalletID,
			ReceiverWalletID: 0,
			Amount:           t.Amount,
			Description:      t.Description,
			CreatedAt:        t.CreatedAt,
			SenderName:       t.UserName,
			SenderNIM:        t.NimNip,
		})
	}

	return transfers, total, nil
}

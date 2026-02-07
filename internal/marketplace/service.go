package marketplace

import (
	"errors"
	"fmt"
	"math"
	"wallet-point/internal/auth"
	"wallet-point/internal/wallet"

	"gorm.io/gorm"
)

type MarketplaceService struct {
	repo          *MarketplaceRepository
	walletService *wallet.WalletService
	authService   *auth.AuthService
	db            *gorm.DB
}

func NewMarketplaceService(repo *MarketplaceRepository, walletService *wallet.WalletService, authService *auth.AuthService, db *gorm.DB) *MarketplaceService {
	return &MarketplaceService{
		repo:          repo,
		walletService: walletService,
		authService:   authService,
		db:            db,
	}
}

// GetAllProducts gets all products with pagination and filters
func (s *MarketplaceService) GetAllProducts(params ProductListParams) (*ProductListResponse, error) {
	// Default pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 20
	}

	products, total, err := s.repo.GetAll(params)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetProductByID gets product by ID
func (s *MarketplaceService) GetProductByID(productID uint) (*Product, error) {
	return s.repo.FindByID(productID)
}

// CreateProduct creates a new product
func (s *MarketplaceService) CreateProduct(req *CreateProductRequest, adminID uint) (*Product, error) {
	product := &Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		Status:      "active",
		CreatedBy:   adminID,
	}

	if err := s.repo.Create(product); err != nil {
		return nil, errors.New("failed to create product")
	}

	return product, nil
}

// UpdateProduct updates product
func (s *MarketplaceService) UpdateProduct(productID uint, req *UpdateProductRequest) (*Product, error) {
	_, err := s.repo.FindByID(productID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Price > 0 {
		updates["price"] = req.Price
	}
	if req.Stock >= 0 {
		updates["stock"] = req.Stock
	}
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		if err := s.repo.Update(productID, updates); err != nil {
			return nil, errors.New("failed to update product")
		}
	}

	return s.repo.FindByID(productID)
}

// DeleteProduct deletes product
func (s *MarketplaceService) DeleteProduct(productID uint) error {
	_, err := s.repo.FindByID(productID)
	if err != nil {
		return err
	}
	return s.repo.Delete(productID)
}

// PurchaseProduct handles product purchase without a dedicated marketplace_transactions table
func (s *MarketplaceService) PurchaseProduct(userID uint, req *PurchaseRequest) error {
	// 1. Verify PIN if using direct wallet
	if req.PaymentMethod == "wallet" || req.PaymentMethod == "" {
		if err := s.authService.VerifyPIN(userID, req.PIN); err != nil {
			return err
		}
	}

	product, err := s.repo.FindByID(req.ProductID)
	if err != nil {
		return err
	}

	if product.Status == "inactive" {
		return errors.New("product is not active")
	}
	if product.Stock < 1 {
		return errors.New("product out of stock")
	}

	studentWallet, err := s.walletService.GetWalletByUserID(userID)
	if err != nil {
		return err
	}

	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}
	totalPrice := product.Price * quantity

	if studentWallet.Balance < totalPrice {
		return fmt.Errorf("insufficient balance. Required: %d", totalPrice)
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Debit Student Wallet
		desc := fmt.Sprintf("Purchase: %dx %s", quantity, product.Name)
		if err := s.walletService.DebitWithTransaction(tx, studentWallet.ID, totalPrice, "marketplace", desc); err != nil {
			return err
		}

		// 2. Reduce Stock
		if err := s.repo.UpdateStock(tx, product.ID, -quantity); err != nil {
			return err
		}

		// 3. Record in Marketplace Transactions
		txn := &MarketplaceTransaction{
			WalletID:      studentWallet.ID,
			ProductID:     product.ID,
			Amount:        product.Price,
			TotalAmount:   totalPrice,
			Quantity:      quantity,
			StudentName:   req.StudentName,
			StudentNPM:    req.StudentNPM,
			StudentMajor:  req.StudentMajor,
			StudentBatch:  req.StudentBatch,
			PaymentMethod: "wallet",
			Status:        "success",
		}
		if err := s.repo.CreateMarketplaceTransaction(tx, txn); err != nil {
			return err
		}

		return nil
	})

	return err
}

// GetTransactions retrieves all marketplace transactions from consolidated wallet_transactions (Admin)
func (s *MarketplaceService) GetTransactions(limit, page int) ([]MarketplaceTransactionWithDetails, int64, error) {
	if page < 1 {
		page = 1
	}
	return s.repo.GetTransactions(limit, page)
}

// Cart Methods

func (s *MarketplaceService) AddToCart(userID uint, req AddToCartRequest) error {
	// Check if product exists and has stock
	product, err := s.repo.FindByID(req.ProductID)
	if err != nil {
		return err
	}
	if product.Stock < req.Quantity {
		return errors.New("stok tidak mencukupi")
	}

	item := &CartItem{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}
	return s.repo.AddToCart(item)
}

func (s *MarketplaceService) GetCart(userID uint) (*CartResponse, error) {
	items, err := s.repo.GetCart(userID)
	if err != nil {
		return nil, err
	}

	totalPrice := 0
	for _, item := range items {
		totalPrice += item.Product.Price * item.Quantity
	}

	return &CartResponse{
		Items:      items,
		TotalPrice: totalPrice,
	}, nil
}

func (s *MarketplaceService) UpdateCartItem(userID, itemID uint, quantity int) error {
	return s.repo.UpdateCartItem(userID, itemID, quantity)
}

func (s *MarketplaceService) RemoveFromCart(userID, itemID uint) error {
	return s.repo.RemoveFromCart(userID, itemID)
}

func (s *MarketplaceService) Checkout(userID uint, req CartCheckoutRequest) error {
	// 1. Verify PIN
	if err := s.authService.VerifyPIN(userID, req.PIN); err != nil {
		return err
	}

	// 2. Get Cart Items
	items, err := s.repo.GetCart(userID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return errors.New("keranjang belanja kosong")
	}

	// 3. Calculate total and check stock
	totalPrice := 0
	for _, item := range items {
		if item.Product.Stock < item.Quantity {
			return fmt.Errorf("stok produk '%s' tidak mencukupi", item.Product.Name)
		}
		totalPrice += item.Product.Price * item.Quantity
	}

	// 4. Check balance
	wallet, err := s.walletService.GetWalletByUserID(userID)
	if err != nil {
		return err
	}
	if wallet.Balance < totalPrice {
		return fmt.Errorf("saldo tidak cukup. Total: %d, Saldo: %d", totalPrice, wallet.Balance)
	}

	// 5. Execute Transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			// Debit wallet for each item
			desc := fmt.Sprintf("Purchase: %dx %s", item.Quantity, item.Product.Name)
			if err := s.walletService.DebitWithTransaction(tx, wallet.ID, item.Product.Price*item.Quantity, "marketplace", desc); err != nil {
				return err
			}

			// Reduce stock
			if err := s.repo.UpdateStock(tx, item.ProductID, -item.Quantity); err != nil {
				return err
			}

			// Record in Marketplace Transactions
			txn := &MarketplaceTransaction{
				WalletID:      wallet.ID,
				ProductID:     item.ProductID,
				Amount:        item.Product.Price,
				TotalAmount:   item.Product.Price * item.Quantity,
				Quantity:      item.Quantity,
				PaymentMethod: "wallet",
				Status:        "success",
			}
			if err := s.repo.CreateMarketplaceTransaction(tx, txn); err != nil {
				return err
			}
		}

		// Clear cart
		return s.repo.ClearCart(tx, userID)
	})
}

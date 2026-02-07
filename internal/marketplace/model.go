package marketplace

import (
	"time"
)

type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	Price       int       `json:"price" gorm:"not null"`
	Stock       int       `json:"stock" gorm:"default:0;not null"`
	ImageURL    string    `json:"image_url" gorm:"size:500"`
	Status      string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedBy   uint      `json:"created_by" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Product) TableName() string {
	return "products"
}

type MarketplaceTransaction struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	WalletID      uint      `json:"wallet_id" gorm:"not null;index"`
	ProductID     uint      `json:"product_id" gorm:"not null;index"`
	Amount        int       `json:"amount" gorm:"not null"`                           // Individual item price
	TotalAmount   int       `json:"total_amount" gorm:"column:total_amount;not null"` // This fixes the DB constraint error
	Quantity      int       `json:"quantity" gorm:"default:1;not null"`
	StudentName   string    `json:"student_name" gorm:"size:255"`
	StudentNPM    string    `json:"student_npm" gorm:"size:100"`
	StudentMajor  string    `json:"student_major" gorm:"size:255"`
	StudentBatch  string    `json:"student_batch" gorm:"size:50"`
	PaymentMethod string    `json:"payment_method" gorm:"size:50;default:'wallet'"`
	Status        string    `json:"status" gorm:"type:enum('success','failed');default:'success'"`
	CreatedAt     time.Time `json:"created_at"`
}

type PurchaseRequest struct {
	ProductID     uint   `json:"product_id" binding:"required"`
	Quantity      int    `json:"quantity" binding:"omitempty,gt=0"`
	PaymentMethod string `json:"payment_method" binding:"omitempty,oneof=wallet qr"`
	PaymentToken  string `json:"payment_token"`
	PIN           string `json:"pin"`
	StudentName   string `json:"student_name"`
	StudentNPM    string `json:"student_npm"`
	StudentMajor  string `json:"student_major"`
	StudentBatch  string `json:"student_batch"`
}

func (MarketplaceTransaction) TableName() string {
	return "marketplace_transactions"
}

type MarketplaceTransactionWithDetails struct {
	ID            uint      `json:"id"`
	WalletID      uint      `json:"wallet_id"`
	ProductID     uint      `json:"product_id"`
	Amount        int       `json:"amount"`
	TotalAmount   int       `json:"total_amount"`
	Quantity      int       `json:"quantity"`
	StudentName   string    `json:"student_name"`
	StudentNPM    string    `json:"student_npm"`
	StudentMajor  string    `json:"student_major"`
	StudentBatch  string    `json:"student_batch"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	ProductName   string    `json:"product_name"`
	UserName      string    `json:"user_name"`
	UserEmail     string    `json:"user_email"`
}

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Price       int    `json:"price" binding:"required,gt=0"`
	Stock       int    `json:"stock" binding:"gte=0"`
	ImageURL    string `json:"image_url"`
}

type UpdateProductRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Price       int    `json:"price,omitempty" binding:"omitempty,gt=0"`
	Stock       int    `json:"stock,omitempty" binding:"omitempty,gte=0"`
	ImageURL    string `json:"image_url,omitempty"`
	Status      string `json:"status,omitempty" binding:"omitempty,oneof=active inactive"`
}

type ProductListParams struct {
	Status string
	Page   int
	Limit  int
}

type ProductListResponse struct {
	Products   []Product `json:"products"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

type CartItem struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	ProductID uint      `json:"product_id" gorm:"not null;index"`
	Quantity  int       `json:"quantity" gorm:"not null;default:1"`
	Product   Product   `json:"product" gorm:"foreignKey:ProductID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CartItem) TableName() string {
	return "cart_items"
}

type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type CartCheckoutRequest struct {
	PIN           string `json:"pin" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"oneof=wallet"`
}

type CartResponse struct {
	Items      []CartItem `json:"items"`
	TotalPrice int        `json:"total_price"`
}

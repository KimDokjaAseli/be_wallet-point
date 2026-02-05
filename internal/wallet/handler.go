package wallet

import (
	"math"
	"net/http"
	"strconv"
	"wallet-point/internal/audit"
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	service      *WalletService
	auditService *audit.AuditService
}

func NewWalletHandler(service *WalletService, auditService *audit.AuditService) *WalletHandler {
	return &WalletHandler{service: service, auditService: auditService}
}

// GetAllWallets handles getting all wallets
// @Summary Get all wallets
// @Description Get list of all wallets with user information (Admin only)
// @Tags Admin - Wallets
// @Security BearerAuth
// @Produce json
// @Success 200 {object} utils.Response{data=[]WalletWithUser}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /admin/wallets [get]
func (h *WalletHandler) GetAllWallets(c *gin.Context) {
	wallets, err := h.service.GetAllWallets()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve wallets", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Wallets retrieved successfully", wallets)
}

// GetWalletByID handles getting wallet by ID
// @Summary Get wallet by ID
// @Description Get wallet details by ID (Admin only)
// @Tags Admin - Wallets
// @Security BearerAuth
// @Produce json
// @Param id path int true "Wallet ID"
// @Success 200 {object} utils.Response{data=Wallet}
// @Failure 404 {object} utils.Response
// @Router /admin/wallets/{id} [get]
func (h *WalletHandler) GetWalletByID(c *gin.Context) {
	walletID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID", nil)
		return
	}

	wallet, err := h.service.GetWalletByID(uint(walletID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Wallet retrieved successfully", wallet)
}

// AdjustPoints handles manual point adjustment
// @Summary Adjust wallet points
// @Description Manually adjust wallet points (Admin only)
// @Tags Admin - Wallets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AdjustmentRequest true "Adjustment details"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /admin/wallet/adjustment [post]
func (h *WalletHandler) AdjustPoints(c *gin.Context) {
	adminID := c.GetUint("user_id")

	var req AdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := h.service.AdjustPoints(&req, adminID); err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "wallet not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Points adjusted successfully", nil)

	// Log activity
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    adminID,
		Action:    "ADJUST_POINTS",
		Entity:    "WALLET",
		EntityID:  req.WalletID,
		Details:   "Admin adjusted points: " + req.Direction + " " + strconv.Itoa(req.Amount) + " | Reason: " + req.Description,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// ResetWallet handles wallet reset
// @Summary Reset wallet balance
// @Description Reset wallet to specific balance (Admin only - Emergency use)
// @Tags Admin - Wallets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ResetWalletRequest true "Reset details"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils. Response
// @Router /admin/wallet/reset [post]
func (h *WalletHandler) ResetWallet(c *gin.Context) {
	adminID := c.GetUint("user_id")

	var req ResetWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := h.service.ResetWallet(&req, adminID); err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "wallet not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Wallet reset successfully", nil)

	// Log activity
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    adminID,
		Action:    "RESET_WALLET",
		Entity:    "WALLET",
		EntityID:  req.WalletID,
		Details:   "Admin reset wallet to " + strconv.Itoa(req.NewBalance) + " | Reason: " + req.Reason,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// GetAllTransactions handles getting all transactions
// @Summary Get all transactions
// @Description Get list of all transactions with filters (Admin only)
// @Tags Admin - Transactions
// @Security BearerAuth
// @Produce json
// @Param type query string false "Filter by type"
// @Param status query string false "Filter by status"
// @Param direction query string false "Filter by direction"
// @Param from_date query string false "Filter from date (YYYY-MM-DD)"
// @Param to_date query string false "Filter to date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} utils.Response{data=TransactionListResponse}
// @Failure 401 {object} utils.Response
// @Router /admin/transactions [get]
func (h *WalletHandler) GetAllTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	params := TransactionListParams{
		Type:      c.Query("type"),
		Status:    c.Query("status"),
		Direction: c.Query("direction"),
		FromDate:  c.Query("from_date"),
		ToDate:    c.Query("to_date"),
		Page:      page,
		Limit:     limit,
	}

	transactions, total, err := h.service.GetAllTransactions(params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve transactions", err.Error())
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := TransactionListResponse{
		Transactions: transactions,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}

	utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved successfully", response)
}

// GetWalletTransactions handles getting transactions for specific wallet
// @Summary Get wallet transactions
// @Description Get transaction history for specific wallet (Admin only)
// @Tags Admin - Wallets
// @Security BearerAuth
// @Produce json
// @Param id path int true "Wallet ID"
// @Param limit query int false "Number of transactions" default(50)
// @Success 200 {object} utils.Response{data=[]WalletTransaction}
// @Failure 404 {object} utils.Response
// @Router /admin/wallets/{id}/transactions [get]
func (h *WalletHandler) GetWalletTransactions(c *gin.Context) {
	walletID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid wallet ID", nil)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	transactions, err := h.service.GetWalletTransactions(uint(walletID), limit)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "wallet not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved successfully", transactions)
}

// GetLeaderboard handles getting leaderboard
// @Summary Get leaderboard
// @Description Get top users by wallet balance
// @Tags Wallet
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Success 200 {object} utils.Response{data=[]WalletWithUser}
// @Router /wallets/leaderboard [get]
func (h *WalletHandler) GetLeaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	leaderboard, err := h.service.GetLeaderboard(limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve leaderboard", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Leaderboard retrieved", leaderboard)
}

// GetMyWallet handles getting current user's wallet
// @Summary Get my wallet
// @Description Get current authenticated user's wallet details
// @Tags Wallet
// @Security BearerAuth
// @Produce json
// @Success 200 {object} utils.Response{data=Wallet}
// @Router /mahasiswa/wallet [get]
func (h *WalletHandler) GetMyWallet(c *gin.Context) {
	userID := c.GetUint("user_id")

	wallet, err := h.service.GetWalletByUserID(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Wallet not found", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Wallet retrieved successfully", wallet)
}

// GetMyTransactions handles getting current user's transaction history
// @Summary Get my transactions
// @Description Get current authenticated user's wallet transactions
// @Tags Wallet
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" default(50)
// @Success 200 {object} utils.Response{data=[]WalletTransaction}
// @Router /mahasiswa/transactions [get]
func (h *WalletHandler) GetMyTransactions(c *gin.Context) {
	userID := c.GetUint("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Find wallet first
	wallet, err := h.service.GetWalletByUserID(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Wallet not found", nil)
		return
	}

	transactions, err := h.service.GetWalletTransactions(wallet.ID, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve transactions", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved successfully", map[string]interface{}{
		"transactions": transactions,
	})
}

// GeneratePaymentToken handles generating a QR payment token
// @Summary Generate payment token
// @Description Generate a secure token for QR payment (Mahasiswa only)
// @Tags Wallet
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body PaymentTokenRequest true "Token details"
// @Success 200 {object} utils.Response{data=PaymentToken}
// @Router /mahasiswa/payment/token [post]
func (h *WalletHandler) GeneratePaymentToken(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req PaymentTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	token, err := h.service.GeneratePaymentToken(req, userID, req.RecipientID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment token generated successfully", token)
}

// CheckTokenStatus returns payment details for a token
func (h *WalletHandler) CheckTokenStatus(c *gin.Context) {
	tokenCode := c.Param("token")

	token, err := h.service.GetTokenDetails(tokenCode)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Token tidak valid atau sudah kadaluarsa", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Token info retrieved", token)
}

// ExecuteStudentPayment allows a student to pay for a scanned token
func (h *WalletHandler) ExecuteStudentPayment(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		Token string `json:"token" binding:"required"`
		PIN   string `json:"pin" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	err := h.service.StudentPayToken(req.Token, userID, req.PIN)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pembayaran berhasil!", nil)
}

// GetAdminStats handles retrieving administrative dashboard statistics
func (h *WalletHandler) GetAdminStats(c *gin.Context) {
	stats, err := h.service.GetAdminStats()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error fetching admin stats", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Admin stats retrieved", stats)
}

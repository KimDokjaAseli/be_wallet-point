package transfer

import (
	"fmt"
	"net/http"
	"strconv"
	"wallet-point/internal/audit"
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for transfers
type Handler struct {
	service      *Service
	auditService *audit.AuditService
}

// NewHandler creates a new transfer handler
func NewHandler(service *Service, auditService *audit.AuditService) *Handler {
	return &Handler{service: service, auditService: auditService}
}

// CreateTransfer handles POST /transfer
func (h *Handler) CreateTransfer(c *gin.Context) {
	senderUserID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	transfer, err := h.service.CreateTransfer(senderUserID.(uint), req.ReceiverUserID, req.Amount, req.Description, req.PIN)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transfer completed successfully", gin.H{
		"transfer": transfer,
	})

	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    senderUserID.(uint),
		Action:    "TRANSFER_POINTS",
		Entity:    "WALLET_TRANSACTION",
		Details:   fmt.Sprintf("Transferred %d points to user %d", req.Amount, req.ReceiverUserID),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// GetMyTransfers handles GET /transfer/history
func (h *Handler) GetMyTransfers(c *gin.Context) {
	userID, _ := c.Get("user_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	transfers, total, err := h.service.GetUserTransfers(userID.(uint), limit, page)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transfer history retrieved successfully", gin.H{
		"transfers": transfers,
		"total":     total,
		"limit":     limit,
		"page":      page,
	})
}

// GetRecipientInfo handles GET /transfer/recipient/:id
func (h *Handler) GetRecipientInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	recipient, err := h.service.FindRecipient(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Recipient found", recipient)
}

// GetAllTransfers handles GET /admin/transfers
func (h *Handler) GetAllTransfers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	transfers, total, err := h.service.GetAllTransfers(limit, page)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "All transfers retrieved", gin.H{
		"transfers": transfers,
		"total":     total,
		"limit":     limit,
		"page":      page,
	})
}

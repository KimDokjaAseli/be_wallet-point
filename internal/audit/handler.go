package audit

import (
	"net/http"
	"strconv"
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service *AuditService
}

func NewAuditHandler(service *AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

// GetAll handles fetching audit logs
// @Summary Get audit logs
// @Description Get system audit logs (Admin only)
// @Tags Admin - Monitoring
// @Security BearerAuth
// @Produce json
// @Param user_id query int false "Filter by User ID"
// @Param action query string false "Filter by Action"
// @Param date query string false "Filter by Date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} utils.Response{data=AuditListResponse}
// @Failure 401 {object} utils.Response
// @Router /admin/audit-logs [get]
func (h *AuditHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	userID, _ := strconv.Atoi(c.Query("user_id"))

	params := AuditListParams{
		UserID: userID,
		Action: c.Query("action"),
		Date:   c.Query("date"),
		Page:   page,
		Limit:  limit,
	}

	response, err := h.service.GetLogs(params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve audit logs", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Audit logs retrieved successfully", response)
}

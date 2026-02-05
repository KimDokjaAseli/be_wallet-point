package user

import (
	"net/http"
	"strconv"
	"wallet-point/internal/audit"
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service      *UserService
	auditService *audit.AuditService
}

func NewUserHandler(service *UserService, auditService *audit.AuditService) *UserHandler {
	return &UserHandler{service: service, auditService: auditService}
}

// GetAll handles getting all users
// @Summary Get all users
// @Description Get list of all users with pagination and filters (Admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param role query string false "Filter by role" Enums(admin, dosen, mahasiswa)
// @Param status query string false "Filter by status" Enums(active, inactive, suspended)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} utils.Response{data=UserListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /admin/users [get]
func (h *UserHandler) GetAll(c *gin.Context) {
	// Parse query parameters
	role := c.Query("role")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	params := UserListParams{
		Role:   role,
		Status: status,
		Page:   page,
		Limit:  limit,
	}

	response, err := h.service.GetAllUsers(params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve users", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Users retrieved successfully", response)
}

// GetByID handles getting user by ID
// @Summary Get user by ID
// @Description Get user details by ID (Admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} utils.Response{data=UserWithWallet}
// @Failure 404 {object} utils.Response
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	user, err := h.service.GetUserByID(uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", user)
}

// Update handles updating user
// @Summary Update user
// @Description Update user information (Admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body UpdateUserRequest true "Update data"
// @Success 200 {object} utils.Response{data=User}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /admin/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	user, err := h.service.UpdateUser(uint(userID), &req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User updated successfully", user)

	// Log activity
	adminID := c.GetUint("user_id")
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    adminID,
		Action:    "UPDATE_USER",
		Entity:    "USER",
		EntityID:  user.ID,
		Details:   "Admin updated user profile: " + user.Email,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// Deactivate handles deactivating user
// @Summary Deactivate user
// @Description Deactivate user account (Admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /admin/users/{id} [delete]
func (h *UserHandler) Deactivate(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	if err := h.service.DeactivateUser(uint(userID)); err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User deactivated successfully", nil)

	// Log activity
	adminID := c.GetUint("user_id")
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    adminID,
		Action:    "DEACTIVATE_USER",
		Entity:    "USER",
		EntityID:  uint(userID),
		Details:   "Admin deactivated user account",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// ChangePassword handles changing user password
// @Summary Change user password
// @Description Change user password (Admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body ChangePasswordRequest true "New password"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /admin/users/{id}/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := h.service.ChangeUserPassword(uint(userID), req.NewPassword); err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)

	// Log activity
	adminID := c.GetUint("user_id")
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    adminID,
		Action:    "RESET_PASSWORD",
		Entity:    "USER",
		EntityID:  uint(userID),
		Details:   "Admin reset user password",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// LookupUser handles looking up a user for public operations (like transfer)
// @Summary Lookup user
// @Description Get basic user info by ID (Public/Student)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id query int true "User ID"
// @Success 200 {object} utils.Response{data=object{id=int,full_name=string}}
// @Failure 404 {object} utils.Response
// @Router /mahasiswa/users/lookup [get]
func (h *UserHandler) LookupUser(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	user, err := h.service.GetUserByID(uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", nil)
		return
	}

	// Return only safe info
	utils.SuccessResponse(c, http.StatusOK, "User found", gin.H{
		"id":        user.ID,
		"full_name": user.FullName,
		"role":      user.Role,
	})
}

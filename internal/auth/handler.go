package auth

import (
	"net/http"
	"wallet-point/internal/audit"
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service      *AuthService
	auditService *audit.AuditService
}

func NewAuthHandler(service *AuthService, auditService *audit.AuditService) *AuthHandler {
	return &AuthHandler{service: service, auditService: auditService}
}

// Login handles user login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} utils.Response{data=LoginResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)

	// Log activity
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    response.User.ID,
		Action:    "LOGIN",
		Entity:    "USER",
		EntityID:  response.User.ID,
		Details:   "User logged in successfully",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// Register handles user registration (admin only)
// @Summary Admin register new user
// @Description Create a new user account (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RegisterRequest true "User details"
// @Success 201 {object} utils.Response{data=User}
// @Failure 400 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Router /admin/users [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	user, err := h.service.Register(&req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "email already registered" || err.Error() == "NIM/NIP already registered" {
			statusCode = http.StatusConflict
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", user)

	// Log activity
	adminID := c.GetUint("user_id")
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    adminID,
		Action:    "REGISTER",
		Entity:    "USER",
		EntityID:  user.ID,
		Details:   "Admin registered new user: " + user.Email,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// PublicRegister handles self registration
// @Summary Self registration
// @Description Create a new mahasiswa account (Public)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body PublicRegisterRequest true "User details"
// @Success 201 {object} utils.Response{data=User}
// @Failure 400 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Router /auth/register [post]
func (h *AuthHandler) PublicRegister(c *gin.Context) {
	var req PublicRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	user, err := h.service.PublicRegister(&req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "email already registered" || err.Error() == "NIM/NIP already registered" {
			statusCode = http.StatusConflict
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Rekening berhasil dibuat. Silakan login.", user)

	// Log activity (without UserID as they are not logged in yet)
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    user.ID, // Log with the new user's ID
		Action:    "SELF_REGISTER",
		Entity:    "USER",
		EntityID:  user.ID,
		Details:   "New user self-registered: " + user.Email,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// Me handles get current user profile
// @Summary Get current user
// @Description Get authenticated user profile
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=User}
// @Failure 401 {object} utils.Response
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetUint("user_id")

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", user)
}

// UpdateProfile handles user profile update
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	user, err := h.service.UpdateProfile(userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile updated successfully", user)

	// Log activity
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    userID,
		Action:    "UPDATE_PROFILE",
		Entity:    "USER",
		EntityID:  userID,
		Details:   "User updated their own profile",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// UpdatePassword handles user password change
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := h.service.UpdatePassword(userID, &req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password updated successfully", nil)

	// Log activity
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    userID,
		Action:    "UPDATE_PASSWORD",
		Entity:    "USER",
		EntityID:  userID,
		Details:   "User changed their own password",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

func (h *AuthHandler) UpdatePin(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req UpdatePinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := h.service.UpdatePIN(userID, &req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "PIN updated successfully", nil)

	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    userID,
		Action:    "UPDATE_PIN",
		Entity:    "USER",
		EntityID:  userID,
		Details:   "User updated transaction PIN",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

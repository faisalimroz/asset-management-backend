package handlers

import (
	"net/http"

	"asset-management-backend/internal/models"
	"asset-management-backend/internal/services"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid request payload", err.Error())
		return
	}

	ip := c.ClientIP()
	user, err := h.authService.Register(&req, ip)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to register user", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "user registered successfully", user)
}

// Login handles user login and JWT issuance
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid login credentials payload", err.Error())
		return
	}

	res, err := h.authService.Login(&req)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusUnauthorized, "login failed", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "login successful", res)
}

// GetProfile returns current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	profile, err := h.authService.GetProfile(userID.(string))
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to get profile", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "profile retrieved successfully", profile)
}

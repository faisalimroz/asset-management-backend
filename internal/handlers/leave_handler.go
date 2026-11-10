package handlers

import (
	"net/http"

	"asset-management-backend/internal/middleware"
	"asset-management-backend/internal/models"
	"asset-management-backend/internal/services"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// LeaveHandler handles leave request HTTP endpoints
type LeaveHandler struct {
	leaveService services.LeaveService
}

// NewLeaveHandler creates a new LeaveHandler instance
func NewLeaveHandler(leaveService services.LeaveService) *LeaveHandler {
	return &LeaveHandler{leaveService: leaveService}
}

// CreateLeave handles an employee submitting a leave request
func (h *LeaveHandler) CreateLeave(c *gin.Context) {
	var req models.CreateLeaveRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid leave request payload", err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	userEmail, _ := c.Get("email")
	ip := c.ClientIP()

	leave, err := h.leaveService.CreateLeaveRequest(userID.(string), userEmail.(string), ip, &req)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to submit leave request", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "leave request submitted successfully", leave)
}

// ListLeaves lists leave requests (HR views all; Employee views own)
func (h *LeaveHandler) ListLeaves(c *gin.Context) {
	var filter models.LeaveFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid query parameters", err.Error())
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		utils.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	leaves, total, err := h.leaveService.ListLeaves(filter, currentUser)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to list leave requests", err.Error())
		return
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 10
	}

	utils.SendPaginated(c, http.StatusOK, "leave requests retrieved successfully", leaves, page, limit, total)
}

// GetLeave gets a single leave request by ID
func (h *LeaveHandler) GetLeave(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendError(c, http.StatusBadRequest, "leave ID is required", nil)
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		utils.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	leave, err := h.leaveService.GetLeaveByID(id, currentUser)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to retrieve leave request", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "leave request retrieved successfully", leave)
}

// UpdateLeaveStatus handles approving or rejecting a leave request (HR only)
func (h *LeaveHandler) UpdateLeaveStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendError(c, http.StatusBadRequest, "leave ID is required", nil)
		return
	}

	var req models.UpdateLeaveStatusDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid status update payload", err.Error())
		return
	}

	hrID, _ := c.Get("user_id")
	hrEmail, _ := c.Get("email")
	ip := c.ClientIP()

	leave, err := h.leaveService.UpdateLeaveStatus(id, &req, hrID.(string), hrEmail.(string), ip)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to update leave status", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "leave status updated successfully", leave)
}

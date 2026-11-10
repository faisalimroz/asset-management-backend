package handlers

import (
	"net/http"

	"asset-management-backend/internal/models"
	"asset-management-backend/internal/services"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AssignmentHandler handles asset assignment HTTP endpoints
type AssignmentHandler struct {
	assignmentService services.AssignmentService
}

// NewAssignmentHandler creates a new AssignmentHandler instance
func NewAssignmentHandler(assignmentService services.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{assignmentService: assignmentService}
}

// AssignAsset handles allocating an asset to an employee (IT only)
func (h *AssignmentHandler) AssignAsset(c *gin.Context) {
	var req models.AssignAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid assignment payload", err.Error())
		return
	}

	actorID, _ := c.Get("user_id")
	actorEmail, _ := c.Get("email")
	ip := c.ClientIP()

	assignment, err := h.assignmentService.AssignAsset(&req, actorID.(string), actorEmail.(string), ip)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to assign asset", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "asset assigned successfully", assignment)
}

// ReturnAsset handles marking an asset returned (IT and HR)
func (h *AssignmentHandler) ReturnAsset(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendError(c, http.StatusBadRequest, "assignment ID is required", nil)
		return
	}

	var req models.ReturnAssetRequest
	// Note: body is optional for returns (e.g. { "returned_status": "available", "notes": "..." })
	_ = c.ShouldBindJSON(&req)

	actorID, _ := c.Get("user_id")
	actorEmail, _ := c.Get("email")
	ip := c.ClientIP()

	assignment, err := h.assignmentService.ReturnAsset(id, &req, actorID.(string), actorEmail.(string), ip)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to return asset", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "asset marked as returned successfully", assignment)
}

// GetMyAssets handles an employee retrieving their allocated assets (active and past)
func (h *AssignmentHandler) GetMyAssets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	assignments, err := h.assignmentService.GetMyAssets(userID.(string))
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to fetch user assets", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "employee assets retrieved successfully", assignments)
}

// ListAssignments handles listing all asset assignments (IT & HR)
func (h *AssignmentHandler) ListAssignments(c *gin.Context) {
	var filter models.AssignmentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid query parameters", err.Error())
		return
	}

	assignments, total, err := h.assignmentService.ListAssignments(filter)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to list assignments", err.Error())
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

	utils.SendPaginated(c, http.StatusOK, "assignments retrieved successfully", assignments, page, limit, total)
}

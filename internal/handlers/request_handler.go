package handlers

import (
	"net/http"

	"asset-management-backend/internal/middleware"
	"asset-management-backend/internal/models"
	"asset-management-backend/internal/services"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RequestHandler handles asset request HTTP endpoints
type RequestHandler struct {
	requestService services.RequestService
}

// NewRequestHandler creates a new RequestHandler instance
func NewRequestHandler(requestService services.RequestService) *RequestHandler {
	return &RequestHandler{requestService: requestService}
}

// CreateRequest handles an employee submitting an asset request
func (h *RequestHandler) CreateRequest(c *gin.Context) {
	var req models.CreateAssetRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid asset request payload", err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	userEmail, _ := c.Get("email")
	ip := c.ClientIP()

	request, err := h.requestService.CreateRequest(&req, userID.(string), userEmail.(string), ip)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to submit asset request", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "asset request submitted successfully", request)
}

// ListRequests lists asset requests (IT/HR sees all; Employee sees their own)
func (h *RequestHandler) ListRequests(c *gin.Context) {
	var filter models.RequestFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid query parameters", err.Error())
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		utils.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	requests, total, err := h.requestService.ListRequests(filter, currentUser)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to list asset requests", err.Error())
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

	utils.SendPaginated(c, http.StatusOK, "asset requests retrieved successfully", requests, page, limit, total)
}

// GetRequest gets a single asset request by ID
func (h *RequestHandler) GetRequest(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendError(c, http.StatusBadRequest, "request ID is required", nil)
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		utils.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	req, err := h.requestService.GetRequestByID(id, currentUser)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to retrieve request", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "asset request retrieved successfully", req)
}

// UpdateRequestStatus handles approving/rejecting a request (IT & HR)
func (h *RequestHandler) UpdateRequestStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendError(c, http.StatusBadRequest, "request ID is required", nil)
		return
	}

	var req models.UpdateRequestStatusDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid status update payload", err.Error())
		return
	}

	reviewerID, _ := c.Get("user_id")
	reviewerEmail, _ := c.Get("email")
	ip := c.ClientIP()

	updatedReq, err := h.requestService.UpdateRequestStatus(id, &req, reviewerID.(string), reviewerEmail.(string), ip)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to update request status", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "request status updated successfully", updatedReq)
}

package handlers

import (
	"net/http"
	"strconv"

	"asset-management-backend/internal/services"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP endpoints for HR/IT
type UserHandler struct {
	reportService services.ReportService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(reportService services.ReportService) *UserHandler {
	return &UserHandler{reportService: reportService}
}

// ListEmployees lists employees with their active assigned asset counts (HR & IT)
func (h *UserHandler) ListEmployees(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	employees, total, err := h.reportService.GetEmployeeAllocations(page, limit)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to list employees", err.Error())
		return
	}

	utils.SendPaginated(c, http.StatusOK, "employees retrieved successfully", employees, page, limit, total)
}

// GetEmployeeAssets lists all assets allocated to a specific employee (HR & IT)
func (h *UserHandler) GetEmployeeAssets(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.SendError(c, http.StatusBadRequest, "user ID is required", nil)
		return
	}

	assignments, err := h.reportService.GetEmployeeAssetHistory(userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to get employee assets", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "employee assets retrieved successfully", assignments)
}

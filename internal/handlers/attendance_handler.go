package handlers

import (
	"net/http"
	"strconv"

	"asset-management-backend/internal/models"
	"asset-management-backend/internal/services"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AttendanceHandler handles attendance HTTP endpoints
type AttendanceHandler struct {
	attendanceService services.AttendanceService
}

// NewAttendanceHandler creates a new AttendanceHandler instance
func NewAttendanceHandler(attendanceService services.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{attendanceService: attendanceService}
}

// CheckIn records daily check-in (Employee)
func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	var req models.CheckInRequest
	_ = c.ShouldBindJSON(&req) // Body is optional (can check-in with empty JSON)

	userID, _ := c.Get("user_id")
	userEmail, _ := c.Get("email")
	ip := c.ClientIP()

	log, err := h.attendanceService.CheckIn(userID.(string), userEmail.(string), ip, &req)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to record check-in", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "checked in successfully", log)
}

// CheckOut records daily check-out and computes work hours (Employee)
func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	var req models.CheckOutRequest
	_ = c.ShouldBindJSON(&req) // Body is optional

	userID, _ := c.Get("user_id")
	userEmail, _ := c.Get("email")
	ip := c.ClientIP()

	log, err := h.attendanceService.CheckOut(userID.(string), userEmail.(string), ip, &req)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to record check-out", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "checked out successfully", log)
}

// GetMyLogs returns personal attendance history (Employee)
func (h *AttendanceHandler) GetMyLogs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, total, err := h.attendanceService.GetMyLogs(userID.(string), page, limit)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to fetch attendance logs", err.Error())
		return
	}

	utils.SendPaginated(c, http.StatusOK, "attendance logs retrieved successfully", logs, page, limit, total)
}

// GetAttendanceReport returns company-wide attendance reports with filters (HR only)
func (h *AttendanceHandler) GetAttendanceReport(c *gin.Context) {
	var filter models.AttendanceFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid query parameters", err.Error())
		return
	}

	logs, total, err := h.attendanceService.GetAttendanceReport(filter)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to generate attendance report", err.Error())
		return
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}

	utils.SendPaginated(c, http.StatusOK, "attendance report generated successfully", logs, page, limit, total)
}

package handlers

import (
	"net/http"

	"asset-management-backend/internal/models"
	"asset-management-backend/internal/services"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ReportHandler handles inventory statistics and audit log endpoints
type ReportHandler struct {
	reportService services.ReportService
}

// NewReportHandler creates a new ReportHandler instance
func NewReportHandler(reportService services.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// GetInventorySummary returns metrics on assets (IT & HR)
func (h *ReportHandler) GetInventorySummary(c *gin.Context) {
	summary, err := h.reportService.GetInventorySummary()
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to get inventory summary", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "inventory summary retrieved successfully", summary)
}

// GetAuditLogs returns system-wide audit logs (IT only)
func (h *ReportHandler) GetAuditLogs(c *gin.Context) {
	var filter models.AuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid query parameters", err.Error())
		return
	}

	logs, total, err := h.reportService.GetAuditLogs(filter)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to retrieve audit logs", err.Error())
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

	utils.SendPaginated(c, http.StatusOK, "audit logs retrieved successfully", logs, page, limit, total)
}

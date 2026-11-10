package handlers

import (
	"net/http"

	"asset-management-backend/internal/models"
	"asset-management-backend/internal/services"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AssetHandler handles asset-related HTTP requests
type AssetHandler struct {
	assetService services.AssetService
}

// NewAssetHandler creates a new AssetHandler instance
func NewAssetHandler(assetService services.AssetService) *AssetHandler {
	return &AssetHandler{assetService: assetService}
}

// CreateAsset handles adding a new asset (IT only)
func (h *AssetHandler) CreateAsset(c *gin.Context) {
	var req models.CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid asset payload", err.Error())
		return
	}

	actorID, _ := c.Get("user_id")
	actorEmail, _ := c.Get("email")
	ip := c.ClientIP()

	asset, err := h.assetService.CreateAsset(&req, actorID.(string), actorEmail.(string), ip)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to create asset", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "asset created successfully", asset)
}

// GetAsset handles retrieving a single asset and its history
func (h *AssetHandler) GetAsset(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendError(c, http.StatusBadRequest, "asset ID is required", nil)
		return
	}

	asset, err := h.assetService.GetAssetByID(id)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to get asset", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "asset retrieved successfully", asset)
}

// ListAssets handles listing assets with optional filters
func (h *AssetHandler) ListAssets(c *gin.Context) {
	var filter models.AssetFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid query parameters", err.Error())
		return
	}

	assets, total, err := h.assetService.ListAssets(filter)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to list assets", err.Error())
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

	utils.SendPaginated(c, http.StatusOK, "assets retrieved successfully", assets, page, limit, total)
}

// UpdateAsset handles updating an existing asset (IT only)
func (h *AssetHandler) UpdateAsset(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendError(c, http.StatusBadRequest, "asset ID is required", nil)
		return
	}

	var req models.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid update payload", err.Error())
		return
	}

	actorID, _ := c.Get("user_id")
	actorEmail, _ := c.Get("email")
	ip := c.ClientIP()

	asset, err := h.assetService.UpdateAsset(id, &req, actorID.(string), actorEmail.(string), ip)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to update asset", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "asset updated successfully", asset)
}

// DeleteAsset handles retiring or deleting an asset (IT only)
func (h *AssetHandler) DeleteAsset(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendError(c, http.StatusBadRequest, "asset ID is required", nil)
		return
	}

	actorID, _ := c.Get("user_id")
	actorEmail, _ := c.Get("email")
	ip := c.ClientIP()

	err := h.assetService.DeleteAsset(id, actorID.(string), actorEmail.(string), ip)
	if err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			utils.SendError(c, appErr.Code, appErr.Message, nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to delete asset", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "asset retired and deleted successfully", nil)
}

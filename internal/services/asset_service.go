package services

import (
	"errors"
	"fmt"

	"asset-management-backend/internal/models"
	"asset-management-backend/internal/repository"
	"asset-management-backend/pkg/utils"

	"gorm.io/gorm"
)

// AssetService interface defines asset business operations
type AssetService interface {
	CreateAsset(req *models.CreateAssetRequest, actorID, actorEmail, ip string) (*models.Asset, error)
	GetAssetByID(id string) (*models.Asset, error)
	ListAssets(filter models.AssetFilter) ([]models.Asset, int64, error)
	UpdateAsset(id string, req *models.UpdateAssetRequest, actorID, actorEmail, ip string) (*models.Asset, error)
	DeleteAsset(id string, actorID, actorEmail, ip string) error
}

type assetService struct {
	assetRepo repository.AssetRepository
	auditRepo repository.AuditRepository
}

// NewAssetService creates a new AssetService instance
func NewAssetService(assetRepo repository.AssetRepository, auditRepo repository.AuditRepository) AssetService {
	return &assetService{
		assetRepo: assetRepo,
		auditRepo: auditRepo,
	}
}

func (s *assetService) CreateAsset(req *models.CreateAssetRequest, actorID, actorEmail, ip string) (*models.Asset, error) {
	// Verify unique serial number
	existing, err := s.assetRepo.FindBySerialNumber(req.SerialNumber)
	if err == nil && existing != nil {
		return nil, utils.NewConflictError(fmt.Sprintf("asset with serial number '%s' already exists", req.SerialNumber))
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.NewInternalServerError("failed to check serial number uniqueness", err)
	}

	status := req.Status
	if status == "" {
		status = models.AssetStatusAvailable
	}

	asset := &models.Asset{
		Name:           req.Name,
		SerialNumber:   req.SerialNumber,
		Category:       req.Category,
		Status:         status,
		PurchaseDate:   req.PurchaseDate,
		PurchaseCost:   req.PurchaseCost,
		Specifications: req.Specifications,
	}

	if err := s.assetRepo.Create(asset); err != nil {
		return nil, utils.NewInternalServerError("failed to create asset", err)
	}

	// Audit log
	s.auditRepo.Create(&models.AuditLog{
		UserID:    actorID,
		UserEmail: actorEmail,
		Action:    models.ActionAssetCreated,
		Entity:    "asset",
		EntityID:  asset.ID,
		Details:   fmt.Sprintf("Created asset %s (SN: %s, Category: %s)", asset.Name, asset.SerialNumber, asset.Category),
		IPAddress: ip,
	})

	return asset, nil
}

func (s *assetService) GetAssetByID(id string) (*models.Asset, error) {
	asset, err := s.assetRepo.FindByIDWithHistory(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError(fmt.Sprintf("asset with ID '%s' not found", id))
		}
		return nil, utils.NewInternalServerError("failed to retrieve asset details", err)
	}
	return asset, nil
}

func (s *assetService) ListAssets(filter models.AssetFilter) ([]models.Asset, int64, error) {
	return s.assetRepo.List(filter)
}

func (s *assetService) UpdateAsset(id string, req *models.UpdateAssetRequest, actorID, actorEmail, ip string) (*models.Asset, error) {
	asset, err := s.assetRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError(fmt.Sprintf("asset with ID '%s' not found", id))
		}
		return nil, utils.NewInternalServerError("failed to find asset for update", err)
	}

	// If serial number is being updated, verify uniqueness
	if req.SerialNumber != nil && *req.SerialNumber != asset.SerialNumber {
		existing, err := s.assetRepo.FindBySerialNumber(*req.SerialNumber)
		if err == nil && existing != nil && existing.ID != id {
			return nil, utils.NewConflictError(fmt.Sprintf("asset with serial number '%s' already exists", *req.SerialNumber))
		}
		asset.SerialNumber = *req.SerialNumber
	}

	if req.Name != nil {
		asset.Name = *req.Name
	}
	if req.Category != nil {
		asset.Category = *req.Category
	}
	if req.Status != nil {
		asset.Status = *req.Status
	}
	if req.PurchaseDate != nil {
		asset.PurchaseDate = *req.PurchaseDate
	}
	if req.PurchaseCost != nil {
		asset.PurchaseCost = *req.PurchaseCost
	}
	if req.Specifications != nil {
		asset.Specifications = *req.Specifications
	}

	if err := s.assetRepo.Update(asset); err != nil {
		return nil, utils.NewInternalServerError("failed to update asset", err)
	}

	// Audit log
	s.auditRepo.Create(&models.AuditLog{
		UserID:    actorID,
		UserEmail: actorEmail,
		Action:    models.ActionAssetUpdated,
		Entity:    "asset",
		EntityID:  asset.ID,
		Details:   fmt.Sprintf("Updated asset %s (Status: %s)", asset.Name, asset.Status),
		IPAddress: ip,
	})

	return asset, nil
}

func (s *assetService) DeleteAsset(id string, actorID, actorEmail, ip string) error {
	asset, err := s.assetRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewNotFoundError(fmt.Sprintf("asset with ID '%s' not found", id))
		}
		return utils.NewInternalServerError("failed to find asset for deletion", err)
	}

	// Mark status as retired if deleting
	asset.Status = models.AssetStatusRetired
	s.assetRepo.Update(asset)

	if err := s.assetRepo.Delete(id); err != nil {
		return utils.NewInternalServerError("failed to delete asset", err)
	}

	// Audit log
	s.auditRepo.Create(&models.AuditLog{
		UserID:    actorID,
		UserEmail: actorEmail,
		Action:    models.ActionAssetDeleted,
		Entity:    "asset",
		EntityID:  id,
		Details:   fmt.Sprintf("Retired and deleted asset %s (SN: %s)", asset.Name, asset.SerialNumber),
		IPAddress: ip,
	})

	return nil
}

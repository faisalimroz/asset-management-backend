package services

import (
	"errors"
	"fmt"
	"time"

	"asset-management-backend/internal/models"
	"asset-management-backend/internal/repository"
	"asset-management-backend/pkg/utils"

	"gorm.io/gorm"
)

// AssignmentService interface defines asset allocation operations
type AssignmentService interface {
	AssignAsset(req *models.AssignAssetRequest, assignedByID, assignedByEmail, ip string) (*models.AssetAssignment, error)
	ReturnAsset(assignmentID string, req *models.ReturnAssetRequest, returnedByID, returnedByEmail, ip string) (*models.AssetAssignment, error)
	GetMyAssets(userID string) ([]models.AssetAssignment, error)
	ListAssignments(filter models.AssignmentFilter) ([]models.AssetAssignment, int64, error)
}

type assignmentService struct {
	assignmentRepo repository.AssignmentRepository
	assetRepo      repository.AssetRepository
	userRepo       repository.UserRepository
	auditRepo      repository.AuditRepository
}

// NewAssignmentService creates a new AssignmentService instance
func NewAssignmentService(
	assignmentRepo repository.AssignmentRepository,
	assetRepo repository.AssetRepository,
	userRepo repository.UserRepository,
	auditRepo repository.AuditRepository,
) AssignmentService {
	return &assignmentService{
		assignmentRepo: assignmentRepo,
		assetRepo:      assetRepo,
		userRepo:       userRepo,
		auditRepo:      auditRepo,
	}
}

func (s *assignmentService) AssignAsset(req *models.AssignAssetRequest, assignedByID, assignedByEmail, ip string) (*models.AssetAssignment, error) {
	// Verify user exists
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError(fmt.Sprintf("user with ID '%s' not found", req.UserID))
		}
		return nil, utils.NewInternalServerError("failed to verify employee", err)
	}

	// Verify asset exists
	asset, err := s.assetRepo.FindByID(req.AssetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError(fmt.Sprintf("asset with ID '%s' not found", req.AssetID))
		}
		return nil, utils.NewInternalServerError("failed to verify asset", err)
	}

	// Verify asset is available for assignment
	if asset.Status != models.AssetStatusAvailable {
		return nil, utils.NewBadRequestError(fmt.Sprintf("asset '%s' cannot be assigned because its current status is '%s'", asset.Name, asset.Status))
	}

	// Check if there is an existing active assignment for this asset
	activeAssignment, err := s.assignmentRepo.FindActiveByAssetID(req.AssetID)
	if err == nil && activeAssignment != nil {
		return nil, utils.NewConflictError("asset already has an active assignment")
	}

	// Create assignment
	assignment := &models.AssetAssignment{
		AssetID:    req.AssetID,
		UserID:     req.UserID,
		AssignedBy: assignedByID,
		AssignedAt: time.Now(),
		Notes:      req.Notes,
	}

	if err := s.assignmentRepo.Create(assignment); err != nil {
		return nil, utils.NewInternalServerError("failed to create assignment record", err)
	}

	// Update asset status to assigned
	asset.Status = models.AssetStatusAssigned
	if err := s.assetRepo.Update(asset); err != nil {
		return nil, utils.NewInternalServerError("failed to update asset status to assigned", err)
	}

	// Reload with relations
	fullAssignment, _ := s.assignmentRepo.FindByID(assignment.ID)

	// Audit log
	s.auditRepo.Create(&models.AuditLog{
		UserID:    assignedByID,
		UserEmail: assignedByEmail,
		Action:    models.ActionAssetAssigned,
		Entity:    "assignment",
		EntityID:  assignment.ID,
		Details:   fmt.Sprintf("Assigned asset %s (SN: %s) to %s (%s)", asset.Name, asset.SerialNumber, user.Name, user.Email),
		IPAddress: ip,
	})

	return fullAssignment, nil
}

func (s *assignmentService) ReturnAsset(assignmentID string, req *models.ReturnAssetRequest, returnedByID, returnedByEmail, ip string) (*models.AssetAssignment, error) {
	assignment, err := s.assignmentRepo.FindByID(assignmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError(fmt.Sprintf("assignment with ID '%s' not found", assignmentID))
		}
		return nil, utils.NewInternalServerError("failed to find assignment", err)
	}

	if assignment.ReturnedAt != nil {
		return nil, utils.NewBadRequestError("this asset assignment has already been marked as returned")
	}

	now := time.Now()
	assignment.ReturnedAt = &now
	if req.Notes != "" {
		if assignment.Notes != "" {
			assignment.Notes += "\n[Return Notes]: " + req.Notes
		} else {
			assignment.Notes = "[Return Notes]: " + req.Notes
		}
	}

	if err := s.assignmentRepo.Update(assignment); err != nil {
		return nil, utils.NewInternalServerError("failed to update assignment return", err)
	}

	// Update asset status
	asset, err := s.assetRepo.FindByID(assignment.AssetID)
	if err == nil && asset != nil {
		status := models.AssetStatusAvailable
		if req.ReturnedStatus != "" {
			status = req.ReturnedStatus
		}
		asset.Status = status
		s.assetRepo.Update(asset)
	}

	// Reload full assignment
	fullAssignment, _ := s.assignmentRepo.FindByID(assignmentID)

	// Audit log
	s.auditRepo.Create(&models.AuditLog{
		UserID:    returnedByID,
		UserEmail: returnedByEmail,
		Action:    models.ActionAssetReturned,
		Entity:    "assignment",
		EntityID:  assignment.ID,
		Details:   fmt.Sprintf("Asset %s marked as returned (Status set to %s)", assignment.AssetID, asset.Status),
		IPAddress: ip,
	})

	return fullAssignment, nil
}

func (s *assignmentService) GetMyAssets(userID string) ([]models.AssetAssignment, error) {
	return s.assignmentRepo.ListByUserID(userID)
}

func (s *assignmentService) ListAssignments(filter models.AssignmentFilter) ([]models.AssetAssignment, int64, error) {
	return s.assignmentRepo.List(filter)
}

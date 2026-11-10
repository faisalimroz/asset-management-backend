package services

import (
	"errors"
	"fmt"

	"asset-management-backend/internal/models"
	"asset-management-backend/internal/repository"
	"asset-management-backend/pkg/utils"

	"gorm.io/gorm"
)

// RequestService interface defines asset request operations
type RequestService interface {
	CreateRequest(req *models.CreateAssetRequestDTO, userID, userEmail, ip string) (*models.AssetRequest, error)
	ListRequests(filter models.RequestFilter, currentUser *models.User) ([]models.AssetRequest, int64, error)
	GetRequestByID(id string, currentUser *models.User) (*models.AssetRequest, error)
	UpdateRequestStatus(requestID string, req *models.UpdateRequestStatusDTO, reviewerID, reviewerEmail, ip string) (*models.AssetRequest, error)
}

type requestService struct {
	requestRepo repository.RequestRepository
	auditRepo   repository.AuditRepository
}

// NewRequestService creates a new RequestService instance
func NewRequestService(requestRepo repository.RequestRepository, auditRepo repository.AuditRepository) RequestService {
	return &requestService{
		requestRepo: requestRepo,
		auditRepo:   auditRepo,
	}
}

func (s *requestService) CreateRequest(req *models.CreateAssetRequestDTO, userID, userEmail, ip string) (*models.AssetRequest, error) {
	assetReq := &models.AssetRequest{
		UserID:        userID,
		AssetCategory: req.AssetCategory,
		ItemRequested: req.ItemRequested,
		Reason:        req.Reason,
		Status:        models.RequestStatusPending,
	}

	if err := s.requestRepo.Create(assetReq); err != nil {
		return nil, utils.NewInternalServerError("failed to submit asset request", err)
	}

	// Audit log
	s.auditRepo.Create(&models.AuditLog{
		UserID:    userID,
		UserEmail: userEmail,
		Action:    models.ActionRequestCreated,
		Entity:    "request",
		EntityID:  assetReq.ID,
		Details:   fmt.Sprintf("Requested %s (%s)", req.ItemRequested, req.AssetCategory),
		IPAddress: ip,
	})

	return assetReq, nil
}

func (s *requestService) ListRequests(filter models.RequestFilter, currentUser *models.User) ([]models.AssetRequest, int64, error) {
	// If current user is employee, restrict them to their own requests
	if currentUser.Role == models.RoleEmployee {
		filter.UserID = currentUser.ID
	}

	return s.requestRepo.List(filter)
}

func (s *requestService) GetRequestByID(id string, currentUser *models.User) (*models.AssetRequest, error) {
	req, err := s.requestRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError(fmt.Sprintf("request with ID '%s' not found", id))
		}
		return nil, utils.NewInternalServerError("failed to retrieve asset request", err)
	}

	// If employee, verify ownership
	if currentUser.Role == models.RoleEmployee && req.UserID != currentUser.ID {
		return nil, utils.NewForbiddenError("you do not have permission to view this request")
	}

	return req, nil
}

func (s *requestService) UpdateRequestStatus(requestID string, req *models.UpdateRequestStatusDTO, reviewerID, reviewerEmail, ip string) (*models.AssetRequest, error) {
	assetReq, err := s.requestRepo.FindByID(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError(fmt.Sprintf("request with ID '%s' not found", requestID))
		}
		return nil, utils.NewInternalServerError("failed to find asset request", err)
	}

	assetReq.Status = req.Status
	assetReq.ReviewerID = &reviewerID
	assetReq.ReviewNotes = req.ReviewNotes

	if err := s.requestRepo.Update(assetReq); err != nil {
		return nil, utils.NewInternalServerError("failed to update request status", err)
	}

	// Audit log
	s.auditRepo.Create(&models.AuditLog{
		UserID:    reviewerID,
		UserEmail: reviewerEmail,
		Action:    models.ActionRequestStatusChange,
		Entity:    "request",
		EntityID:  assetReq.ID,
		Details:   fmt.Sprintf("Request status updated to '%s'. Notes: %s", req.Status, req.ReviewNotes),
		IPAddress: ip,
	})

	return assetReq, nil
}

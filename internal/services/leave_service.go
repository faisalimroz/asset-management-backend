package services

import (
	"errors"
	"fmt"

	"asset-management-backend/internal/models"
	"asset-management-backend/internal/repository"
	"asset-management-backend/pkg/utils"

	"gorm.io/gorm"
)

// LeaveService interface defines leave operations
type LeaveService interface {
	CreateLeaveRequest(userID, userEmail, ip string, req *models.CreateLeaveRequestDTO) (*models.LeaveRequest, error)
	ListLeaves(filter models.LeaveFilter, currentUser *models.User) ([]models.LeaveRequest, int64, error)
	GetLeaveByID(id string, currentUser *models.User) (*models.LeaveRequest, error)
	UpdateLeaveStatus(id string, req *models.UpdateLeaveStatusDTO, hrID, hrEmail, ip string) (*models.LeaveRequest, error)
}

type leaveService struct {
	leaveRepo repository.LeaveRepository
	auditRepo repository.AuditRepository
}

// NewLeaveService creates a new LeaveService instance
func NewLeaveService(leaveRepo repository.LeaveRepository, auditRepo repository.AuditRepository) LeaveService {
	return &leaveService{
		leaveRepo: leaveRepo,
		auditRepo: auditRepo,
	}
}

func (s *leaveService) CreateLeaveRequest(userID, userEmail, ip string, req *models.CreateLeaveRequestDTO) (*models.LeaveRequest, error) {
	if req.EndDate.Before(req.StartDate) {
		return nil, utils.NewBadRequestError("end_date cannot be before start_date")
	}

	leave := &models.LeaveRequest{
		UserID:    userID,
		LeaveType: req.LeaveType,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Reason:    req.Reason,
		Status:    models.LeaveStatusPending,
	}

	if err := s.leaveRepo.Create(leave); err != nil {
		return nil, utils.NewInternalServerError("failed to submit leave request", err)
	}

	s.auditRepo.Create(&models.AuditLog{
		UserID:    userID,
		UserEmail: userEmail,
		Action:    models.ActionLeaveRequested,
		Entity:    "leave",
		EntityID:  leave.ID,
		Details:   fmt.Sprintf("Requested %s leave from %s to %s", leave.LeaveType, leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02")),
		IPAddress: ip,
	})

	return leave, nil
}

func (s *leaveService) ListLeaves(filter models.LeaveFilter, currentUser *models.User) ([]models.LeaveRequest, int64, error) {
	// If current user is employee, restrict to own leaves
	if currentUser.Role == models.RoleEmployee {
		filter.UserID = currentUser.ID
	}

	return s.leaveRepo.List(filter)
}

func (s *leaveService) GetLeaveByID(id string, currentUser *models.User) (*models.LeaveRequest, error) {
	leave, err := s.leaveRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError(fmt.Sprintf("leave request with ID '%s' not found", id))
		}
		return nil, utils.NewInternalServerError("failed to find leave request", err)
	}

	if currentUser.Role == models.RoleEmployee && leave.UserID != currentUser.ID {
		return nil, utils.NewForbiddenError("you do not have permission to view this leave request")
	}

	return leave, nil
}

func (s *leaveService) UpdateLeaveStatus(id string, req *models.UpdateLeaveStatusDTO, hrID, hrEmail, ip string) (*models.LeaveRequest, error) {
	leave, err := s.leaveRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError(fmt.Sprintf("leave request with ID '%s' not found", id))
		}
		return nil, utils.NewInternalServerError("failed to find leave request", err)
	}

	leave.Status = req.Status
	leave.ApprovedBy = &hrID
	leave.ReviewNotes = req.ReviewNotes

	if err := s.leaveRepo.Update(leave); err != nil {
		return nil, utils.NewInternalServerError("failed to update leave status", err)
	}

	s.auditRepo.Create(&models.AuditLog{
		UserID:    hrID,
		UserEmail: hrEmail,
		Action:    models.ActionLeaveStatusChange,
		Entity:    "leave",
		EntityID:  leave.ID,
		Details:   fmt.Sprintf("Leave status updated to '%s' by HR. Notes: %s", req.Status, req.ReviewNotes),
		IPAddress: ip,
	})

	return leave, nil
}

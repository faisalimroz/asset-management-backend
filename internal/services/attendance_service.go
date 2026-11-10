package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"asset-management-backend/internal/models"
	"asset-management-backend/internal/repository"
	"asset-management-backend/pkg/utils"

	"gorm.io/gorm"
)

// AttendanceService interface defines attendance operations
type AttendanceService interface {
	CheckIn(userID, userEmail, ip string, req *models.CheckInRequest) (*models.AttendanceLog, error)
	CheckOut(userID, userEmail, ip string, req *models.CheckOutRequest) (*models.AttendanceLog, error)
	GetMyLogs(userID string, page, limit int) ([]models.AttendanceLog, int64, error)
	GetAttendanceReport(filter models.AttendanceFilter) ([]models.AttendanceLog, int64, error)
}

type attendanceService struct {
	attendanceRepo repository.AttendanceRepository
	auditRepo      repository.AuditRepository
}

// NewAttendanceService creates a new AttendanceService instance
func NewAttendanceService(attendanceRepo repository.AttendanceRepository, auditRepo repository.AuditRepository) AttendanceService {
	return &attendanceService{
		attendanceRepo: attendanceRepo,
		auditRepo:      auditRepo,
	}
}

func (s *attendanceService) CheckIn(userID, userEmail, ip string, req *models.CheckInRequest) (*models.AttendanceLog, error) {
	today := time.Now().Format("2006-01-02")
	existing, err := s.attendanceRepo.FindByUserAndDate(userID, today)

	now := time.Now()
	status := models.AttendanceStatusPresent
	if req != nil && req.Status != "" {
		status = req.Status
	}

	notes := ""
	if req != nil {
		notes = req.Notes
	}

	if err == nil && existing != nil {
		if existing.CheckInTime != nil {
			return nil, utils.NewConflictError("you have already checked in for today (" + today + ")")
		}
		// If an existing record was pre-created (e.g. marked absent/leave), update it with check-in
		existing.CheckInTime = &now
		existing.Status = status
		if notes != "" {
			existing.Notes = notes
		}
		if err := s.attendanceRepo.Update(existing); err != nil {
			return nil, utils.NewInternalServerError("failed to update check-in", err)
		}

		s.auditRepo.Create(&models.AuditLog{
			UserID:    userID,
			UserEmail: userEmail,
			Action:    models.ActionAttendanceCheckIn,
			Entity:    "attendance",
			EntityID:  existing.ID,
			Details:   fmt.Sprintf("Checked in at %s with status %s", now.Format(time.RFC3339), status),
			IPAddress: ip,
		})
		return existing, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.NewInternalServerError("failed to check attendance status", err)
	}

	// Create new attendance record
	log := &models.AttendanceLog{
		UserID:      userID,
		Date:        today,
		CheckInTime: &now,
		Status:      status,
		Notes:       notes,
	}

	if err := s.attendanceRepo.Create(log); err != nil {
		return nil, utils.NewInternalServerError("failed to record check-in", err)
	}

	s.auditRepo.Create(&models.AuditLog{
		UserID:    userID,
		UserEmail: userEmail,
		Action:    models.ActionAttendanceCheckIn,
		Entity:    "attendance",
		EntityID:  log.ID,
		Details:   fmt.Sprintf("Checked in at %s with status %s", now.Format(time.RFC3339), status),
		IPAddress: ip,
	})

	return log, nil
}

func (s *attendanceService) CheckOut(userID, userEmail, ip string, req *models.CheckOutRequest) (*models.AttendanceLog, error) {
	today := time.Now().Format("2006-01-02")
	log, err := s.attendanceRepo.FindByUserAndDate(userID, today)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewBadRequestError("no check-in record found for today (" + today + "). Please check in first.")
		}
		return nil, utils.NewInternalServerError("failed to retrieve attendance record", err)
	}

	if log.CheckInTime == nil {
		return nil, utils.NewBadRequestError("cannot check out without checking in first")
	}

	if log.CheckOutTime != nil {
		return nil, utils.NewConflictError("you have already checked out for today (" + today + ")")
	}

	now := time.Now()
	log.CheckOutTime = &now

	// Calculate working hours
	diff := now.Sub(*log.CheckInTime).Hours()
	if diff < 0 {
		diff = 0
	}
	// Round to 2 decimal places
	log.WorkHours = math.Round(diff*100) / 100

	if req != nil && req.Notes != "" {
		if log.Notes != "" {
			log.Notes += " | Check-out notes: " + req.Notes
		} else {
			log.Notes = "Check-out notes: " + req.Notes
		}
	}

	if err := s.attendanceRepo.Update(log); err != nil {
		return nil, utils.NewInternalServerError("failed to record check-out", err)
	}

	s.auditRepo.Create(&models.AuditLog{
		UserID:    userID,
		UserEmail: userEmail,
		Action:    models.ActionAttendanceCheckOut,
		Entity:    "attendance",
		EntityID:  log.ID,
		Details:   fmt.Sprintf("Checked out at %s. Total work hours: %.2f", now.Format(time.RFC3339), log.WorkHours),
		IPAddress: ip,
	})

	return log, nil
}

func (s *attendanceService) GetMyLogs(userID string, page, limit int) ([]models.AttendanceLog, int64, error) {
	return s.attendanceRepo.ListByUser(userID, page, limit)
}

func (s *attendanceService) GetAttendanceReport(filter models.AttendanceFilter) ([]models.AttendanceLog, int64, error) {
	return s.attendanceRepo.ListReport(filter)
}

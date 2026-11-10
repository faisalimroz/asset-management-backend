package repository

import (
	"asset-management-backend/internal/models"

	"gorm.io/gorm"
)

// AttendanceRepository interface defines data access for attendance records
type AttendanceRepository interface {
	Create(log *models.AttendanceLog) error
	FindByID(id string) (*models.AttendanceLog, error)
	FindByUserAndDate(userID, date string) (*models.AttendanceLog, error)
	Update(log *models.AttendanceLog) error
	ListByUser(userID string, page, limit int) ([]models.AttendanceLog, int64, error)
	ListReport(filter models.AttendanceFilter) ([]models.AttendanceLog, int64, error)
}

type attendanceRepository struct {
	db *gorm.DB
}

// NewAttendanceRepository creates a new AttendanceRepository instance
func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) Create(log *models.AttendanceLog) error {
	return r.db.Create(log).Error
}

func (r *attendanceRepository) FindByID(id string) (*models.AttendanceLog, error) {
	var log models.AttendanceLog
	err := r.db.Preload("User").First(&log, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *attendanceRepository) FindByUserAndDate(userID, date string) (*models.AttendanceLog, error) {
	var log models.AttendanceLog
	err := r.db.Preload("User").Where("user_id = ? AND date = ?", userID, date).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *attendanceRepository) Update(log *models.AttendanceLog) error {
	return r.db.Save(log).Error
}

func (r *attendanceRepository) ListByUser(userID string, page, limit int) ([]models.AttendanceLog, int64, error) {
	var logs []models.AttendanceLog
	var total int64

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := r.db.Model(&models.AttendanceLog{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("date DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *attendanceRepository) ListReport(filter models.AttendanceFilter) ([]models.AttendanceLog, int64, error) {
	var logs []models.AttendanceLog
	var total int64

	query := r.db.Model(&models.AttendanceLog{}).Preload("User")

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.Department != "" {
		query = query.Joins("JOIN users ON users.id = attendance_logs.user_id").
			Where("LOWER(users.department) = LOWER(?)", filter.Department)
	}

	if filter.StartDate != "" {
		query = query.Where("date >= ?", filter.StartDate)
	}

	if filter.EndDate != "" {
		query = query.Where("date <= ?", filter.EndDate)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	err := query.Order("date DESC, created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

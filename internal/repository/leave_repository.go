package repository

import (
	"asset-management-backend/internal/models"

	"gorm.io/gorm"
)

// LeaveRepository interface defines data access for leave requests
type LeaveRepository interface {
	Create(leave *models.LeaveRequest) error
	FindByID(id string) (*models.LeaveRequest, error)
	Update(leave *models.LeaveRequest) error
	List(filter models.LeaveFilter) ([]models.LeaveRequest, int64, error)
}

type leaveRepository struct {
	db *gorm.DB
}

// NewLeaveRepository creates a new LeaveRepository instance
func NewLeaveRepository(db *gorm.DB) LeaveRepository {
	return &leaveRepository{db: db}
}

func (r *leaveRepository) Create(leave *models.LeaveRequest) error {
	return r.db.Create(leave).Error
}

func (r *leaveRepository) FindByID(id string) (*models.LeaveRequest, error) {
	var leave models.LeaveRequest
	err := r.db.Preload("User").Preload("Approver").First(&leave, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &leave, nil
}

func (r *leaveRepository) Update(leave *models.LeaveRequest) error {
	return r.db.Save(leave).Error
}

func (r *leaveRepository) List(filter models.LeaveFilter) ([]models.LeaveRequest, int64, error) {
	var leaves []models.LeaveRequest
	var total int64

	query := r.db.Model(&models.LeaveRequest{}).Preload("User").Preload("Approver")

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.LeaveType != "" {
		query = query.Where("leave_type = ?", filter.LeaveType)
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
		limit = 10
	}
	offset := (page - 1) * limit

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&leaves).Error
	return leaves, total, err
}

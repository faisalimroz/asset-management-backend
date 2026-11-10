package repository

import (
	"asset-management-backend/internal/models"

	"gorm.io/gorm"
)

// AuditRepository interface defines data access for audit logs
type AuditRepository interface {
	Create(log *models.AuditLog) error
	List(filter models.AuditLogFilter) ([]models.AuditLog, int64, error)
}

type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository creates a new AuditRepository instance
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditRepository) List(filter models.AuditLogFilter) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.Model(&models.AuditLog{})

	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}

	if filter.Entity != "" {
		query = query.Where("entity = ?", filter.Entity)
	}

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
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

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

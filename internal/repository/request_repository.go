package repository

import (
	"asset-management-backend/internal/models"

	"gorm.io/gorm"
)

// RequestRepository interface defines data access for asset requests
type RequestRepository interface {
	Create(req *models.AssetRequest) error
	FindByID(id string) (*models.AssetRequest, error)
	List(filter models.RequestFilter) ([]models.AssetRequest, int64, error)
	Update(req *models.AssetRequest) error
}

type requestRepository struct {
	db *gorm.DB
}

// NewRequestRepository creates a new RequestRepository instance
func NewRequestRepository(db *gorm.DB) RequestRepository {
	return &requestRepository{db: db}
}

func (r *requestRepository) Create(req *models.AssetRequest) error {
	return r.db.Create(req).Error
}

func (r *requestRepository) FindByID(id string) (*models.AssetRequest, error) {
	var request models.AssetRequest
	err := r.db.
		Preload("User").
		Preload("Reviewer").
		First(&request, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *requestRepository) List(filter models.RequestFilter) ([]models.AssetRequest, int64, error) {
	var requests []models.AssetRequest
	var total int64

	query := r.db.Model(&models.AssetRequest{}).
		Preload("User").
		Preload("Reviewer")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
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
		limit = 10
	}
	offset := (page - 1) * limit

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&requests).Error
	return requests, total, err
}

func (r *requestRepository) Update(req *models.AssetRequest) error {
	return r.db.Save(req).Error
}

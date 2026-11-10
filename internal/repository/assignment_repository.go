package repository

import (
	"asset-management-backend/internal/models"

	"gorm.io/gorm"
)

// AssignmentRepository interface defines data access for assignments
type AssignmentRepository interface {
	Create(assignment *models.AssetAssignment) error
	FindByID(id string) (*models.AssetAssignment, error)
	FindActiveByAssetID(assetID string) (*models.AssetAssignment, error)
	List(filter models.AssignmentFilter) ([]models.AssetAssignment, int64, error)
	ListByUserID(userID string) ([]models.AssetAssignment, error)
	Update(assignment *models.AssetAssignment) error
}

type assignmentRepository struct {
	db *gorm.DB
}

// NewAssignmentRepository creates a new AssignmentRepository instance
func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	return &assignmentRepository{db: db}
}

func (r *assignmentRepository) Create(assignment *models.AssetAssignment) error {
	return r.db.Create(assignment).Error
}

func (r *assignmentRepository) FindByID(id string) (*models.AssetAssignment, error) {
	var assignment models.AssetAssignment
	err := r.db.
		Preload("Asset").
		Preload("User").
		Preload("Assigner").
		First(&assignment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *assignmentRepository) FindActiveByAssetID(assetID string) (*models.AssetAssignment, error) {
	var assignment models.AssetAssignment
	err := r.db.
		Preload("Asset").
		Preload("User").
		Where("asset_id = ? AND returned_at IS NULL", assetID).
		First(&assignment).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *assignmentRepository) List(filter models.AssignmentFilter) ([]models.AssetAssignment, int64, error) {
	var assignments []models.AssetAssignment
	var total int64

	query := r.db.Model(&models.AssetAssignment{}).
		Preload("Asset").
		Preload("User").
		Preload("Assigner")

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.AssetID != "" {
		query = query.Where("asset_id = ?", filter.AssetID)
	}

	if filter.ActiveOnly {
		query = query.Where("returned_at IS NULL")
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

	err := query.Order("assigned_at DESC").Limit(limit).Offset(offset).Find(&assignments).Error
	return assignments, total, err
}

func (r *assignmentRepository) ListByUserID(userID string) ([]models.AssetAssignment, error) {
	var assignments []models.AssetAssignment
	err := r.db.
		Preload("Asset").
		Preload("Assigner").
		Where("user_id = ?", userID).
		Order("assigned_at DESC").
		Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepository) Update(assignment *models.AssetAssignment) error {
	return r.db.Save(assignment).Error
}

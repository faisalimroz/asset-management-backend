package repository

import (
	"strings"

	"asset-management-backend/internal/models"

	"gorm.io/gorm"
)

// AssetRepository interface defines data access for assets
type AssetRepository interface {
	Create(asset *models.Asset) error
	FindByID(id string) (*models.Asset, error)
	FindByIDWithHistory(id string) (*models.Asset, error)
	FindBySerialNumber(serialNumber string) (*models.Asset, error)
	List(filter models.AssetFilter) ([]models.Asset, int64, error)
	Update(asset *models.Asset) error
	Delete(id string) error
	GetInventorySummary() (map[string]interface{}, error)
}

type assetRepository struct {
	db *gorm.DB
}

// NewAssetRepository creates a new AssetRepository instance
func NewAssetRepository(db *gorm.DB) AssetRepository {
	return &assetRepository{db: db}
}

func (r *assetRepository) Create(asset *models.Asset) error {
	return r.db.Create(asset).Error
}

func (r *assetRepository) FindByID(id string) (*models.Asset, error) {
	var asset models.Asset
	err := r.db.First(&asset, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *assetRepository) FindByIDWithHistory(id string) (*models.Asset, error) {
	var asset models.Asset
	err := r.db.
		Preload("Assignments", func(db *gorm.DB) *gorm.DB {
			return db.Order("assigned_at DESC")
		}).
		Preload("Assignments.User").
		Preload("Assignments.Assigner").
		First(&asset, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *assetRepository) FindBySerialNumber(serialNumber string) (*models.Asset, error) {
	var asset models.Asset
	err := r.db.Where("serial_number = ?", serialNumber).First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *assetRepository) List(filter models.AssetFilter) ([]models.Asset, int64, error) {
	var assets []models.Asset
	var total int64

	query := r.db.Model(&models.Asset{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Category != "" {
		query = query.Where("LOWER(category) = LOWER(?)", filter.Category)
	}

	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(serial_number) LIKE ? OR LOWER(specifications) LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
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

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&assets).Error
	return assets, total, err
}

func (r *assetRepository) Update(asset *models.Asset) error {
	return r.db.Save(asset).Error
}

func (r *assetRepository) Delete(id string) error {
	return r.db.Delete(&models.Asset{}, "id = ?", id).Error
}

func (r *assetRepository) GetInventorySummary() (map[string]interface{}, error) {
	var totalAssets int64
	var totalValuation float64

	if err := r.db.Model(&models.Asset{}).Count(&totalAssets).Error; err != nil {
		return nil, err
	}

	row := r.db.Model(&models.Asset{}).Select("COALESCE(SUM(purchase_cost), 0)").Row()
	if err := row.Scan(&totalValuation); err != nil {
		return nil, err
	}

	// Status breakdown
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusCounts []StatusCount
	if err := r.db.Model(&models.Asset{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}

	// Category breakdown
	type CategoryCount struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	var categoryCounts []CategoryCount
	if err := r.db.Model(&models.Asset{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&categoryCounts).Error; err != nil {
		return nil, err
	}

	statusMap := make(map[string]int64)
	for _, sc := range statusCounts {
		statusMap[sc.Status] = sc.Count
	}

	categoryMap := make(map[string]int64)
	for _, cc := range categoryCounts {
		categoryMap[cc.Category] = cc.Count
	}

	return map[string]interface{}{
		"total_assets":    totalAssets,
		"total_valuation": totalValuation,
		"by_status":       statusMap,
		"by_category":     categoryMap,
	}, nil
}

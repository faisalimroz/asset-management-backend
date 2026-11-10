package repository

import (
	"asset-management-backend/internal/models"

	"gorm.io/gorm"
)

// UserWithAssetsDTO custom DTO for HR view of employees and assigned assets count
type UserWithAssetsDTO struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	Role              string `json:"role"`
	Department        string `json:"department"`
	ActiveAssetsCount int64  `json:"active_assets_count"`
}

// UserRepository interface defines data access for users
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id string) (*models.User, error)
	List(page, limit int) ([]models.User, int64, error)
	ListEmployeesWithAssetCount(page, limit int) ([]UserWithAssetsDTO, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("LOWER(email) = LOWER(?)", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error
	return users, total, err
}

func (r *userRepository) ListEmployeesWithAssetCount(page, limit int) ([]UserWithAssetsDTO, int64, error) {
	var users []models.User
	var total int64

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := r.db.Model(&models.User{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("name ASC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	result := make([]UserWithAssetsDTO, len(users))
	for i, u := range users {
		var activeCount int64
		r.db.Model(&models.AssetAssignment{}).
			Where("user_id = ? AND returned_at IS NULL", u.ID).
			Count(&activeCount)

		result[i] = UserWithAssetsDTO{
			ID:                u.ID,
			Name:              u.Name,
			Email:             u.Email,
			Role:              u.Role,
			Department:        u.Department,
			ActiveAssetsCount: activeCount,
		}
	}

	return result, total, nil
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Asset status constants
const (
	AssetStatusAvailable   = "available"
	AssetStatusAssigned    = "assigned"
	AssetStatusMaintenance = "maintenance"
	AssetStatusRetired     = "retired"
)

// IsValidAssetStatus checks if the status is valid
func IsValidAssetStatus(status string) bool {
	switch status {
	case AssetStatusAvailable, AssetStatusAssigned, AssetStatusMaintenance, AssetStatusRetired:
		return true
	default:
		return false
	}
}

// Asset model represents a company physical or digital asset
type Asset struct {
	ID             string          `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name           string          `gorm:"type:varchar(150);not null" json:"name"`
	SerialNumber   string          `gorm:"type:varchar(100);uniqueIndex;not null" json:"serial_number"`
	Category       string          `gorm:"type:varchar(100);not null;index" json:"category"`
	Status         string          `gorm:"type:varchar(30);not null;default:'available';index" json:"status"`
	PurchaseDate   time.Time       `gorm:"not null" json:"purchase_date"`
	PurchaseCost   float64         `gorm:"type:decimal(12,2);default:0.00" json:"purchase_cost"`
	Specifications string          `gorm:"type:text" json:"specifications,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	Assignments []AssetAssignment `gorm:"foreignKey:AssetID" json:"assignments,omitempty"`
}

// BeforeCreate hook to generate UUID if not provided
func (a *Asset) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Status == "" {
		a.Status = AssetStatusAvailable
	}
	return nil
}

// CreateAssetRequest payload to add a new asset
type CreateAssetRequest struct {
	Name           string    `json:"name" binding:"required,min=2,max=150"`
	SerialNumber   string    `json:"serial_number" binding:"required,min=2,max=100"`
	Category       string    `json:"category" binding:"required,min=2,max=100"`
	Status         string    `json:"status" binding:"omitempty,oneof=available assigned maintenance retired"`
	PurchaseDate   time.Time `json:"purchase_date" binding:"required"`
	PurchaseCost   float64   `json:"purchase_cost" binding:"gte=0"`
	Specifications string    `json:"specifications"`
}

// UpdateAssetRequest payload to update an asset
type UpdateAssetRequest struct {
	Name           *string    `json:"name" binding:"omitempty,min=2,max=150"`
	SerialNumber   *string    `json:"serial_number" binding:"omitempty,min=2,max=100"`
	Category       *string    `json:"category" binding:"omitempty,min=2,max=100"`
	Status         *string    `json:"status" binding:"omitempty,oneof=available assigned maintenance retired"`
	PurchaseDate   *time.Time `json:"purchase_date"`
	PurchaseCost   *float64   `json:"purchase_cost" binding:"omitempty,gte=0"`
	Specifications *string    `json:"specifications"`
}

// AssetFilter represents query parameters for listing assets
type AssetFilter struct {
	Status   string `form:"status"`
	Category string `form:"category"`
	Search   string `form:"search"`
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=10"`
}

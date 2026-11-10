package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetAssignment represents an asset allocation event/history
type AssetAssignment struct {
	ID         string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	AssetID    string     `gorm:"type:varchar(36);not null;index" json:"asset_id"`
	UserID     string     `gorm:"type:varchar(36);not null;index" json:"user_id"`
	AssignedBy string     `gorm:"type:varchar(36);not null" json:"assigned_by"`
	AssignedAt time.Time  `gorm:"not null" json:"assigned_at"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`
	Notes      string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Relationships
	Asset    *Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
	User     *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Assigner *User  `gorm:"foreignKey:AssignedBy" json:"assigner,omitempty"`
}

// BeforeCreate hook to generate UUID if not provided
func (a *AssetAssignment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.AssignedAt.IsZero() {
		a.AssignedAt = time.Now()
	}
	return nil
}

// AssignAssetRequest payload for allocating an asset
type AssignAssetRequest struct {
	AssetID string `json:"asset_id" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
	Notes   string `json:"notes"`
}

// ReturnAssetRequest payload for returning an asset
type ReturnAssetRequest struct {
	ReturnedStatus string `json:"returned_status" binding:"omitempty,oneof=available maintenance retired"`
	Notes          string `json:"notes"`
}

// AssignmentFilter represents query parameters for assignments list
type AssignmentFilter struct {
	UserID     string `form:"user_id"`
	AssetID    string `form:"asset_id"`
	ActiveOnly bool   `form:"active_only"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=10"`
}

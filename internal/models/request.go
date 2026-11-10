package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Request status constants
const (
	RequestStatusPending   = "pending"
	RequestStatusApproved  = "approved"
	RequestStatusRejected  = "rejected"
	RequestStatusFulfilled = "fulfilled"
)

// IsValidRequestStatus checks if request status is valid
func IsValidRequestStatus(status string) bool {
	switch status {
	case RequestStatusPending, RequestStatusApproved, RequestStatusRejected, RequestStatusFulfilled:
		return true
	default:
		return false
	}
}

// AssetRequest represents an employee's request for hardware or peripherals
type AssetRequest struct {
	ID            string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID        string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	AssetCategory string    `gorm:"type:varchar(100);not null" json:"asset_category"`
	ItemRequested string    `gorm:"type:varchar(150);not null" json:"item_requested"`
	Reason        string    `gorm:"type:text;not null" json:"reason"`
	Status        string    `gorm:"type:varchar(30);not null;default:'pending';index" json:"status"`
	ReviewerID    *string   `gorm:"type:varchar(36)" json:"reviewer_id,omitempty"`
	ReviewNotes   string    `gorm:"type:text" json:"review_notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	User     *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Reviewer *User `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
}

// BeforeCreate hook to generate UUID if not provided
func (r *AssetRequest) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.Status == "" {
		r.Status = RequestStatusPending
	}
	return nil
}

// CreateAssetRequestDTO payload for employee submitting request
type CreateAssetRequestDTO struct {
	AssetCategory string `json:"asset_category" binding:"required,min=2,max=100"`
	ItemRequested string `json:"item_requested" binding:"required,min=2,max=150"`
	Reason        string `json:"reason" binding:"required,min=5"`
}

// UpdateRequestStatusDTO payload for IT/HR reviewing request
type UpdateRequestStatusDTO struct {
	Status      string `json:"status" binding:"required,oneof=approved rejected fulfilled"`
	ReviewNotes string `json:"review_notes" binding:"omitempty"`
}

// RequestFilter represents query parameters for requests list
type RequestFilter struct {
	Status string `form:"status"`
	UserID string `form:"user_id"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
}

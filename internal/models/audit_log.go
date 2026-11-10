package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Audit action constants
const (
	ActionUserRegistered      = "USER_REGISTERED"
	ActionAssetCreated        = "ASSET_CREATED"
	ActionAssetUpdated        = "ASSET_UPDATED"
	ActionAssetDeleted        = "ASSET_DELETED"
	ActionAssetAssigned       = "ASSET_ASSIGNED"
	ActionAssetReturned       = "ASSET_RETURNED"
	ActionRequestCreated      = "REQUEST_CREATED"
	ActionRequestStatusChange = "REQUEST_STATUS_CHANGED"
)

// AuditLog tracks critical business actions for compliance and IT oversight
type AuditLog struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID    string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	UserEmail string    `gorm:"type:varchar(150);not null" json:"user_email"`
	Action    string    `gorm:"type:varchar(50);not null;index" json:"action"`
	Entity    string    `gorm:"type:varchar(50);not null;index" json:"entity"`
	EntityID  string    `gorm:"type:varchar(36);not null" json:"entity_id"`
	Details   string    `gorm:"type:text" json:"details"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// BeforeCreate hook to generate UUID if not provided
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// AuditLogFilter query parameters for audit logs
type AuditLogFilter struct {
	Action string `form:"action"`
	Entity string `form:"entity"`
	UserID string `form:"user_id"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=20"`
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Leave type constants
const (
	LeaveTypeSick   = "sick"
	LeaveTypeCasual = "casual"
	LeaveTypeAnnual = "annual"
	LeaveTypeUnpaid = "unpaid"
)

// Leave status constants
const (
	LeaveStatusPending  = "pending"
	LeaveStatusApproved = "approved"
	LeaveStatusRejected = "rejected"
)

// LeaveRequest represents an employee's requested time off
type LeaveRequest struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID      string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	LeaveType   string    `gorm:"type:varchar(30);not null" json:"leave_type"`
	StartDate   time.Time `gorm:"not null" json:"start_date"`
	EndDate     time.Time `gorm:"not null" json:"end_date"`
	Reason      string    `gorm:"type:text;not null" json:"reason"`
	Status      string    `gorm:"type:varchar(30);not null;default:'pending';index" json:"status"`
	ApprovedBy  *string   `gorm:"type:varchar(36)" json:"approved_by,omitempty"`
	ReviewNotes string    `gorm:"type:text" json:"review_notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User     *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Approver *User `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

// BeforeCreate hook to generate UUID if not provided
func (l *LeaveRequest) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	if l.Status == "" {
		l.Status = LeaveStatusPending
	}
	return nil
}

// CreateLeaveRequestDTO payload for submitting a leave request
type CreateLeaveRequestDTO struct {
	LeaveType string    `json:"leave_type" binding:"required,oneof=sick casual annual unpaid"`
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" binding:"required"`
	Reason    string    `json:"reason" binding:"required,min=5"`
}

// UpdateLeaveStatusDTO payload for HR reviewing a leave request
type UpdateLeaveStatusDTO struct {
	Status      string `json:"status" binding:"required,oneof=approved rejected"`
	ReviewNotes string `json:"review_notes"`
}

// LeaveFilter query parameters for filtering leave requests
type LeaveFilter struct {
	UserID    string `form:"user_id"`
	Status    string `form:"status"`
	LeaveType string `form:"leave_type"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=10"`
}

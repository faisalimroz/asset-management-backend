package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Attendance status constants
const (
	AttendanceStatusPresent = "present"
	AttendanceStatusAbsent  = "absent"
	AttendanceStatusHalfDay = "half_day"
	AttendanceStatusRemote  = "remote"
)

// IsValidAttendanceStatus checks if attendance status is valid
func IsValidAttendanceStatus(status string) bool {
	switch status {
	case AttendanceStatusPresent, AttendanceStatusAbsent, AttendanceStatusHalfDay, AttendanceStatusRemote:
		return true
	default:
		return false
	}
}

// AttendanceLog represents an employee's daily check-in/check-out and working hours
type AttendanceLog struct {
	ID           string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID       string     `gorm:"type:varchar(36);not null;index:idx_user_date,unique" json:"user_id"`
	Date         string     `gorm:"type:varchar(10);not null;index:idx_user_date,unique;index" json:"date"` // Format: YYYY-MM-DD
	CheckInTime  *time.Time `json:"check_in_time,omitempty"`
	CheckOutTime *time.Time `json:"check_out_time,omitempty"`
	WorkHours    float64    `gorm:"type:decimal(5,2);default:0.00" json:"work_hours"`
	Status       string     `gorm:"type:varchar(30);not null;default:'present';index" json:"status"`
	Notes        string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// BeforeCreate hook to generate UUID and format current date if missing
func (a *AttendanceLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Date == "" {
		a.Date = time.Now().Format("2006-01-02")
	}
	if a.Status == "" {
		a.Status = AttendanceStatusPresent
	}
	return nil
}

// CheckInRequest payload for checking in
type CheckInRequest struct {
	Status string `json:"status" binding:"omitempty,oneof=present remote half_day"`
	Notes  string `json:"notes"`
}

// CheckOutRequest payload for checking out
type CheckOutRequest struct {
	Notes string `json:"notes"`
}

// AttendanceFilter query parameters for filtering attendance records
type AttendanceFilter struct {
	UserID     string `form:"user_id"`
	Department string `form:"department"`
	StartDate  string `form:"start_date"` // YYYY-MM-DD
	EndDate    string `form:"end_date"`   // YYYY-MM-DD
	Status     string `form:"status"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
}

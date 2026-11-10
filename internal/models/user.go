package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User role constants
const (
	RoleEmployee = "employee"
	RoleHR       = "hr"
	RoleIT       = "it"
)

// IsValidRole checks if the role string is one of the supported roles
func IsValidRole(role string) bool {
	switch role {
	case RoleEmployee, RoleHR, RoleIT:
		return true
	default:
		return false
	}
}

// User model represents a registered system user
type User struct {
	ID           string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	Email        string    `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Role         string    `gorm:"type:varchar(20);not null;default:'employee'" json:"role"`
	Department   string    `gorm:"type:varchar(100);not null" json:"department"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not provided
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// UserResponse sanitizes user data for public/API return
type UserResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:         u.ID,
		Name:       u.Name,
		Email:      u.Email,
		Role:       u.Role,
		Department: u.Department,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

// RegisterRequest payload for registering a user
type RegisterRequest struct {
	Name       string `json:"name" binding:"required,min=2,max=100"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Role       string `json:"role" binding:"required,oneof=employee hr it"`
	Department string `json:"department" binding:"required,min=2,max=100"`
}

// LoginRequest payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse response after successful login
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

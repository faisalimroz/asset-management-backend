package services

import (
	"errors"
	"fmt"

	"asset-management-backend/internal/config"
	"asset-management-backend/internal/models"
	"asset-management-backend/internal/repository"
	"asset-management-backend/pkg/utils"

	"gorm.io/gorm"
)

// AuthService interface defines authentication operations
type AuthService interface {
	Register(req *models.RegisterRequest, ip string) (*models.UserResponse, error)
	Login(req *models.LoginRequest) (*models.LoginResponse, error)
	GetProfile(userID string) (*models.UserResponse, error)
}

type authService struct {
	userRepo  repository.UserRepository
	auditRepo repository.AuditRepository
	cfg       *config.Config
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo repository.UserRepository, auditRepo repository.AuditRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		cfg:       cfg,
	}
}

func (s *authService) Register(req *models.RegisterRequest, ip string) (*models.UserResponse, error) {
	// Check if user with email already exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, utils.NewConflictError("user with this email already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.NewInternalServerError("failed to check existing user", err)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to hash password", err)
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		Department:   req.Department,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, utils.NewInternalServerError("failed to create user", err)
	}

	// Audit log
	s.auditRepo.Create(&models.AuditLog{
		UserID:    user.ID,
		UserEmail: user.Email,
		Action:    models.ActionUserRegistered,
		Entity:    "user",
		EntityID:  user.ID,
		Details:   fmt.Sprintf("User registered with role %s in department %s", user.Role, user.Department),
		IPAddress: ip,
	})

	res := user.ToResponse()
	return &res, nil
}

func (s *authService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewUnauthorizedError("invalid email or password")
		}
		return nil, utils.NewInternalServerError("failed to find user", err)
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, utils.NewUnauthorizedError("invalid email or password")
	}

	token, err := utils.GenerateJWT(
		user.ID,
		user.Email,
		user.Role,
		user.Department,
		s.cfg.JWTSecret,
		s.cfg.JWTExpirationHr,
	)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to generate access token", err)
	}

	return &models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

func (s *authService) GetProfile(userID string) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError("user not found")
		}
		return nil, utils.NewInternalServerError("failed to fetch user profile", err)
	}

	res := user.ToResponse()
	return &res, nil
}

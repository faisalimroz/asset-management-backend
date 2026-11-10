package services

import (
	"asset-management-backend/internal/models"
	"asset-management-backend/internal/repository"
)

// ReportService interface defines reporting and audit operations
type ReportService interface {
	GetInventorySummary() (map[string]interface{}, error)
	GetAuditLogs(filter models.AuditLogFilter) ([]models.AuditLog, int64, error)
	GetEmployeeAllocations(page, limit int) ([]repository.UserWithAssetsDTO, int64, error)
	GetEmployeeAssetHistory(userID string) ([]models.AssetAssignment, error)
}

type reportService struct {
	assetRepo      repository.AssetRepository
	assignmentRepo repository.AssignmentRepository
	userRepo       repository.UserRepository
	auditRepo      repository.AuditRepository
}

// NewReportService creates a new ReportService instance
func NewReportService(
	assetRepo repository.AssetRepository,
	assignmentRepo repository.AssignmentRepository,
	userRepo repository.UserRepository,
	auditRepo repository.AuditRepository,
) ReportService {
	return &reportService{
		assetRepo:      assetRepo,
		assignmentRepo: assignmentRepo,
		userRepo:       userRepo,
		auditRepo:      auditRepo,
	}
}

func (s *reportService) GetInventorySummary() (map[string]interface{}, error) {
	return s.assetRepo.GetInventorySummary()
}

func (s *reportService) GetAuditLogs(filter models.AuditLogFilter) ([]models.AuditLog, int64, error) {
	return s.auditRepo.List(filter)
}

func (s *reportService) GetEmployeeAllocations(page, limit int) ([]repository.UserWithAssetsDTO, int64, error) {
	return s.userRepo.ListEmployeesWithAssetCount(page, limit)
}

func (s *reportService) GetEmployeeAssetHistory(userID string) ([]models.AssetAssignment, error) {
	return s.assignmentRepo.ListByUserID(userID)
}

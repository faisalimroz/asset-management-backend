package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asset-management-backend/internal/config"
	"asset-management-backend/internal/handlers"
	"asset-management-backend/internal/models"
	"asset-management-backend/internal/repository"
	"asset-management-backend/internal/router"
	"asset-management-backend/internal/services"
	"asset-management-backend/pkg/database"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TestSuite struct {
	DB     *gorm.DB
	Router *gin.Engine
	Config *config.Config

	ITAdminToken  string
	HRToken       string
	EmployeeToken string

	ITAdminUser  models.User
	HRUser       models.User
	EmployeeUser models.User
}

// SetupTestApp creates an in-memory test database, runs migrations, seeds users, and configures the router
func SetupTestApp(t *testing.T) *TestSuite {
	gin.SetMode(gin.TestMode)

	// In-memory unique SQLite database for test isolation
	dbName := "file:memdb_" + uuid.New().String() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to initialize test SQLite DB: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to auto-migrate test DB: %v", err)
	}

	cfg := &config.Config{
		Environment:     "test",
		JWTSecret:       "test-secret-key-1234567890!",
		JWTExpirationHr: 1,
		AutoSeed:        false,
	}

	// Create test users
	hash, _ := utils.HashPassword("Password@123")

	itAdmin := models.User{
		Name:         "IT Admin Test",
		Email:        "it.test@" + uuid.New().String() + ".com",
		PasswordHash: hash,
		Role:         models.RoleIT,
		Department:   "IT",
	}
	db.Create(&itAdmin)

	hrUser := models.User{
		Name:         "HR Manager Test",
		Email:        "hr.test@" + uuid.New().String() + ".com",
		PasswordHash: hash,
		Role:         models.RoleHR,
		Department:   "Human Resources",
	}
	db.Create(&hrUser)

	empUser := models.User{
		Name:         "Employee Test",
		Email:        "emp.test@" + uuid.New().String() + ".com",
		PasswordHash: hash,
		Role:         models.RoleEmployee,
		Department:   "Engineering",
	}
	db.Create(&empUser)

	itToken, _ := utils.GenerateJWT(itAdmin.ID, itAdmin.Email, itAdmin.Role, itAdmin.Department, cfg.JWTSecret, cfg.JWTExpirationHr)
	hrToken, _ := utils.GenerateJWT(hrUser.ID, hrUser.Email, hrUser.Role, hrUser.Department, cfg.JWTSecret, cfg.JWTExpirationHr)
	empToken, _ := utils.GenerateJWT(empUser.ID, empUser.Email, empUser.Role, empUser.Department, cfg.JWTSecret, cfg.JWTExpirationHr)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	assignmentRepo := repository.NewAssignmentRepository(db)
	requestRepo := repository.NewRequestRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	leaveRepo := repository.NewLeaveRepository(db)

	// Services
	authService := services.NewAuthService(userRepo, auditRepo, cfg)
	assetService := services.NewAssetService(assetRepo, auditRepo)
	assignmentService := services.NewAssignmentService(assignmentRepo, assetRepo, userRepo, auditRepo)
	requestService := services.NewRequestService(requestRepo, auditRepo)
	reportService := services.NewReportService(assetRepo, assignmentRepo, userRepo, auditRepo)
	attendanceService := services.NewAttendanceService(attendanceRepo, auditRepo)
	leaveService := services.NewLeaveService(leaveRepo, auditRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	assetHandler := handlers.NewAssetHandler(assetService)
	assignmentHandler := handlers.NewAssignmentHandler(assignmentService)
	requestHandler := handlers.NewRequestHandler(requestService)
	reportHandler := handlers.NewReportHandler(reportService)
	userHandler := handlers.NewUserHandler(reportService)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceService)
	leaveHandler := handlers.NewLeaveHandler(leaveService)

	r := router.SetupRouter(
		cfg,
		authHandler,
		assetHandler,
		assignmentHandler,
		requestHandler,
		reportHandler,
		userHandler,
		attendanceHandler,
		leaveHandler,
	)

	return &TestSuite{
		DB:            db,
		Router:        r,
		Config:        cfg,
		ITAdminToken:  itToken,
		HRToken:       hrToken,
		EmployeeToken: empToken,
		ITAdminUser:   itAdmin,
		HRUser:        hrUser,
		EmployeeUser:  empUser,
	}
}

// PerformRequest helper function to simulate HTTP requests against the Gin router
func PerformRequest(r *gin.Engine, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// CreateTestAsset helper to insert an asset directly into the DB for tests
func (s *TestSuite) CreateTestAsset(name, serial, category, status string, cost float64) models.Asset {
	asset := models.Asset{
		Name:           name,
		SerialNumber:   serial,
		Category:       category,
		Status:         status,
		PurchaseDate:   time.Now().AddDate(-1, 0, 0),
		PurchaseCost:   cost,
		Specifications: "Standard test specs",
	}
	s.DB.Create(&asset)
	return asset
}

package router

import (
	"net/http"

	"asset-management-backend/internal/config"
	"asset-management-backend/internal/handlers"
	"asset-management-backend/internal/middleware"
	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes Gin engine with middlewares and routes
func SetupRouter(
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	assetHandler *handlers.AssetHandler,
	assignmentHandler *handlers.AssignmentHandler,
	requestHandler *handlers.RequestHandler,
	reportHandler *handlers.ReportHandler,
	userHandler *handlers.UserHandler,
	attendanceHandler *handlers.AttendanceHandler,
	leaveHandler *handlers.LeaveHandler,
) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middlewares
	r.Use(gin.Logger())
	r.Use(middleware.ErrorMiddleware())
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		utils.SendSuccess(c, http.StatusOK, "asset management API is operational", gin.H{
			"status": "healthy",
			"env":    cfg.Environment,
		})
	})

	api := r.Group("/api/v1")
	{
		// Health under API prefix too
		api.GET("/health", func(c *gin.Context) {
			utils.SendSuccess(c, http.StatusOK, "healthy", nil)
		})

		// ---------------------------------------------------------
		// Public Auth Endpoints
		// ---------------------------------------------------------
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		// ---------------------------------------------------------
		// Protected Endpoints (Requires valid JWT)
		// ---------------------------------------------------------
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// Auth Profile
			protected.GET("/auth/me", authHandler.GetProfile)

			// -----------------------------------------------------
			// Assets: IT and HR can view; IT can mutate
			// -----------------------------------------------------
			assetsGroup := protected.Group("/assets")
			{
				// IT and HR view endpoints
				assetsGroup.GET("", middleware.RequireRoles(models.RoleIT, models.RoleHR), assetHandler.ListAssets)
				assetsGroup.GET("/:id", middleware.RequireRoles(models.RoleIT, models.RoleHR), assetHandler.GetAsset)

				// IT only mutations
				assetsGroup.POST("", middleware.RequireRoles(models.RoleIT), assetHandler.CreateAsset)
				assetsGroup.PUT("/:id", middleware.RequireRoles(models.RoleIT), assetHandler.UpdateAsset)
				assetsGroup.DELETE("/:id", middleware.RequireRoles(models.RoleIT), assetHandler.DeleteAsset)
			}

			// -----------------------------------------------------
			// Assignments: IT assigns, IT/HR returns, Employees view own
			// -----------------------------------------------------
			assignmentsGroup := protected.Group("/assignments")
			{
				// Employee (and all authenticated): View own assigned assets
				assignmentsGroup.GET("/my-assets", assignmentHandler.GetMyAssets)

				// IT & HR: View all assignments
				assignmentsGroup.GET("", middleware.RequireRoles(models.RoleIT, models.RoleHR), assignmentHandler.ListAssignments)

				// IT only: Assign asset to employee
				assignmentsGroup.POST("", middleware.RequireRoles(models.RoleIT), assignmentHandler.AssignAsset)

				// IT and HR: Mark asset as returned (offboarding / upgrades)
				assignmentsGroup.PUT("/:id/return", middleware.RequireRoles(models.RoleIT, models.RoleHR), assignmentHandler.ReturnAsset)
			}

			// -----------------------------------------------------
			// Asset Requests: Employee submits, IT/HR reviews
			// -----------------------------------------------------
			requestsGroup := protected.Group("/requests")
			{
				// Any authenticated user can submit a request (primarily employees)
				requestsGroup.POST("", requestHandler.CreateRequest)

				// Employees view own requests; IT and HR view all requests
				requestsGroup.GET("", requestHandler.ListRequests)
				requestsGroup.GET("/:id", requestHandler.GetRequest)

				// IT & HR can approve / reject / fulfill requests
				requestsGroup.PUT("/:id/status", middleware.RequireRoles(models.RoleIT, models.RoleHR), requestHandler.UpdateRequestStatus)
			}

			// -----------------------------------------------------
			// Users (HR / IT employee inventory tracking)
			// -----------------------------------------------------
			usersGroup := protected.Group("/users")
			usersGroup.Use(middleware.RequireRoles(models.RoleIT, models.RoleHR))
			{
				usersGroup.GET("", userHandler.ListEmployees)
				usersGroup.GET("/:id/assets", userHandler.GetEmployeeAssets)
			}

			// -----------------------------------------------------
			// Attendance Module: Check-in, Check-out, My logs, HR Report
			// -----------------------------------------------------
			attendanceGroup := protected.Group("/attendance")
			{
				attendanceGroup.POST("/check-in", attendanceHandler.CheckIn)
				attendanceGroup.POST("/check-out", attendanceHandler.CheckOut)
				attendanceGroup.GET("/my-logs", attendanceHandler.GetMyLogs)
				attendanceGroup.GET("/report", middleware.RequireRoles(models.RoleHR), attendanceHandler.GetAttendanceReport)
			}

			// -----------------------------------------------------
			// Leave Management Module: Submit, List, Review
			// -----------------------------------------------------
			leavesGroup := protected.Group("/leaves")
			{
				leavesGroup.POST("", leaveHandler.CreateLeave)
				leavesGroup.GET("", leaveHandler.ListLeaves)
				leavesGroup.GET("/:id", leaveHandler.GetLeave)
				leavesGroup.PUT("/:id/status", middleware.RequireRoles(models.RoleHR), leaveHandler.UpdateLeaveStatus)
			}

			// -----------------------------------------------------
			// Reports & Audit Logs
			// -----------------------------------------------------
			reportsGroup := protected.Group("/reports")
			{
				// IT & HR: Inventory breakdown & financial valuation summary
				reportsGroup.GET("/inventory", middleware.RequireRoles(models.RoleIT, models.RoleHR), reportHandler.GetInventorySummary)

				// IT only: Complete audit log trail
				reportsGroup.GET("/audit-logs", middleware.RequireRoles(models.RoleIT), reportHandler.GetAuditLogs)
			}
		}
	}

	return r
}

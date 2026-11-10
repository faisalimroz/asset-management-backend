package database

import (
	"log"
	"time"

	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"gorm.io/gorm"
)

// SeedData inserts initial users and assets if database is empty
func SeedData(db *gorm.DB) error {
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return err
	}

	if userCount > 0 {
		log.Println("Database already contains users. Skipping initial seeding.")
		return nil
	}

	log.Println("Seeding initial users and demo assets...")

	// 1. Create Users
	adminPassword, _ := utils.HashPassword("Admin@123")
	hrPassword, _ := utils.HashPassword("HR@123")
	empPassword, _ := utils.HashPassword("Employee@123")

	itAdmin := models.User{
		Name:         "Alex Turner (IT Admin)",
		Email:        "it.admin@company.com",
		PasswordHash: adminPassword,
		Role:         models.RoleIT,
		Department:   "Information Technology",
	}

	hrManager := models.User{
		Name:         "Sarah Connor (HR Manager)",
		Email:        "hr.manager@company.com",
		PasswordHash: hrPassword,
		Role:         models.RoleHR,
		Department:   "Human Resources",
	}

	emp1 := models.User{
		Name:         "John Doe (Backend Engineer)",
		Email:        "john.doe@company.com",
		PasswordHash: empPassword,
		Role:         models.RoleEmployee,
		Department:   "Engineering",
	}

	emp2 := models.User{
		Name:         "Jane Smith (UI/UX Designer)",
		Email:        "jane.smith@company.com",
		PasswordHash: empPassword,
		Role:         models.RoleEmployee,
		Department:   "Product Design",
	}

	if err := db.Create(&itAdmin).Error; err != nil {
		return err
	}
	if err := db.Create(&hrManager).Error; err != nil {
		return err
	}
	if err := db.Create(&emp1).Error; err != nil {
		return err
	}
	if err := db.Create(&emp2).Error; err != nil {
		return err
	}

	// 2. Create Assets
	now := time.Now()
	sixMonthsAgo := now.AddDate(0, -6, 0)
	oneYearAgo := now.AddDate(-1, 0, 0)

	assets := []models.Asset{
		{
			Name:           "MacBook Pro 16 M3 Max",
			SerialNumber:   "MBP16-2024-001",
			Category:       "Laptop",
			Status:         models.AssetStatusAvailable,
			PurchaseDate:   sixMonthsAgo,
			PurchaseCost:   3499.00,
			Specifications: "Apple M3 Max, 36GB Unified RAM, 1TB NVMe SSD, Space Black",
		},
		{
			Name:           "Dell XPS 15 9530",
			SerialNumber:   "DELL-XPS-2024-002",
			Category:       "Laptop",
			Status:         models.AssetStatusAssigned,
			PurchaseDate:   oneYearAgo,
			PurchaseCost:   2199.00,
			Specifications: "Intel Core i7-13700H, 32GB RAM, 1TB SSD, RTX 4060 OLED",
		},
		{
			Name:           "Dell UltraSharp 27 4K Monitor (U2723QE)",
			SerialNumber:   "MON-U2723QE-003",
			Category:       "Monitor",
			Status:         models.AssetStatusAssigned,
			PurchaseDate:   oneYearAgo,
			PurchaseCost:   579.00,
			Specifications: "27-inch 4K UHD, IPS Black panel, 98% DCI-P3, USB-C 90W PD",
		},
		{
			Name:           "Apple Studio Display 27 5K",
			SerialNumber:   "APL-STD-2024-004",
			Category:       "Monitor",
			Status:         models.AssetStatusAvailable,
			PurchaseDate:   sixMonthsAgo,
			PurchaseCost:   1599.00,
			Specifications: "27-inch 5K Retina display, 12MP Ultra Wide camera, 6 speakers",
		},
		{
			Name:           "Keychron Q1 Pro Wireless Custom Keyboard",
			SerialNumber:   "KB-Q1PRO-005",
			Category:       "Accessory",
			Status:         models.AssetStatusAvailable,
			PurchaseDate:   sixMonthsAgo,
			PurchaseCost:   199.00,
			Specifications: "75% layout, CNC aluminum body, Banana switches, Bluetooth 5.1",
		},
		{
			Name:           "Logitech MX Master 3S Wireless Mouse",
			SerialNumber:   "MSE-MX3S-006",
			Category:       "Accessory",
			Status:         models.AssetStatusMaintenance,
			PurchaseDate:   oneYearAgo,
			PurchaseCost:   99.00,
			Specifications: "8K DPI optical sensor, quiet clicks, MagSpeed electromagnetic scroll",
		},
		{
			Name:           "YubiKey 5C NFC Security Key",
			SerialNumber:   "YUBI-5CNFC-007",
			Category:       "Security",
			Status:         models.AssetStatusAvailable,
			PurchaseDate:   sixMonthsAgo,
			PurchaseCost:   55.00,
			Specifications: "FIDO2, U2F, Smart Card, OTP, NFC and USB-C",
		},
	}

	for i := range assets {
		if err := db.Create(&assets[i]).Error; err != nil {
			return err
		}
	}

	// 3. Create Demo Asset Assignments (Assign Dell XPS & Monitor to John Doe)
	assignment1 := models.AssetAssignment{
		AssetID:    assets[1].ID, // Dell XPS 15
		UserID:     emp1.ID,
		AssignedBy: itAdmin.ID,
		AssignedAt: oneYearAgo,
		Notes:      "Standard engineering laptop kit issued during onboarding",
	}

	assignment2 := models.AssetAssignment{
		AssetID:    assets[2].ID, // Dell UltraSharp
		UserID:     emp1.ID,
		AssignedBy: itAdmin.ID,
		AssignedAt: oneYearAgo,
		Notes:      "Office workstation dual-monitor setup",
	}

	if err := db.Create(&assignment1).Error; err != nil {
		return err
	}
	if err := db.Create(&assignment2).Error; err != nil {
		return err
	}

	// 4. Create Sample Requests
	req1 := models.AssetRequest{
		UserID:        emp2.ID, // Jane Smith
		AssetCategory: "Monitor",
		ItemRequested: "Apple Studio Display 27 5K",
		Reason:        "Need color-accurate 5K display for Figma design system work and brand asset creation.",
		Status:        models.RequestStatusPending,
	}

	approvedStatus := models.RequestStatusApproved
	itAdminID := itAdmin.ID
	req2 := models.AssetRequest{
		UserID:        emp1.ID, // John Doe
		AssetCategory: "Accessory",
		ItemRequested: "Logitech MX Master 3S Mouse",
		Reason:        "Replacement mouse for wrist ergonomics.",
		Status:        approvedStatus,
		ReviewerID:    &itAdminID,
		ReviewNotes:   "Approved. Awaiting stock delivery from vendor.",
	}

	if err := db.Create(&req1).Error; err != nil {
		return err
	}
	if err := db.Create(&req2).Error; err != nil {
		return err
	}

	// 5. Create Demo Attendance Logs
	yesterday := now.AddDate(0, 0, -1)
	yesterdayStr := yesterday.Format("2006-01-02")
	todayStr := now.Format("2006-01-02")

	yCheckIn1 := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 9, 0, 0, 0, time.Local)
	yCheckOut1 := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 17, 30, 0, 0, time.Local)
	tCheckIn1 := time.Date(now.Year(), now.Month(), now.Day(), 9, 15, 0, 0, time.Local)

	yCheckIn2 := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 9, 30, 0, 0, time.Local)
	yCheckOut2 := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 18, 0, 0, 0, time.Local)

	attendanceLogs := []models.AttendanceLog{
		{
			UserID:       emp1.ID,
			Date:         yesterdayStr,
			CheckInTime:  &yCheckIn1,
			CheckOutTime: &yCheckOut1,
			WorkHours:    8.5,
			Status:       models.AttendanceStatusPresent,
			Notes:        "On-site engineering shift",
		},
		{
			UserID:      emp1.ID,
			Date:        todayStr,
			CheckInTime: &tCheckIn1,
			WorkHours:   0,
			Status:      models.AttendanceStatusRemote,
			Notes:       "Working remotely today",
		},
		{
			UserID:       emp2.ID,
			Date:         yesterdayStr,
			CheckInTime:  &yCheckIn2,
			CheckOutTime: &yCheckOut2,
			WorkHours:    8.5,
			Status:       models.AttendanceStatusPresent,
		},
	}

	for i := range attendanceLogs {
		if err := db.Create(&attendanceLogs[i]).Error; err != nil {
			return err
		}
	}

	// 6. Create Demo Leave Requests
	hrManagerID := hrManager.ID
	leave1 := models.LeaveRequest{
		UserID:      emp1.ID, // John Doe
		LeaveType:   models.LeaveTypeAnnual,
		StartDate:   now.AddDate(0, 0, 10),
		EndDate:     now.AddDate(0, 0, 14),
		Reason:      "Family vacation",
		Status:      models.LeaveStatusApproved,
		ApprovedBy:  &hrManagerID,
		ReviewNotes: "Approved by HR. Have a great vacation!",
	}

	leave2 := models.LeaveRequest{
		UserID:    emp2.ID, // Jane Smith
		LeaveType: models.LeaveTypeSick,
		StartDate: now.AddDate(0, 0, 1),
		EndDate:   now.AddDate(0, 0, 2),
		Reason:    "Doctor appointment and recovery",
		Status:    models.LeaveStatusPending,
	}

	if err := db.Create(&leave1).Error; err != nil {
		return err
	}
	if err := db.Create(&leave2).Error; err != nil {
		return err
	}

	// 7. Initial Audit Log
	audit := models.AuditLog{
		UserID:    itAdmin.ID,
		UserEmail: itAdmin.Email,
		Action:    models.ActionAssetCreated,
		Entity:    "system",
		EntityID:  "initial-seed",
		Details:   "Initial system demonstration data populated successfully.",
		IPAddress: "127.0.0.1",
	}
	db.Create(&audit)

	log.Println("Seed data initialized successfully.")
	return nil
}

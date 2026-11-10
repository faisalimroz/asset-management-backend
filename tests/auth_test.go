package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuth_Register(t *testing.T) {
	suite := SetupTestApp(t)

	t.Run("successful registration", func(t *testing.T) {
		req := models.RegisterRequest{
			Name:       "New Engineer",
			Email:      "engineer_" + uuid.New().String() + "@company.com",
			Password:   "SecurePass@123",
			Role:       models.RoleEmployee,
			Department: "Engineering",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/auth/register", "", req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "user registered successfully", resp.Message)
	})

	t.Run("duplicate email registration fails", func(t *testing.T) {
		email := "duplicate_" + uuid.New().String() + "@company.com"
		req := models.RegisterRequest{
			Name:       "First User",
			Email:      email,
			Password:   "SecurePass@123",
			Role:       models.RoleEmployee,
			Department: "Marketing",
		}

		w1 := PerformRequest(suite.Router, "POST", "/api/v1/auth/register", "", req)
		assert.Equal(t, http.StatusCreated, w1.Code)

		// Attempt duplicate registration
		w2 := PerformRequest(suite.Router, "POST", "/api/v1/auth/register", "", req)
		assert.Equal(t, http.StatusConflict, w2.Code)
	})

	t.Run("invalid role fails validation", func(t *testing.T) {
		req := models.RegisterRequest{
			Name:       "Invalid Role User",
			Email:      "invalid_role_" + uuid.New().String() + "@company.com",
			Password:   "SecurePass@123",
			Role:       "superadmin", // not in employee/hr/it
			Department: "Executive",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/auth/register", "", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuth_Login(t *testing.T) {
	suite := SetupTestApp(t)

	t.Run("successful login with correct credentials", func(t *testing.T) {
		email := "login_test_" + uuid.New().String() + "@company.com"
		regReq := models.RegisterRequest{
			Name:       "Login Tester",
			Email:      email,
			Password:   "CorrectPassword123!",
			Role:       models.RoleHR,
			Department: "HR",
		}
		PerformRequest(suite.Router, "POST", "/api/v1/auth/register", "", regReq)

		loginReq := models.LoginRequest{
			Email:    email,
			Password: "CorrectPassword123!",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/auth/login", "", loginReq)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		assert.NotEmpty(t, dataMap["token"])
	})

	t.Run("login fails with wrong password", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Email:    suite.ITAdminUser.Email,
			Password: "WrongPassword!",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/auth/login", "", loginReq)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("login fails with non-existent email", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Email:    "nonexistent@company.com",
			Password: "Password123!",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/auth/login", "", loginReq)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuth_GetProfile(t *testing.T) {
	suite := SetupTestApp(t)

	t.Run("get profile with valid token", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/auth/me", suite.EmployeeToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		assert.Equal(t, suite.EmployeeUser.Email, dataMap["email"])
		assert.Equal(t, models.RoleEmployee, dataMap["role"])
	})

	t.Run("get profile fails without token", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/auth/me", "", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

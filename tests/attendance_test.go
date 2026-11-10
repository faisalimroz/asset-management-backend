package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestAttendance_Lifecycle(t *testing.T) {
	suite := SetupTestApp(t)

	t.Run("Employee can check in successfully", func(t *testing.T) {
		checkInReq := models.CheckInRequest{
			Status: models.AttendanceStatusPresent,
			Notes:  "Morning check-in at office desk",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/attendance/check-in", suite.EmployeeToken, checkInReq)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		assert.NotEmpty(t, dataMap["check_in_time"])
		assert.Equal(t, models.AttendanceStatusPresent, dataMap["status"])
	})

	t.Run("Duplicate check-in on the same day fails with 409 Conflict", func(t *testing.T) {
		checkInReq := models.CheckInRequest{
			Status: models.AttendanceStatusPresent,
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/attendance/check-in", suite.EmployeeToken, checkInReq)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Employee can check out successfully", func(t *testing.T) {
		checkOutReq := models.CheckOutRequest{
			Notes: "End of shift",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/attendance/check-out", suite.EmployeeToken, checkOutReq)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		assert.NotEmpty(t, dataMap["check_out_time"])
		assert.Contains(t, dataMap["notes"], "End of shift")
	})

	t.Run("Duplicate check-out on the same day fails with 409 Conflict", func(t *testing.T) {
		checkOutReq := models.CheckOutRequest{
			Notes: "Second checkout attempt",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/attendance/check-out", suite.EmployeeToken, checkOutReq)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Employee can view personal attendance logs", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/attendance/my-logs?page=1&limit=10", suite.EmployeeToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.GreaterOrEqual(t, resp.Meta.Total, int64(1))
	})

	t.Run("HR can view company-wide attendance report", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/attendance/report", suite.HRToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.GreaterOrEqual(t, resp.Meta.Total, int64(1))
	})

	t.Run("Employee cannot access HR attendance report (403 Forbidden)", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/attendance/report", suite.EmployeeToken, nil)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("IT admin cannot access HR attendance report (403 Forbidden)", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/attendance/report", suite.ITAdminToken, nil)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestLeave_Lifecycle(t *testing.T) {
	suite := SetupTestApp(t)

	var leaveID string
	now := time.Now()

	t.Run("Employee can submit leave request", func(t *testing.T) {
		req := models.CreateLeaveRequestDTO{
			LeaveType: models.LeaveTypeCasual,
			StartDate: now.AddDate(0, 0, 5),
			EndDate:   now.AddDate(0, 0, 7),
			Reason:    "Personal family commitment",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/leaves", suite.EmployeeToken, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		leaveID = dataMap["id"].(string)
		assert.NotEmpty(t, leaveID)
		assert.Equal(t, models.LeaveStatusPending, dataMap["status"])
		assert.Equal(t, models.LeaveTypeCasual, dataMap["leave_type"])
	})

	t.Run("Invalid dates (end date before start date) fails with 400 Bad Request", func(t *testing.T) {
		req := models.CreateLeaveRequestDTO{
			LeaveType: models.LeaveTypeSick,
			StartDate: now.AddDate(0, 0, 10),
			EndDate:   now.AddDate(0, 0, 5), // end is before start
			Reason:    "Invalid dates test",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/leaves", suite.EmployeeToken, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Employee can view their own leave requests", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/leaves", suite.EmployeeToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.GreaterOrEqual(t, resp.Meta.Total, int64(1))
	})

	t.Run("HR can view all company leave requests", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/leaves", suite.HRToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.GreaterOrEqual(t, resp.Meta.Total, int64(1))
	})

	t.Run("Employee cannot approve leave (403 Forbidden)", func(t *testing.T) {
		updateReq := models.UpdateLeaveStatusDTO{
			Status:      models.LeaveStatusApproved,
			ReviewNotes: "Self approval attempt",
		}

		w := PerformRequest(suite.Router, "PUT", "/api/v1/leaves/"+leaveID+"/status", suite.EmployeeToken, updateReq)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("IT admin cannot approve leave (403 Forbidden - HR only)", func(t *testing.T) {
		updateReq := models.UpdateLeaveStatusDTO{
			Status:      models.LeaveStatusApproved,
			ReviewNotes: "IT admin approval attempt",
		}

		w := PerformRequest(suite.Router, "PUT", "/api/v1/leaves/"+leaveID+"/status", suite.ITAdminToken, updateReq)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("HR manager can approve leave request", func(t *testing.T) {
		updateReq := models.UpdateLeaveStatusDTO{
			Status:      models.LeaveStatusApproved,
			ReviewNotes: "Approved by HR. Coverage confirmed.",
		}

		w := PerformRequest(suite.Router, "PUT", "/api/v1/leaves/"+leaveID+"/status", suite.HRToken, updateReq)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		assert.Equal(t, models.LeaveStatusApproved, dataMap["status"])
		assert.Equal(t, suite.HRUser.ID, dataMap["approved_by"])
		assert.Equal(t, "Approved by HR. Coverage confirmed.", dataMap["review_notes"])
	})
}

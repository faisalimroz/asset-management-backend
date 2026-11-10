package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestRequest_Lifecycle(t *testing.T) {
	suite := SetupTestApp(t)

	var requestID string

	t.Run("Employee submits asset request", func(t *testing.T) {
		req := models.CreateAssetRequestDTO{
			AssetCategory: "Peripherals",
			ItemRequested: "Ergonomic Split Keyboard",
			Reason:        "Prevent repetitive strain injury during coding",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/requests", suite.EmployeeToken, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		requestID = dataMap["id"].(string)
		assert.NotEmpty(t, requestID)
		assert.Equal(t, models.RequestStatusPending, dataMap["status"])
	})

	t.Run("Employee views their requests", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/requests", suite.EmployeeToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		items := resp.Data.([]interface{})
		assert.GreaterOrEqual(t, len(items), 1)
	})

	t.Run("IT admin views all requests", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/requests", suite.ITAdminToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.GreaterOrEqual(t, resp.Meta.Total, int64(1))
	})

	t.Run("HR views all requests", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/requests", suite.HRToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Employee cannot approve request (403 Forbidden)", func(t *testing.T) {
		statusReq := models.UpdateRequestStatusDTO{
			Status:      models.RequestStatusApproved,
			ReviewNotes: "Self approval attempt",
		}

		w := PerformRequest(suite.Router, "PUT", "/api/v1/requests/"+requestID+"/status", suite.EmployeeToken, statusReq)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("IT admin approves request", func(t *testing.T) {
		statusReq := models.UpdateRequestStatusDTO{
			Status:      models.RequestStatusApproved,
			ReviewNotes: "Approved by IT, ordering from supplier",
		}

		w := PerformRequest(suite.Router, "PUT", "/api/v1/requests/"+requestID+"/status", suite.ITAdminToken, statusReq)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		assert.Equal(t, models.RequestStatusApproved, dataMap["status"])
	})
}

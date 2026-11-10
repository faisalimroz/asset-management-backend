package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestReportsAndUsers(t *testing.T) {
	suite := SetupTestApp(t)

	// Create a couple of assets and an assignment
	asset1 := suite.CreateTestAsset("Dell Precision 5570", "SN-PREC-5570", "Laptop", models.AssetStatusAvailable, 2500.00)
	asset2 := suite.CreateTestAsset("LG 34 UltraWide", "SN-LG-34UW", "Monitor", models.AssetStatusAvailable, 800.00)

	assignReq := models.AssignAssetRequest{
		AssetID: asset1.ID,
		UserID:  suite.EmployeeUser.ID,
		Notes:   "Assigned for workstation",
	}
	PerformRequest(suite.Router, "POST", "/api/v1/assignments", suite.ITAdminToken, assignReq)

	t.Run("IT admin can get inventory summary", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/reports/inventory", suite.ITAdminToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		assert.GreaterOrEqual(t, dataMap["total_assets"].(float64), float64(2))
		assert.GreaterOrEqual(t, dataMap["total_valuation"].(float64), float64(3300.00))
	})

	t.Run("HR can get inventory summary", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/reports/inventory", suite.HRToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Employee cannot get inventory summary (403 Forbidden)", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/reports/inventory", suite.EmployeeToken, nil)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("IT admin can view audit logs", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/reports/audit-logs", suite.ITAdminToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.GreaterOrEqual(t, resp.Meta.Total, int64(1))
	})

	t.Run("HR cannot view audit logs (403 Forbidden)", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/reports/audit-logs", suite.HRToken, nil)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("HR can view employees with active asset counts", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/users", suite.HRToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		users := resp.Data.([]interface{})
		assert.GreaterOrEqual(t, len(users), 1)
	})

	t.Run("HR can view specific employee allocated assets history", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/users/"+suite.EmployeeUser.ID+"/assets", suite.HRToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		allocations := resp.Data.([]interface{})
		assert.Equal(t, 1, len(allocations))
	})

	_ = asset2
}

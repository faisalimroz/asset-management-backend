package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestAssignment_Lifecycle(t *testing.T) {
	suite := SetupTestApp(t)

	// Create available asset
	asset := suite.CreateTestAsset("MacBook Pro 16", "SN-ASSIGN-01", "Laptop", models.AssetStatusAvailable, 2499.00)

	var assignmentID string

	t.Run("IT admin assigns available asset to employee", func(t *testing.T) {
		req := models.AssignAssetRequest{
			AssetID: asset.ID,
			UserID:  suite.EmployeeUser.ID,
			Notes:   "Assigned for new project work",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/assignments", suite.ITAdminToken, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		assignmentID = dataMap["id"].(string)
		assert.NotEmpty(t, assignmentID)

		// Verify asset status changed to assigned
		wAsset := PerformRequest(suite.Router, "GET", "/api/v1/assets/"+asset.ID, suite.ITAdminToken, nil)
		var assetResp utils.StandardResponse
		json.Unmarshal(wAsset.Body.Bytes(), &assetResp)
		assetData := assetResp.Data.(map[string]interface{})
		assert.Equal(t, models.AssetStatusAssigned, assetData["status"])
	})

	t.Run("Cannot assign already assigned asset", func(t *testing.T) {
		req := models.AssignAssetRequest{
			AssetID: asset.ID,
			UserID:  suite.EmployeeUser.ID,
			Notes:   "Attempt duplicate assignment",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/assignments", suite.ITAdminToken, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Employee views their own allocated assets", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/assignments/my-assets", suite.EmployeeToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		items := resp.Data.([]interface{})
		assert.GreaterOrEqual(t, len(items), 1)
	})

	t.Run("HR marks asset as returned during offboarding", func(t *testing.T) {
		returnReq := models.ReturnAssetRequest{
			ReturnedStatus: models.AssetStatusAvailable,
			Notes:          "Returned in pristine condition during employee offboarding",
		}

		w := PerformRequest(suite.Router, "PUT", "/api/v1/assignments/"+assignmentID+"/return", suite.HRToken, returnReq)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Verify asset status is back to available
		wAsset := PerformRequest(suite.Router, "GET", "/api/v1/assets/"+asset.ID, suite.ITAdminToken, nil)
		var assetResp utils.StandardResponse
		json.Unmarshal(wAsset.Body.Bytes(), &assetResp)
		assetData := assetResp.Data.(map[string]interface{})
		assert.Equal(t, models.AssetStatusAvailable, assetData["status"])
	})

	t.Run("Returning an already returned asset fails with 400 Bad Request", func(t *testing.T) {
		returnReq := models.ReturnAssetRequest{
			Notes: "Second return attempt",
		}

		w := PerformRequest(suite.Router, "PUT", "/api/v1/assignments/"+assignmentID+"/return", suite.ITAdminToken, returnReq)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Employee cannot assign asset (403 Forbidden)", func(t *testing.T) {
		req := models.AssignAssetRequest{
			AssetID: asset.ID,
			UserID:  suite.EmployeeUser.ID,
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/assignments", suite.EmployeeToken, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

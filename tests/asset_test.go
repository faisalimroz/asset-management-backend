package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"asset-management-backend/internal/models"
	"asset-management-backend/pkg/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAsset_Create(t *testing.T) {
	suite := SetupTestApp(t)

	t.Run("IT admin can create asset", func(t *testing.T) {
		req := models.CreateAssetRequest{
			Name:           "MacBook Pro 14 M3 Pro",
			SerialNumber:   "MBP14-" + uuid.New().String()[:8],
			Category:       "Laptop",
			Status:         models.AssetStatusAvailable,
			PurchaseDate:   time.Now(),
			PurchaseCost:   1999.00,
			Specifications: "18GB Unified RAM, 512GB SSD",
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/assets", suite.ITAdminToken, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		dataMap := resp.Data.(map[string]interface{})
		assert.Equal(t, req.Name, dataMap["name"])
		assert.Equal(t, req.SerialNumber, dataMap["serial_number"])
		assert.Equal(t, models.AssetStatusAvailable, dataMap["status"])
	})

	t.Run("Duplicate serial number fails with 409 Conflict", func(t *testing.T) {
		sn := "DUP-SN-" + uuid.New().String()[:8]
		req := models.CreateAssetRequest{
			Name:         "ThinkPad X1 Carbon",
			SerialNumber: sn,
			Category:     "Laptop",
			PurchaseDate: time.Now(),
			PurchaseCost: 1499.00,
		}

		w1 := PerformRequest(suite.Router, "POST", "/api/v1/assets", suite.ITAdminToken, req)
		assert.Equal(t, http.StatusCreated, w1.Code)

		// Second creation with same serial number
		w2 := PerformRequest(suite.Router, "POST", "/api/v1/assets", suite.ITAdminToken, req)
		assert.Equal(t, http.StatusConflict, w2.Code)
	})

	t.Run("Employee cannot create asset (403 Forbidden)", func(t *testing.T) {
		req := models.CreateAssetRequest{
			Name:         "Unauthorized Laptop",
			SerialNumber: "SN-FAIL-1",
			Category:     "Laptop",
			PurchaseDate: time.Now(),
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/assets", suite.EmployeeToken, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("HR cannot create asset (403 Forbidden)", func(t *testing.T) {
		req := models.CreateAssetRequest{
			Name:         "Unauthorized Laptop HR",
			SerialNumber: "SN-FAIL-2",
			Category:     "Laptop",
			PurchaseDate: time.Now(),
		}

		w := PerformRequest(suite.Router, "POST", "/api/v1/assets", suite.HRToken, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAsset_ListAndFilter(t *testing.T) {
	suite := SetupTestApp(t)

	// Create test assets
	suite.CreateTestAsset("Dell UltraSharp 32", "SN-MON-32", "Monitor", models.AssetStatusAvailable, 899.00)
	suite.CreateTestAsset("Dell UltraSharp 24", "SN-MON-24", "Monitor", models.AssetStatusAssigned, 399.00)
	suite.CreateTestAsset("MacBook Air M2", "SN-MBA-M2", "Laptop", models.AssetStatusAvailable, 1099.00)

	t.Run("IT admin can list assets with pagination", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/assets?page=1&limit=10", suite.ITAdminToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.GreaterOrEqual(t, resp.Meta.Total, int64(3))
	})

	t.Run("HR can list assets", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/assets", suite.HRToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Filter assets by category", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/assets?category=Monitor", suite.ITAdminToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp.Data.([]interface{})
		assert.Equal(t, 2, len(data))
	})

	t.Run("Filter assets by status", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/assets?status=assigned", suite.ITAdminToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.PaginatedResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp.Data.([]interface{})
		assert.Equal(t, 1, len(data))
	})

	t.Run("Employee cannot list all company assets (403 Forbidden)", func(t *testing.T) {
		w := PerformRequest(suite.Router, "GET", "/api/v1/assets", suite.EmployeeToken, nil)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAsset_UpdateAndDelete(t *testing.T) {
	suite := SetupTestApp(t)

	asset := suite.CreateTestAsset("Surface Pro 9", "SN-SURFACE-9", "Tablet", models.AssetStatusAvailable, 1200.00)

	t.Run("IT admin can update asset", func(t *testing.T) {
		newName := "Surface Pro 9 (Upgraded)"
		newStatus := models.AssetStatusMaintenance
		updateReq := models.UpdateAssetRequest{
			Name:   &newName,
			Status: &newStatus,
		}

		w := PerformRequest(suite.Router, "PUT", "/api/v1/assets/"+asset.ID, suite.ITAdminToken, updateReq)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp utils.StandardResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		dataMap := resp.Data.(map[string]interface{})
		assert.Equal(t, newName, dataMap["name"])
		assert.Equal(t, models.AssetStatusMaintenance, dataMap["status"])
	})

	t.Run("HR cannot update asset (403 Forbidden)", func(t *testing.T) {
		newName := "Hacked Name"
		updateReq := models.UpdateAssetRequest{Name: &newName}
		w := PerformRequest(suite.Router, "PUT", "/api/v1/assets/"+asset.ID, suite.HRToken, updateReq)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("IT admin can delete / retire asset", func(t *testing.T) {
		w := PerformRequest(suite.Router, "DELETE", "/api/v1/assets/"+asset.ID, suite.ITAdminToken, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify asset is no longer returned in active list
		wGet := PerformRequest(suite.Router, "GET", "/api/v1/assets/"+asset.ID, suite.ITAdminToken, nil)
		assert.Equal(t, http.StatusNotFound, wGet.Code)
	})
}

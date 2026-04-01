package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"purchase-order/config"
	"purchase-order/models"
	"purchase-order/routers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupCreateOrderTestDB initializes the test database and cleans it for create order tests
func setupCreateOrderTestDB() {
	// Initialize database connection if not already initialized
	if config.GetDB() == nil {
		cfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "123456",
			DBName:   "purchase_order_db",
		}
		if err := config.InitDatabase(cfg); err != nil {
			panic(err)
		}
	}

	db := config.GetDB()
	db.AutoMigrate(&models.PurchaseOrder{})
	db.Exec("DELETE FROM purchase_orders")
}

// setupCreateOrderTestRouter creates a test router in test mode for create order tests
func setupCreateOrderTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return routers.SetupRouter()
}

// TestCreateOrder_Success tests successful order creation with valid data
func TestCreateOrder_Success(t *testing.T) {
	setupCreateOrderTestDB()
	router := setupCreateOrderTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "12345.67",
		"status":       1,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusCreated, w.Code)

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Assert response structure
	assert.Equal(t, float64(0), response["code"])
	assert.Equal(t, "success", response["message"])
	assert.NotNil(t, response["data"])

	// Assert orderNo format
	data := response["data"].(map[string]interface{})
	orderNo := data["orderNo"].(string)
	assert.Contains(t, orderNo, "PO20260401")
	assert.Equal(t, "测试供应商", data["supplierName"])
	assert.Equal(t, "12345.67", data["totalAmount"])
	assert.Equal(t, float64(1), data["status"])
}

// TestCreateOrder_MissingSupplierName tests error when supplierName is missing
func TestCreateOrder_MissingSupplierName(t *testing.T) {
	setupCreateOrderTestDB()
	router := setupCreateOrderTestRouter()

	requestBody := map[string]interface{}{
		"orderDate":   "2026-04-01",
		"totalAmount": "12345.67",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Assert response code
	assert.Equal(t, float64(400), response["code"])
}

// TestCreateOrder_EmptySupplierName tests error when supplierName is whitespace only
func TestCreateOrder_EmptySupplierName(t *testing.T) {
	setupCreateOrderTestDB()
	router := setupCreateOrderTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "   ",
		"orderDate":    "2026-04-01",
		"totalAmount":  "12345.67",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Assert error message
	assert.Equal(t, float64(400), response["code"])
	assert.Contains(t, response["message"].(string), "供应商名称不能为空")
}

// TestCreateOrder_InvalidDateFormat tests error when date format is invalid
func TestCreateOrder_InvalidDateFormat(t *testing.T) {
	setupCreateOrderTestDB()
	router := setupCreateOrderTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "invalid-date",
		"totalAmount":  "12345.67",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Assert error message
	assert.Equal(t, float64(400), response["code"])
	assert.Contains(t, response["message"].(string), "日期格式无效")
}

// TestCreateOrder_InvalidAmount tests error when amount is negative
func TestCreateOrder_InvalidAmount(t *testing.T) {
	setupCreateOrderTestDB()
	router := setupCreateOrderTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "-100",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Assert error message
	assert.Equal(t, float64(400), response["code"])
	assert.Contains(t, response["message"].(string), "金额必须大于0")
}
func TestCreateOrder_TooManyDecimalPlaces(t *testing.T) {
	setupCreateOrderTestDB()
	router := setupCreateOrderTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "123.456",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Assert error message
	assert.Equal(t, float64(400), response["code"])
	assert.Contains(t, response["message"].(string), "金额最多两位小数")
}

// TestCreateOrder_SequentialOrderNumbers tests that order numbers are sequential
func TestCreateOrder_SequentialOrderNumbers(t *testing.T) {
	setupCreateOrderTestDB()
	router := setupCreateOrderTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "100.00",
	}

	// Create first order
	jsonBody1, _ := json.Marshal(requestBody)
	req1, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	data1 := response1["data"].(map[string]interface{})
	orderNo1 := data1["orderNo"].(string)

	// Assert first order number ends with 0001
	assert.Equal(t, "PO202604010001", orderNo1)

	// Create second order with same date
	requestBody["supplierName"] = "测试供应商2"
	jsonBody2, _ := json.Marshal(requestBody)
	req2, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusCreated, w2.Code)

	var response2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response2)
	data2 := response2["data"].(map[string]interface{})
	orderNo2 := data2["orderNo"].(string)

	// Assert second order number ends with 0002
	assert.Equal(t, "PO202604010002", orderNo2)
}

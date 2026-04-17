package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"purchase-order/config"
	"purchase-order/models"
	"purchase-order/routers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// CreateOrderTestSuite 创建订单测试套件
type CreateOrderTestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *gorm.DB
}

// SetupSuite 初始化测试套件
func (s *CreateOrderTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	cfg := config.DatabaseConfig{
		Host:         "localhost",
		Port:         3306,
		User:         "root",
		Password:     "123456",
		DBName:       "purchase_order_db",
		LogMode:      "silent",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		MaxLifetime:  time.Hour,
	}

	require.NoError(s.T(), config.InitDatabase(cfg))
	s.db = config.GetDB()
	s.db.AutoMigrate(&models.PurchaseOrder{})
}

// SetupTest 每个测试前的设置
func (s *CreateOrderTestSuite) SetupTest() {
	s.db.Exec("DELETE FROM purchase_orders")
	s.router = routers.SetupRouter()
}

// TearDownSuite 清理测试套件
func (s *CreateOrderTestSuite) TearDownSuite() {
	s.db.Exec("DELETE FROM purchase_orders")
}

// TestCreateOrder_Success 测试成功创建订单
func (s *CreateOrderTestSuite) TestCreateOrder_Success() {
	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "12345.67",
		"status":       1,
	}

	w := s.sendRequest(requestBody)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(s.T(), float64(0), response["code"])
	assert.Equal(s.T(), "success", response["message"])
	assert.NotNil(s.T(), response["data"])

	data := response["data"].(map[string]interface{})
	orderNo := data["orderNo"].(string)
	assert.Contains(s.T(), orderNo, "PO20260401")
	assert.Equal(s.T(), "测试供应商", data["supplierName"])
	assert.Equal(s.T(), "12345.67", data["totalAmount"])
	assert.Equal(s.T(), float64(1), data["status"])
}

// TestCreateOrder_MissingSupplierName 测试缺少供应商名称
func (s *CreateOrderTestSuite) TestCreateOrder_MissingSupplierName() {
	requestBody := map[string]interface{}{
		"orderDate":   "2026-04-01",
		"totalAmount": "12345.67",
	}

	w := s.sendRequest(requestBody)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(s.T(), float64(400), response["code"])
}

// TestCreateOrder_EmptySupplierName 测试空供应商名称
func (s *CreateOrderTestSuite) TestCreateOrder_EmptySupplierName() {
	requestBody := map[string]interface{}{
		"supplierName": "   ",
		"orderDate":    "2026-04-01",
		"totalAmount":  "12345.67",
	}

	w := s.sendRequest(requestBody)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(s.T(), float64(400), response["code"])
	assert.Contains(s.T(), response["message"].(string), "供应商名称不能为空")
}

// TestCreateOrder_InvalidDateFormat 测试无效日期格式
func (s *CreateOrderTestSuite) TestCreateOrder_InvalidDateFormat() {
	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "invalid-date",
		"totalAmount":  "12345.67",
	}

	w := s.sendRequest(requestBody)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(s.T(), float64(400), response["code"])
	assert.Contains(s.T(), response["message"].(string), "日期格式无效")
}

// TestCreateOrder_InvalidAmount 测试无效金额
func (s *CreateOrderTestSuite) TestCreateOrder_InvalidAmount() {
	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "-100",
	}

	w := s.sendRequest(requestBody)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(s.T(), float64(400), response["code"])
	assert.Contains(s.T(), response["message"].(string), "金额必须大于0")
}

// TestCreateOrder_TooManyDecimalPlaces 测试小数位过多
func (s *CreateOrderTestSuite) TestCreateOrder_TooManyDecimalPlaces() {
	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "123.456",
	}

	w := s.sendRequest(requestBody)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(s.T(), float64(400), response["code"])
	assert.Contains(s.T(), response["message"].(string), "金额最多两位小数")
}

// TestCreateOrder_SequentialOrderNumbers 测试订单号连续性
func (s *CreateOrderTestSuite) TestCreateOrder_SequentialOrderNumbers() {
	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "100.00",
	}

	// 创建第一个订单
	w1 := s.sendRequest(requestBody)
	assert.Equal(s.T(), http.StatusCreated, w1.Code)

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	data1 := response1["data"].(map[string]interface{})
	orderNo1 := data1["orderNo"].(string)

	assert.Equal(s.T(), "PO202604010001", orderNo1)

	// 创建第二个订单
	requestBody["supplierName"] = "测试供应商2"
	w2 := s.sendRequest(requestBody)
	assert.Equal(s.T(), http.StatusCreated, w2.Code)

	var response2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response2)
	data2 := response2["data"].(map[string]interface{})
	orderNo2 := data2["orderNo"].(string)

	assert.Equal(s.T(), "PO202604010002", orderNo2)
}

// sendRequest 发送请求的辅助方法
func (s *CreateOrderTestSuite) sendRequest(body map[string]interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// TestCreateOrderSuite 运行测试套件
func TestCreateOrderSuite(t *testing.T) {
	suite.Run(t, new(CreateOrderTestSuite))
}

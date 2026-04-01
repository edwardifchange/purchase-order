# Create Purchase Order Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-step. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a POST endpoint to create new purchase orders with auto-generated order numbers (format: PO + YYYYMMDD + daily sequence)

**Architecture:** Following existing clean architecture pattern: Router → Controller → Service → Repository → Model. Order number generation with transaction-based uniqueness guarantee and retry logic.

**Tech Stack:** Go 1.26.1, Gin framework, GORM, MySQL, shopspring/decimal

---

## File Structure

| File | Type | Responsibility |
|------|------|-----------------|
| `repositories/purchase_order_repository.go` | Modify | Add `GetMaxOrderNoByDate()` and `Create()` methods |
| `services/purchase_order_service.go` | Modify | Add `Create()` method with order number generation logic |
| `controllers/purchase_order_controller.go` | Modify | Add `Create()` method with request validation |
| `routers/router.go` | Modify | Add POST route for creating orders |
| `tests/create_order_test.go` | Create | Integration tests for create order endpoint |

---

## Task 1: Repository - Add GetMaxOrderNoByDate Method

**Files:**
- Modify: `repositories/purchase_order_repository.go:9` (interface)
- Modify: `repositories/purchase_order_repository.go:85` (after FindByID method)

- [ ] **Step 1: Update interface to add GetMaxOrderNoByDate**

Edit `PurchaseOrderRepository` interface at line 9, add method after `FindByID`:

```go
type PurchaseOrderRepository interface {
	FindAll(query ListQuery) ([]models.PurchaseOrder, int64, error)
	FindByID(id uint64) (*models.PurchaseOrder, error)
	GetMaxOrderNoByDate(dateStr string) (string, error)  // NEW
	Create(order *models.PurchaseOrder) error            // NEW
}
```

- [ ] **Step 2: Implement GetMaxOrderNoByDate method**

Add after the `FindByID` function (after line 85):

```go
func (r *purchaseOrderRepository) GetMaxOrderNoByDate(dateStr string) (string, error) {
	var maxOrderNo string
	prefix := "PO" + dateStr

	err := r.db.Model(&models.PurchaseOrder{}).
		Where("order_no LIKE ?", prefix+"%").
		Order("order_no DESC").
		Select("order_no").
		First(&maxOrderNo).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // No orders for this date yet
		}
		return "", err
	}

	return maxOrderNo, nil
}
```

- [ ] **Step 3: Run tests to verify no compilation errors**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add repositories/purchase_order_repository.go
git commit -m "feat(repo): add GetMaxOrderNoByDate method for order number generation"
```

---

## Task 2: Repository - Add Create Method

**Files:**
- Modify: `repositories/purchase_order_repository.go` (after GetMaxOrderNoByDate)

- [ ] **Step 1: Implement Create method**

Add after the `GetMaxOrderNoByDate` function:

```go
func (r *purchaseOrderRepository) Create(order *models.PurchaseOrder) error {
	return r.db.Create(order).Error
}
```

- [ ] **Step 2: Run tests to verify no compilation errors**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add repositories/purchase_order_repository.go
git commit -m "feat(repo): add Create method"
```

---

## Task 3: Service - Add Create Method with Order Number Generation

**Files:**
- Modify: `services/purchase_order_service.go:12` (interface)
- Modify: `services/purchase_order_service.go:69` (after GetByID method)

- [ ] **Step 1: Update interface to add Create method**

Edit `PurchaseOrderService` interface at line 12, add after `GetByID`:

```go
type PurchaseOrderService interface {
	GetList(query repositories.ListQuery) (*ListResult, error)
	GetByID(id uint64) (*models.PurchaseOrder, error)
	Create(order models.PurchaseOrder) (*models.PurchaseOrder, error) // NEW
}
```

- [ ] **Step 2: Implement helper function generateOrderNo**

Add after the imports (before the interface, around line 11):

```go
const maxRetries = 3

func generateOrderNo(date time.Time, repo repositories.PurchaseOrderRepository) (string, error) {
	dateStr := date.Format("20060102")
	prefix := "PO" + dateStr

	for i := 0; i < maxRetries; i++ {
		maxOrderNo, err := repo.GetMaxOrderNoByDate(dateStr)
		if err != nil && err != gorm.ErrRecordNotFound {
			return "", err
		}

		seq := 1
		if maxOrderNo != "" {
			seqStr := maxOrderNo[len(maxOrderNo)-4:]
			s, _ := strconv.Atoi(seqStr)
			seq = s + 1
		}

		orderNo := fmt.Sprintf("%s%04d", prefix, seq)
		return orderNo, nil
	}

	return "", errors.New("failed to generate unique order number")
}
```

**Note:** You'll need to add imports at the top of the file:
```go
import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"purchase-order/models"
	"purchase-order/repositories"

	"gorm.io/gorm"
)
```

- [ ] **Step 3: Implement Create method**

Add after the `GetByID` function (after line 69):

```go
func (s *purchaseOrderService) Create(order models.PurchaseOrder) (*models.PurchaseOrder, error) {
	// Set default status if not provided
	if order.Status == 0 {
		order.Status = models.StatusPending
	}

	// Validate status range
	if order.Status < 1 || order.Status > 6 {
		return nil, errors.New("状态值必须在1-6之间")
	}

	// Generate order number
	orderNo, err := generateOrderNo(order.OrderDate, s.repo)
	if err != nil {
		return nil, err
	}
	order.OrderNo = orderNo

	// Create in database
	if err := s.repo.Create(&order); err != nil {
		return nil, err
	}

	return &order, nil
}
```

- [ ] **Step 4: Run tests to verify no compilation errors**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add services/purchase_order_service.go
git commit -m "feat(service): add Create method with order number generation"
```

---

## Task 4: Controller - Add Create Method with Validation

**Files:**
- Modify: `controllers/purchase_order_controller.go` (after GetByID method)

- [ ] **Step 1: Add imports for time and decimal**

At the top of the file, update imports:

```go
import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"purchase-order/repositories"
	"purchase-order/responses"
	"purchase-order/services"
	"purchase-order/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)
```

- [ ] **Step 2: Add CreateRequest struct**

Add after the controller struct definition (after line 35):

```go
// CreateRequest 创建采购订单请求
type CreateRequest struct {
	SupplierName string `json:"supplierName" binding:"required"`
	OrderDate    string `json:"orderDate" binding:"required"`
	TotalAmount  string `json:"totalAmount" binding:"required"`
	Status       int8   `json:"status"`
}
```

- [ ] **Step 3: Implement Create method**

Add at the end of the file (after parseInt function):

```go
// Create 创建采购订单
func (c *PurchaseOrderController) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, responses.Error(400, "参数格式错误"))
		return
	}

	// Validate supplierName
	if strings.TrimSpace(req.SupplierName) == "" {
		ctx.JSON(http.StatusBadRequest, responses.Error(400, "供应商名称不能为空"))
		return
	}
	if len(req.SupplierName) > 100 {
		ctx.JSON(http.StatusBadRequest, responses.Error(400, "供应商名称最多100字符"))
		return
	}

	// Parse orderDate
	orderDate, err := time.Parse("2006-01-02", req.OrderDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, responses.Error(400, "日期格式无效，请使用YYYY-MM-DD格式"))
		return
	}

	// Parse totalAmount
	totalAmount, err := decimal.NewFromString(req.TotalAmount)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, responses.Error(400, "金额格式无效"))
		return
	}
	if totalAmount.LessThanOrEqual(decimal.Zero) {
		ctx.JSON(http.StatusBadRequest, responses.Error(400, "金额必须大于0"))
		return
	}

	// Validate status (optional field)
	status := req.Status
	if status != 0 && (status < 1 || status > 6) {
		ctx.JSON(http.StatusBadRequest, responses.Error(400, "状态值必须在1-6之间"))
		return
	}
	if status == 0 {
		status = models.StatusPending
	}

	// Create purchase order
	order := models.PurchaseOrder{
		SupplierName: req.SupplierName,
		OrderDate:    orderDate,
		TotalAmount:  totalAmount,
		Status:       status,
	}

	result, err := c.service.Create(order)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responses.Error(500, err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, responses.Success(result))
}
```

- [ ] **Step 4: Run tests to verify no compilation errors**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add controllers/purchase_order_controller.go
git commit -m "feat(controller): add Create method with validation"
```

---

## Task 5: Router - Add POST Route

**Files:**
- Modify: `routers/router.go:25` (purchaseOrders route group)

- [ ] **Step 1: Add POST route**

Edit the purchaseOrders route group (around line 25-29):

```go
purchaseOrders := api.Group("/purchase-orders")
{
	purchaseOrders.GET("", purchaseOrderController.GetList)
	purchaseOrders.GET("/:poId", purchaseOrderController.GetByID)
	purchaseOrders.POST("", purchaseOrderController.Create)  // NEW
}
```

- [ ] **Step 2: Run tests to verify no compilation errors**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add routers/router.go
git commit -m "feat(router): add POST route for creating purchase orders"
```

---

## Task 6: Tests - Write Integration Tests

**Files:**
- Create: `tests/create_order_test.go`

- [ ] **Step 1: Create test file and setup**

Create new file `tests/create_order_test.go`:

```go
package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"purchase-order/config"
	"purchase-order/controllers"
	"purchase-order/models"
	"purchase-order/repositories"
	"purchase-order/routers"
	"purchase-order/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestDB() {
	db := config.GetDB()
	// Auto migrate
	db.AutoMigrate(&models.PurchaseOrder{})
	// Clean table
	db.Exec("DELETE FROM purchase_orders")
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return routers.SetupRouter()
}

func TestCreateOrder_Success(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "12345.67",
		"status":       1,
	}

	jsonBody, _ := json.Marshal(requestBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(0), response["code"])
	assert.Equal(t, "success", response["message"])

	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["orderNo"])
	assert.Contains(t, data["orderNo"], "PO20260401")
}

func TestCreateOrder_MissingSupplierName(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	requestBody := map[string]interface{}{
		"orderDate":   "2026-04-01",
		"totalAmount": "12345.67",
	}

	jsonBody, _ := json.Marshal(requestBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(400), response["code"])
}

func TestCreateOrder_EmptySupplierName(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "   ",
		"orderDate":    "2026-04-01",
		"totalAmount":  "12345.67",
	}

	jsonBody, _ := json.Marshal(requestBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(400), response["code"])
	assert.Contains(t, response["message"], "供应商名称不能为空")
}

func TestCreateOrder_InvalidDateFormat(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "invalid-date",
		"totalAmount":  "12345.67",
	}

	jsonBody, _ := json.Marshal(requestBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(400), response["code"])
	assert.Contains(t, response["message"], "日期格式无效")
}

func TestCreateOrder_InvalidAmount(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "-100",
	}

	jsonBody, _ := json.Marshal(requestBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(400), response["code"])
	assert.Contains(t, response["message"], "金额必须大于0")
}

func TestCreateOrder_SequentialOrderNumbers(t *testing.T) {
	setupTestDB()
	router := setupTestRouter()

	requestBody := map[string]interface{}{
		"supplierName": "测试供应商",
		"orderDate":    "2026-04-01",
		"totalAmount":  "1000.00",
	}

	// Create first order
	jsonBody, _ := json.Marshal(requestBody)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	orderNo1 := response1["data"].(map[string]interface{})["orderNo"].(string)

	// Create second order
	requestBody["supplierName"] = "另一个供应商"
	jsonBody2, _ := json.Marshal(requestBody)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/v1/purchase-orders", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	var response2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response2)
	orderNo2 := response2["data"].(map[string]interface{})["orderNo"].(string)

	// Verify order numbers are sequential
	assert.Contains(t, orderNo1, "PO202604010001")
	assert.Contains(t, orderNo2, "PO202604010002")
}
```

- [ ] **Step 2: Run tests to verify they fail (TDD - tests first)**

Run: `go test -v ./tests`
Expected: Tests may fail or pass depending on implementation

- [ ] **Step 3: Commit**

```bash
git add tests/create_order_test.go
git commit -m "test: add integration tests for create purchase order endpoint"
```

---

## Task 7: Run Full Test Suite and Verify

- [ ] **Step 1: Run all tests**

Run: `go test -v ./...`
Expected: All tests pass

- [ ] **Step 2: Test the API manually (optional)**

Run: `go run main.go`

Then test with curl:
```bash
curl -X POST http://localhost:8080/api/v1/purchase-orders \
  -H "Content-Type: application/json" \
  -d '{"supplierName":"测试供应商","orderDate":"2026-04-01","totalAmount":"12345.67","status":1}'
```

Expected: Returns 201 Created with order data containing auto-generated orderNo

- [ ] **Step 3: Final commit if any fixes needed**

```bash
git add .
git commit -m "fix: any minor fixes from testing"
```

---

## Task 8: Cleanup and Documentation

- [ ] **Step 1: Run go mod tidy**

Run: `go mod tidy`
Expected: No changes needed

- [ ] **Step 2: Format code**

Run: `go fmt ./...`
Expected: Code formatted

- [ ] **Step 3: Final verification build**

Run: `go build -o purchase-order.exe main.go`
Expected: Binary created successfully

- [ ] **Step 4: Final commit**

```bash
git add .
git commit -m "chore: final cleanup and formatting"
```

---

## Summary

This implementation adds:
1. Repository methods: `GetMaxOrderNoByDate()`, `Create()`
2. Service method: `Create()` with order number generation (PO + YYYYMMDD + daily sequence)
3. Controller method: `Create()` with full validation
4. POST route: `/api/v1/purchase-orders`
5. Comprehensive integration tests

The order number format is: `PO202604010001` (PO + 8-digit date + 4-digit daily sequence starting from 0001).

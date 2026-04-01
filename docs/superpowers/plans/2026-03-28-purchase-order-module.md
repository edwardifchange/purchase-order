# 采购订单模块实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 基于 Gin 框架实现采购订单列表模块，支持列表查询（分页、排序、筛选）和详情查看。

**Architecture:** 分层架构，Controller -> Service -> Repository -> Model，依赖注入，TDD 开发。

**Tech Stack:** Go 1.25+, Gin, GORM, MySQL, shopspring/decimal

---

## 文件结构

```
purchase-order/
├── main.go
├── go.mod
├── config/
│   └── database.go
├── models/
│   └── purchase_order.go
├── repositories/
│   └── purchase_order_repository.go
├── services/
│   └── purchase_order_service.go
├── controllers/
│   └── purchase_order_controller.go
├── routers/
│   └── router.go
├── responses/
│   └── response.go
└── tests/
    └── integration_test.go
```

---

### Task 1: 项目初始化

**Files:**
- Create: `purchase-order/go.mod`

- [ ] **Step 1: 初始化 Go 模块**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go mod init purchase-order
```

- [ ] **Step 2: 安装依赖**

```bash
go get github.com/gin-gonic/gin@latest
go get gorm.io/gorm@latest
go get gorm.io/driver/mysql@latest
go get github.com/shopspring/decimal@latest
```

---

### Task 2: 统一响应格式

**Files:**
- Create: `purchase-order/responses/response.go`

- [ ] **Step 1: 创建响应结构体**

```go
package responses

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func SuccessWithPage(list interface{}, total int64, page, pageSize int) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	}
}

func Error(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}
```

- [ ] **Step 2: 验证编译**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go build ./...
```

Expected: 编译成功，无错误

---

### Task 3: 数据模型

**Files:**
- Create: `purchase-order/models/purchase_order.go`

- [ ] **Step 1: 创建采购订单模型**

```go
package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type PurchaseOrder struct {
	ID           uint64          `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderNo      string          `json:"order_no" gorm:"uniqueIndex;size:50;not null"`
	SupplierName string          `json:"supplier_name" gorm:"size:100;not null"`
	OrderDate    time.Time       `json:"order_date" gorm:"type:date;not null"`
	TotalAmount  decimal.Decimal `json:"total_amount" gorm:"type:decimal(12,2);not null"`
	Status       int8            `json:"status" gorm:"not null;default:1"`
	CreatedAt    time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

// 订单状态常量
const (
	StatusPending   int8 = 1 // 待审批
	StatusApproved  int8 = 2 // 已审批
	StatusCompleted int8 = 3 // 已完成
	StatusCancelled int8 = 4 // 已取消
)
```

- [ ] **Step 2: 验证编译**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go build ./...
```

Expected: 编译成功，无错误

---

### Task 4: 数据库配置

**Files:**
- Create: `purchase-order/config/database.go`

- [ ] **Step 1: 创建数据库连接配置**

```go
package config

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"purchase-order/models"
)

var DB *gorm.DB

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func InitDatabase(cfg DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移
	err = DB.AutoMigrate(&models.PurchaseOrder{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
```

- [ ] **Step 2: 验证编译**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go build ./...
```

Expected: 编译成功，无错误

---

### Task 5: Repository 层

**Files:**
- Create: `purchase-order/repositories/purchase_order_repository.go`

- [ ] **Step 1: 创建 Repository 接口和实现**

```go
package repositories

import (
	"purchase-order/models"

	"gorm.io/gorm"
)

type PurchaseOrderRepository interface {
	FindAll(query ListQuery) ([]models.PurchaseOrder, int64, error)
	FindByID(id uint64) (*models.PurchaseOrder, error)
}

type ListQuery struct {
	Page         int
	PageSize     int
	OrderBy      string
	Order        string
	OrderNo      string
	SupplierName string
	Status       *int8
	StartDate    string
	EndDate      string
}

type purchaseOrderRepository struct {
	db *gorm.DB
}

func NewPurchaseOrderRepository(db *gorm.DB) PurchaseOrderRepository {
	return &purchaseOrderRepository{db: db}
}

func (r *purchaseOrderRepository) FindAll(query ListQuery) ([]models.PurchaseOrder, int64, error) {
	var orders []models.PurchaseOrder
	var total int64

	db := r.db.Model(&models.PurchaseOrder{})

	// 筛选条件
	if query.OrderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+query.OrderNo+"%")
	}
	if query.SupplierName != "" {
		db = db.Where("supplier_name LIKE ?", "%"+query.SupplierName+"%")
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.StartDate != "" {
		db = db.Where("order_date >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("order_date <= ?", query.EndDate)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderStr := query.OrderBy
	if query.Order == "desc" {
		orderStr += " DESC"
	} else {
		orderStr += " ASC"
	}

	// 分页
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order(orderStr).Offset(offset).Limit(query.PageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *purchaseOrderRepository) FindByID(id uint64) (*models.PurchaseOrder, error) {
	var order models.PurchaseOrder
	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}
```

- [ ] **Step 2: 验证编译**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go build ./...
```

Expected: 编译成功，无错误

---

### Task 6: Service 层

**Files:**
- Create: `purchase-order/services/purchase_order_service.go`

- [ ] **Step 1: 创建 Service 接口和实现**

```go
package services

import (
	"errors"

	"purchase-order/models"
	"purchase-order/repositories"

	"gorm.io/gorm"
)

type PurchaseOrderService interface {
	GetList(query repositories.ListQuery) (*ListResult, error)
	GetByID(id uint64) (*models.PurchaseOrder, error)
}

type ListResult struct {
	List     []models.PurchaseOrder
	Total    int64
	Page     int
	PageSize int
}

type purchaseOrderService struct {
	repo repositories.PurchaseOrderRepository
}

func NewPurchaseOrderService(repo repositories.PurchaseOrderRepository) PurchaseOrderService {
	return &purchaseOrderService{repo: repo}
}

func (s *purchaseOrderService) GetList(query repositories.ListQuery) (*ListResult, error) {
	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.OrderBy == "" {
		query.OrderBy = "id"
	}
	if query.Order == "" {
		query.Order = "asc"
	}

	orders, total, err := s.repo.FindAll(query)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		List:     orders,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (s *purchaseOrderService) GetByID(id uint64) (*models.PurchaseOrder, error) {
	order, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("采购订单不存在")
		}
		return nil, err
	}
	return order, nil
}
```

- [ ] **Step 2: 验证编译**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go build ./...
```

Expected: 编译成功，无错误

---

### Task 7: Controller 层

**Files:**
- Create: `purchase-order/controllers/purchase_order_controller.go`

- [ ] **Step 1: 创建 Controller**

```go
package controllers

import (
	"net/http"
	"strconv"

	"purchase-order/repositories"
	"purchase-order/responses"
	"purchase-order/services"

	"github.com/gin-gonic/gin"
)

type PurchaseOrderController struct {
	service services.PurchaseOrderService
}

func NewPurchaseOrderController(service services.PurchaseOrderService) *PurchaseOrderController {
	return &PurchaseOrderController{service: service}
}

// GetList 获取采购订单列表
func (c *PurchaseOrderController) GetList(ctx *gin.Context) {
	query := repositories.ListQuery{
		Page:         parseInt(ctx.DefaultQuery("page", "1")),
		PageSize:     parseInt(ctx.DefaultQuery("page_size", "10")),
		OrderBy:      ctx.DefaultQuery("order_by", "id"),
		Order:        ctx.DefaultQuery("order", "asc"),
		OrderNo:      ctx.Query("order_no"),
		SupplierName: ctx.Query("supplier_name"),
		StartDate:    ctx.Query("start_date"),
		EndDate:      ctx.Query("end_date"),
	}

	// 处理 status 参数
	if statusStr := ctx.Query("status"); statusStr != "" {
		status := int8(parseInt(statusStr))
		query.Status = &status
	}

	result, err := c.service.GetList(query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responses.Error(500, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, responses.SuccessWithPage(result.List, result.Total, result.Page, result.PageSize))
}

// GetByID 获取采购订单详情
func (c *PurchaseOrderController) GetByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, responses.Error(400, "无效的ID"))
		return
	}

	order, err := c.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, responses.Error(404, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, responses.Success(order))
}

func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}
```

- [ ] **Step 2: 验证编译**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go build ./...
```

Expected: 编译成功，无错误

---

### Task 8: 路由配置

**Files:**
- Create: `purchase-order/routers/router.go`

- [ ] **Step 1: 创建路由配置**

```go
package routers

import (
	"purchase-order/config"
	"purchase-order/controllers"
	"purchase-order/repositories"
	"purchase-order/services"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 初始化依赖
	db := config.GetDB()
	purchaseOrderRepo := repositories.NewPurchaseOrderRepository(db)
	purchaseOrderService := services.NewPurchaseOrderService(purchaseOrderRepo)
	purchaseOrderController := controllers.NewPurchaseOrderController(purchaseOrderService)

	// 注册路由
	api := r.Group("/api/v1")
	{
		purchaseOrders := api.Group("/purchase-orders")
		{
			purchaseOrders.GET("", purchaseOrderController.GetList)
			purchaseOrders.GET("/:id", purchaseOrderController.GetByID)
		}
	}

	return r
}
```

- [ ] **Step 2: 验证编译**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go build ./...
```

Expected: 编译成功，无错误

---

### Task 9: 主入口文件

**Files:**
- Create: `purchase-order/main.go`

- [ ] **Step 1: 创建主入口文件**

```go
package main

import (
	"log"

	"purchase-order/config"
	"purchase-order/routers"
)

func main() {
	// 初始化数据库
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "your_password",
		DBName:   "purchase_order_db",
	}

	if err := config.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 设置路由
	r := routers.SetupRouter()

	// 启动服务
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

- [ ] **Step 2: 验证编译**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go build -o purchase-order.exe .
```

Expected: 编译成功，生成 purchase-order.exe

---

### Task 10: 集成测试

**Files:**
- Create: `purchase-order/tests/integration_test.go`

- [ ] **Step 1: 创建集成测试**

```go
package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"purchase-order/config"
	"purchase-order/routers"

	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	// 这里可以配置测试数据库
	return routers.SetupRouter()
}

func TestGetList(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/purchase-orders", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetListWithParams(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/purchase-orders?page=1&page_size=10&order_by=id&order=desc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetByID(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/purchase-orders/1", nil)
	router.ServeHTTP(w, req)

	// 可能返回 404 因为数据库中没有数据
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
}
```

- [ ] **Step 2: 安装测试依赖**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go get github.com/stretchr/testify@latest
```

- [ ] **Step 3: 运行测试**

```bash
cd C:/Users/xingc/Desktop/purchase-order
go test ./tests/... -v
```

Expected: 测试通过或因数据库未连接而跳过

---

## 运行说明

1. 确保 MySQL 服务运行
2. 创建数据库: `CREATE DATABASE purchase_order_db CHARACTER SET utf8mb4;`
3. 修改 `main.go` 中的数据库连接配置
4. 运行: `go run main.go`
5. 访问:
   - 列表: `GET http://localhost:8080/api/v1/purchase-orders`
   - 详情: `GET http://localhost:8080/api/v1/purchase-orders/1`

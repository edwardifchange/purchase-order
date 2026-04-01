# 采购订单模块设计文档

## 概述

基于 Gin 框架开发采购订单列表模块，包含列表展示（支持排序、筛选）和详情查看功能。

## 技术栈

- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL
- **Go 版本**: 1.25+

## 数据模型设计

### 采购订单表 `purchase_orders`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键，自增 |
| order_no | varchar(50) | 订单号，唯一索引 |
| supplier_name | varchar(100) | 供应商名称 |
| order_date | date | 订单日期 |
| total_amount | decimal(12,2) | 总金额 |
| status | tinyint | 状态：1-待审批 2-已审批 3-已完成 4-已取消 5-已结算 6-已到货 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### Go 结构体

```go
type PurchaseOrder struct {
    ID           uint64          `json:"id" gorm:"primaryKey"`
    OrderNo      string          `json:"order_no" gorm:"uniqueIndex;size:50"`
    SupplierName string          `json:"supplier_name" gorm:"size:100"`
    OrderDate    time.Time       `json:"order_date"`
    TotalAmount  decimal.Decimal `json:"total_amount" gorm:"type:decimal(12,2)"`
    Status       int8            `json:"status"`
    CreatedAt    time.Time       `json:"created_at"`
    UpdatedAt    time.Time       `json:"updated_at"`
}
```

## API 接口设计

### 1. 获取采购订单列表

**请求**
```
GET /api/v1/purchase-orders
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 10 |
| order_by | string | 否 | 排序字段，默认 "id" |
| order | string | 否 | 排序方向：asc/desc，默认 "asc" |
| order_no | string | 否 | 订单号，模糊搜索 |
| supplier_name | string | 否 | 供应商名称，模糊搜索 |
| status | int | 否 | 订单状态 |
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "order_no": "PO20240328001",
        "supplier_name": "供应商A",
        "order_date": "2024-03-28",
        "total_amount": "10000.00",
        "status": 1,
        "created_at": "2024-03-28 10:00:00",
        "updated_at": "2024-03-28 10:00:00"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### 2. 获取采购订单详情

**请求**
```
GET /api/v1/purchase-orders/:id
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "order_no": "PO20240328001",
    "supplier_name": "供应商A",
    "order_date": "2024-03-28",
    "total_amount": "10000.00",
    "status": 1,
    "created_at": "2024-03-28 10:00:00",
    "updated_at": "2024-03-28 10:00:00"
  }
}
```

**错误响应**
```json
{
  "code": 404,
  "message": "采购订单不存在",
  "data": null
}
```

## 项目结构

```
purchase-order/
├── main.go                      # 入口文件
├── config/
│   └── database.go              # 数据库连接配置
├── models/
│   └── purchase_order.go        # 数据模型定义
├── controllers/
│   └── purchase_order_controller.go  # 处理 HTTP 请求
├── services/
│   └── purchase_order_service.go     # 业务逻辑处理
├── repositories/
│   └── purchase_order_repository.go  # 数据库操作
├── routers/
│   └── router.go                # 路由配置
├── responses/
│   └── response.go              # 统一响应格式
└── go.mod
```

## 分层架构

```
┌─────────────┐
│  Router     │  路由分发
└──────┬──────┘
       │
┌──────▼──────┐
│ Controller  │  请求处理、参数校验
└──────┬──────┘
       │
┌──────▼──────┐
│  Service    │  业务逻辑
└──────┬──────┘
       │
┌──────▼──────┐
│ Repository  │  数据访问
└──────┬──────┘
       │
┌──────▼──────┐
│   Model     │  数据结构
└─────────────┘
```

## 功能清单

- [x] 采购订单列表查询（分页）
- [x] 按 ID 排序（支持升序/降序）
- [x] 按订单号筛选（模糊搜索）
- [x] 按供应商名称筛选（模糊搜索）
- [x] 按订单状态筛选
- [x] 按日期范围筛选
- [x] 采购订单详情查询

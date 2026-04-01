# 创建采购订单模块设计文档

**日期：** 2026-04-01
**作者：** Claude Code
**状态：** 已批准

## 1. 概述

为采购订单管理系统添加创建采购订单的功能，采用服务端自动生成订单号的方案。

## 2. API 接口设计

### 2.1 端点

```
POST /api/v1/purchase-orders
```

### 2.2 请求体

```json
{
  "supplierName": "供应商名称",
  "orderDate": "2026-04-01",
  "totalAmount": "12345.67",
  "status": 1
}
```

### 2.3 响应

成功响应 (201 Created):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "orderNo": "PO202604010001",
    "supplierName": "供应商名称",
    "orderDate": "2026-04-01T00:00:00Z",
    "totalAmount": "12345.67",
    "status": 1,
    "createdAt": "2026-04-01T10:30:00Z",
    "updatedAt": "2026-04-01T10:30:00Z"
  }
}
```

错误响应:
```json
{
  "code": 400,
  "message": "供应商名称不能为空",
  "data": null
}
```

## 3. 订单号生成逻辑

### 3.1 格式定义

```
PO + YYYYMMDD + 每日序号
```

示例: `PO202604010001`

| 部分 | 长度 | 说明 |
|------|------|------|
| 前缀 | 2字符 | 固定值 "PO" |
| 日期 | 8字符 | 订单日期，格式 YYYYMMDD |
| 序号 | 4字符 | 当日订单序号，从 0001 开始 |

### 3.2 生成步骤

1. 获取订单日期（如果未提供则使用当前日期）
2. 构建日期前缀：`PO20260401`
3. 查询数据库中该日期前缀的最大订单号
4. 提取序号部分并 +1
5. 格式化为4位数字（补零）
6. 拼接生成完整订单号
7. 在事务中检查唯一性，如失败则重试（最多3次）

### 3.3 并发处理

使用数据库事务保证订单号的唯一性。当检测到订单号重复时，进行重试。

## 4. 架构设计

### 4.1 层次结构

```
┌─────────────────────────────────────────┐
│         Router (router.go)              │
│  POST /api/v1/purchase-orders           │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│    Controller (purchase_order_...       │
│    - 解析请求参数                        │
│    - 参数验证                            │
│    - 调用 Service                        │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│     Service (purchase_order_...         │
│     - 生成订单号                         │
│     - 事务管理                           │
│     - 调用 Repository                    │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│   Repository (purchase_order_...        │
│   - GORM 创建记录                        │
│   - 返回结果                             │
└─────────────────────────────────────────┘
```

### 4.2 各层职责

**Controller 层:**
- 解析 JSON 请求体
- 验证必填字段
- 验证字段格式和范围
- 调用 Service 层
- 返回 HTTP 响应

**Service 层:**
- 生成唯一订单号
- 管理数据库事务
- 处理重试逻辑
- 调用 Repository 层

**Repository 层:**
- 查询指定日期的最大订单号
- 创建新记录
- 返回创建结果

## 5. 验证规则

| 字段 | 类型 | 必填 | 默认值 | 验证规则 |
|------|------|------|--------|----------|
| supplierName | string | 是 | - | 长度 1-100 字符 |
| orderDate | string | 是 | 当前日期 | 有效日期格式 |
| totalAmount | string | 是 | - | 正数，最多2位小数 |
| status | int | 否 | 1 | 范围 1-6 |

## 6. 错误处理

| 错误场景 | HTTP状态码 | 错误码 | 消息 |
|----------|-----------|--------|------|
| JSON 解析失败 | 400 | 400 | 参数格式错误 |
| supplierName 为空 | 400 | 400 | 供应商名称不能为空 |
| supplierName 超长 | 400 | 400 | 供应商名称最多100字符 |
| orderDate 格式无效 | 400 | 400 | 日期格式无效 |
| totalAmount 无效 | 400 | 400 | 金额必须大于0 |
| status 超出范围 | 400 | 400 | 状态值无效 |
| 订单号生成失败 | 500 | 500 | 订单号生成失败 |
| 数据库错误 | 500 | 500 | 内部服务器错误 |

## 7. 实现清单

### 7.1 新增/修改的文件

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `repositories/purchase_order_repository.go` | 新增方法 | `GetMaxOrderNoByDate(date string)`, `Create(order *PurchaseOrder)` |
| `services/purchase_order_service.go` | 新增方法 | `Create(order models.PurchaseOrder) (*models.PurchaseOrder, error)` |
| `controllers/purchase_order_controller.go` | 新增方法 | `Create(ctx *gin.Context)` |
| `routers/router.go` | 新增路由 | `purchaseOrders.POST("", purchaseOrderController.Create)` |
| `tests/create_order_test.go` | 新增文件 | 创建订单的集成测试 |

### 7.2 Repository 层实现

**新增方法:**

```go
// GetMaxOrderNoByDate 获取指定日期的最大订单号
func (r *PurchaseOrderRepository) GetMaxOrderNoByDate(date string) (string, error)

// Create 创建采购订单
func (r *PurchaseOrderRepository) Create(order *models.PurchaseOrder) error
```

### 7.3 Service 层实现

**新增方法:**

```go
// Create 创建采购订单
func (s *PurchaseOrderService) Create(order models.PurchaseOrder) (*models.PurchaseOrder, error)
```

**订单号生成逻辑:**

```go
func generateOrderNo(date time.Time, repo PurchaseOrderRepository) (string, error) {
    const maxRetries = 3

    for i := 0; i < maxRetries; i++ {
        dateStr := date.Format("20060102")
        prefix := "PO" + dateStr

        maxOrderNo, err := repo.GetMaxOrderNoByDate(dateStr)
        if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
            return "", err
        }

        seq := 1
        if maxOrderNo != "" {
            seqStr := maxOrderNo[len(maxOrderNo)-4:]
            seq, _ = strconv.Atoi(seqStr)
            seq++
        }

        orderNo := fmt.Sprintf("%s%04d", prefix, seq)
        return orderNo, nil
    }

    return "", errors.New("failed to generate unique order number")
}
```

### 7.4 Controller 层实现

**新增方法:**

```go
// CreateRequest 创建请求
type CreateRequest struct {
    SupplierName string          `json:"supplierName" binding:"required"`
    OrderDate    string          `json:"orderDate" binding:"required"`
    TotalAmount  decimal.Decimal `json:"totalAmount" binding:"required"`
    Status       int8            `json:"status"`
}

// Create 创建采购订单
func (c *PurchaseOrderController) Create(ctx *gin.Context)
```

**验证逻辑:**

1. 供应商名称：1-100 字符
2. 订单日期：解析为 time.Time，格式 YYYY-MM-DD
3. 总金额：使用 decimal 解析，必须大于 0
4. 状态：默认 1，范围 1-6

### 7.5 路由配置

```go
purchaseOrders := api.Group("/purchase-orders")
{
    purchaseOrders.GET("", purchaseOrderController.GetList)
    purchaseOrders.GET("/:poId", purchaseOrderController.GetByID)
    purchaseOrders.POST("", purchaseOrderController.Create)  // 新增
}
```

## 8. 测试计划

### 8.1 单元测试

- Repository 层：测试创建记录和查询最大订单号
- Service 层：测试订单号生成逻辑

### 8.2 集成测试

- 成功创建订单
- 缺少必填字段
- 供应商名称为空
- 日期格式无效
- 金额无效
- 状态超出范围
- 并发创建订单（验证订单号唯一性）

## 9. 后续扩展

- 添加订单明细（items）关联
- 支持批量创建订单
- 添加审批流程接口
- 添加订单状态变更接口

# 采购订单管理系统 API

基于 Go 语言开发的采购订单管理 REST API，提供订单的创建、查询和列表功能。

## 功能特性

- ✅ 创建采购订单（自动生成订单号）
- ✅ 查询采购订单详情
- ✅ 分页查询采购订单列表
- ✅ 按多条件筛选（订单号、供应商、状态、日期范围）
- ✅ 支持排序（多字段、升序/降序）

## 技术栈

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.26.1 | 编程语言 |
| Gin | latest | Web 框架 |
| GORM | latest | ORM 框架 |
| MySQL | 5.7+ | 数据库 |
| shopspring/decimal | latest | 精确的十进制数值计算 |

## 项目结构

```
purchase-order/
├── config/              # 配置和数据库连接
│   └── database.go
├── controllers/         # 控制器层（HTTP 请求处理）
│   └── purchase_order_controller.go
├── models/              # 数据模型
│   └── purchase_order.go
├── repositories/        # 数据访问层
│   └── purchase_order_repository.go
├── responses/           # 统一响应格式
│   └── response.go
├── routers/             # 路由配置
│   └── router.go
├── services/            # 业务逻辑层
│   └── purchase_order_service.go
├── tests/               # 集成测试
│   ├── integration_test.go
│   └── create_order_test.go
├── main.go              # 应用入口
└── README.md            # 项目文档
```

### 架构模式

项目采用**分层架构**，遵循依赖注入原则：

```
Router → Controller → Service → Repository → Model
```

各层职责清晰：
- **Router**: 路由配置和依赖注入
- **Controller**: HTTP 请求处理、参数验证、响应封装
- **Service**: 业务逻辑、事务管理
- **Repository**: 数据库操作
- **Model**: 数据结构定义

## 环境要求

- Go 1.26.1 或更高版本
- MySQL 5.7 或更高版本
- Git（用于克隆代码）

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/edwardifchange/purchase-order.git
cd purchase-order
```

### 2. 配置数据库

创建 MySQL 数据库：

```sql
CREATE DATABASE purchase_order_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

修改 `config/database.go` 中的数据库连接信息：

```go
cfg := config.DatabaseConfig{
    Host:     "localhost",       // MySQL 主机地址
    Port:     3306,              // MySQL 端口
    User:     "root",            // MySQL 用户名
    Password: os.Getenv("DB_PASSWORD"),  // MySQL 密码（从环境变量读取）
    DBName:   "purchase_order_db",      // 数据库名
}
```

**设置环境变量：**

Windows (PowerShell):
```powershell
$env:DB_PASSWORD="your_password"
```

Windows (CMD):
```cmd
set DB_PASSWORD=your_password
```

Linux/Mac:
```bash
export DB_PASSWORD="your_password"
```

### 3. 安装依赖

```bash
go mod download
```

### 4. 运行应用

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

### 5. 构建可执行文件

```bash
go build -o purchase-order.exe main.go
./purchase-order.exe
```

## API 接口文档

### 基础信息

- **Base URL**: `http://localhost:8080`
- **API 版本**: `v1`
- **Content-Type**: `application/json`

### 响应格式

所有 API 响应使用统一的 JSON 格式：

**成功响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

**分页响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [ ... ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

**错误响应：**
```json
{
  "code": 400,
  "message": "错误信息",
  "data": null
}
```

---

### 1. 创建采购订单

**接口：** `POST /api/v1/purchase-orders`

**请求体：**
```json
{
  "supplierName": "供应商名称",
  "orderDate": "2026-04-01",
  "totalAmount": "12345.67",
  "status": 1
}
```

**请求参数说明：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| supplierName | string | 是 | 供应商名称，最多 100 个字符 |
| orderDate | string | 是 | 订单日期，格式：YYYY-MM-DD |
| totalAmount | string | 是 | 订单总金额，最多 2 位小数，必须大于 0 |
| status | int | 否 | 订单状态（1-6），默认为 1 |

**订单状态说明：**

| 值 | 状态 | 说明 |
|----|------|------|
| 1 | 待审批 | StatusPending |
| 2 | 已审批 | StatusApproved |
| 3 | 已完成 | StatusCompleted |
| 4 | 已取消 | StatusCancelled |
| 5 | 已结算 | StatusSettled |
| 6 | 已到货 | StatusDelivered |

**响应示例：**

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

**使用 curl 测试：**
```bash
curl -X POST http://localhost:8080/api/v1/purchase-orders \
  -H "Content-Type: application/json" \
  -d '{
    "supplierName": "测试供应商",
    "orderDate": "2026-04-01",
    "totalAmount": "12345.67",
    "status": 1
  }'
```

---

### 2. 查询采购订单列表

**接口：** `GET /api/v1/purchase-orders`

**查询参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（最大 100） |
| order_by | string | 否 | id | 排序字段 |
| order | string | 否 | asc | 排序方向（asc/desc） |
| order_no | string | 否 | - | 订单号（模糊查询） |
| supplier_name | string | 否 | - | 供应商名称（模糊查询） |
| status | int | 否 | - | 订单状态（1-6） |
| start_date | string | 否 | - | 开始日期（YYYY-MM-DD） |
| end_date | string | 否 | - | 结束日期（YYYY-MM-DD） |

**可排序字段：**
- `id` - 订单 ID
- `order_no` - 订单号
- `supplier_name` - 供应商名称
- `order_date` - 订单日期
- `total_amount` - 订单金额
- `status` - 订单状态
- `created_at` - 创建时间
- `updated_at` - 更新时间

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "orderNo": "PO202604010001",
        "supplierName": "供应商名称",
        "orderDate": "2026-04-01T00:00:00Z",
        "totalAmount": "12345.67",
        "status": 1,
        "createdAt": "2026-04-01T10:30:00Z",
        "updatedAt": "2026-04-01T10:30:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

**使用 curl 测试：**
```bash
# 获取第一页，每页 10 条
curl "http://localhost:8080/api/v1/purchase-orders?page=1&page_size=10"

# 按订单日期降序排序
curl "http://localhost:8080/api/v1/purchase-orders?order_by=order_date&order=desc"

# 搜索特定供应商
curl "http://localhost:8080/api/v1/purchase-orders?supplier_name=测试供应商"

# 筛选特定状态的订单
curl "http://localhost:8080/api/v1/purchase-orders?status=1"

# 日期范围查询
curl "http://localhost:8080/api/v1/purchase-orders?start_date=2026-04-01&end_date=2026-04-30"
```

---

### 3. 查询采购订单详情

**接口：** `GET /api/v1/purchase-orders/:poId`

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| poId | uint64 | 订单 ID |

**响应示例：**
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

**使用 curl 测试：**
```bash
curl "http://localhost:8080/api/v1/purchase-orders/1"
```

---

## 订单号规则

订单号由系统自动生成，格式为：**PO + 日期(YYYYMMDD) + 当日序号(0001)**

**示例：**
- `PO202604010001` - 2026年4月1日的第1个订单
- `PO202604010002` - 2026年4月1日的第2个订单
- `PO202604020001` - 2026年4月2日的第1个订单

**并发处理：**
系统实现了并发冲突重试机制，当多个请求同时创建订单时，会自动重试最多3次以避免订单号冲突。

## 错误码说明

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 200 | 0 | 成功 |
| 201 | 0 | 创建成功 |
| 400 | 400 | 请求参数错误 |
| 400 | 400 | 供应商名称不能为空 |
| 400 | 400 | 供应商名称最多100字符 |
| 400 | 400 | 日期格式无效 |
| 400 | 400 | 金额格式无效 |
| 400 | 400 | 金额必须大于0 |
| 400 | 400 | 金额最多两位小数 |
| 400 | 400 | 状态值必须在1-6之间 |
| 400 | 400 | 订单日期不能为空 |
| 404 | 404 | 采购订单不存在 |
| 500 | 500 | 内部服务器错误 |
| 500 | 500 | 创建订单失败：并发冲突，请稍后重试 |

## 开发指南

### 代码规范

- 遵循 Go 语言官方代码风格
- 使用 `go fmt` 格式化代码
- 使用 `go vet` 检查代码问题

### 添加新接口

在 `routers/router.go` 中按照依赖注入模式添加新路由：

```go
// 1. 获取数据库实例
db := config.GetDB()

// 2. 初始化 repository
repo := repositories.NewXxxRepository(db)

// 3. 初始化 service
service := services.NewXxxService(repo)

// 4. 初始化 controller
controller := controllers.NewXxxController(service)

// 5. 注册路由
api.Group("/resource").GET("", controller.Method)
```

### 数据库迁移

项目使用 GORM 的 AutoMigrate 功能，应用启动时自动创建/更新表结构。

## 测试

### 运行所有测试

```bash
go test ./...
```

### 运行集成测试

```bash
go test -v ./tests
```

### 运行特定测试

```bash
# 运行单个测试文件
go test -v ./tests/create_order_test.go

# 运行特定测试用例
go test -v ./tests -run TestCreateOrder_Success
```

### 测试覆盖率

```bash
go test -cover ./...
```

## 常见问题

### Q: 如何修改数据库连接信息？

A: 编辑 `config/database.go` 文件，修改 `DatabaseConfig` 结构体的值。

### Q: 订单号可以自定义吗？

A: 不可以，订单号由系统自动生成，确保唯一性和连续性。

### Q: 如何处理并发创建订单？

A: 系统已内置并发冲突重试机制，当检测到订单号冲突时会自动重试最多3次。

### Q: 金额字段为什么使用 decimal 类型？

A: 使用 `shopspring/decimal` 类型可以避免浮点数精度问题，确保财务计算的准确性。

### Q: 如何部署到生产环境？

A:
1. 修改 `config/database.go` 中的生产数据库配置
2. 构建可执行文件：`go build -o purchase-order main.go`
3. 运行：`./purchase-order`
4. 推荐使用 systemd 或 supervisor 管理进程

## 贡献指南

欢迎提交 Issue 和 Pull Request！

提交 PR 前，请确保：
- 所有测试通过
- 代码已格式化
- 添加了必要的测试用例
- 更新了相关文档

## 许可证

MIT License

## 联系方式

- 项目地址：[https://github.com/edwardifchange/purchase-order](https://github.com/edwardifchange/purchase-order)
- Issues: [https://github.com/edwardifchange/purchase-order/issues](https://github.com/edwardifchange/purchase-order/issues)

---

**最后更新：** 2026-04-01

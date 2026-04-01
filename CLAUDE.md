# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based REST API for a Purchase Order management system built with the Gin framework. The system uses GORM as the ORM and MySQL as the database.

## Architecture

The project follows a clean architecture pattern with clear separation of concerns:

- **Models**: Data structures and database schema
- **Repositories**: Data access layer
- **Services**: Business logic layer
- **Controllers**: HTTP request handling layer
- **Routers**: Route configuration
- **Config**: Configuration and database setup
- **Responses**: Standardized API response formats

Key architectural pattern:
```
Router → Controller → Service → Repository → Model
```

### Adding New Endpoints

When adding new endpoints, follow the dependency injection pattern in [routers/router.go](routers/router.go):

```go
// 1. Get the database instance
db := config.GetDB()

// 2. Initialize repository
repo := repositories.NewXxxRepository(db)

// 3. Initialize service with repository
service := services.NewXxxService(repo)

// 4. Initialize controller with service
controller := controllers.NewXxxController(service)

// 5. Register routes
api.Group("/resource").GET("", controller.Method)
```

## Database Configuration

The database is configured in [config/database.go](config/database.go) with default settings:
- Host: localhost
- Port: 3306
- User: root
- Password: 123456
- Database: purchase_order_db

The system uses GORM with automatic migrations. The PurchaseOrder model includes status constants:
- StatusPending = 1 (待审批)
- StatusApproved = 2 (已审批)
- StatusCompleted = 3 (已完成)
- StatusCancelled = 4 (已取消)
- StatusSettled = 5 (已结算)

## Common Development Commands

### Building and Running
```bash
# Run the server (listens on :8080 by default)
go run main.go

# Build the binary
go build -o purchase-order main.go

# Run the built binary
./purchase-order
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package (integration tests are in tests/)
go test ./tests

# Run integration tests with coverage
go test -cover ./tests
```

Note: Integration tests are located in the `tests/` directory at the project root.

### Code Quality
```bash
# Format code
go fmt ./...

# Check for issues
go vet ./...

# Tidy dependencies
go mod tidy
```

## API Endpoints

The system provides two main endpoints under `/api/v1/purchase-orders`:

1. **GET /api/v1/purchase-orders** - Get list with pagination and filtering
   - Query parameters: page, page_size, order_by, order, order_no, supplier_name, status, start_date, end_date

2. **GET /api/v1/purchase-orders/:poId** - Get purchase order by ID
   - Route parameter naming uses resource-specific prefixes (e.g., `poId` for purchase order)

## Response Format

All API responses use the standardized format defined in [responses/response.go](responses/response.go):

```go
// Success response
{
  "code": 0,
  "message": "success",
  "data": { ... }
}

// Success with pagination
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

// Error response
{
  "code": 404,
  "message": "error message",
  "data": null
}
```

## Dependencies

Key dependencies:
- `github.com/gin-gonic/gin` - Web framework
- `gorm.io/gorm` - ORM
- `gorm.io/driver/mysql` - MySQL driver
- `github.com/shopspring/decimal` - For precise decimal arithmetic
- `github.com/stretchr/testify` - Testing framework

**Go version**: 1.26.1

## Important Implementation Details

### JSON Field Naming
- Model structs use **camelCase** JSON tags (e.g., `orderNo`, `supplierName`, `totalAmount`)
- Database columns use **snake_case** (e.g., `order_no`, `supplier_name`)
- GORM handles the mapping automatically

### Decimal Handling
The `TotalAmount` field uses `shopspring/decimal` for precise financial calculations. Always use this type for monetary values to avoid floating-point precision issues.

### Validation Patterns

Controller-level validation in [controllers/purchase_order_controller.go](controllers/purchase_order_controller.go):

- **Page size**: Capped at maximum 100
- **Order by fields**: Whitelisted to prevent SQL injection - only allowed fields can be used
  ```go
  var allowedOrderByFields = map[string]bool{
      "id": true, "order_no": true, "supplier_name": true,
      "order_date": true, "total_amount": true, "status": true,
      "created_at": true, "updated_at": true,
  }
  ```
- **Status values**: Must be between 1-5 (inclusive)
- **Order direction**: Only "asc" or "desc" allowed

When adding new endpoints with sorting, maintain the same whitelist pattern for security.

### Error Handling

Services return errors that controllers handle:
- Parse/validation errors → HTTP 400
- Not found errors → HTTP 404
- Internal errors → HTTP 500

### Database Schema

The `purchase_orders` table includes:
- id (bigint, primary key, auto increment)
- order_no (varchar(50), unique, not null)
- supplier_name (varchar(100), not null)
- order_date (date, not null)
- total_amount (decimal(12,2), not null)
- status (tinyint, not null, default 1) - 1:待审批 2:已审批 3:已完成 4:已取消 5:已结算
- created_at, updated_at (datetime, auto managed)

### Status Constants

Defined in [models/purchase_order.go](models/purchase_order.go):
```go
const (
    StatusPending   int8 = 1 // 待审批
    StatusApproved  int8 = 2 // 已审批
    StatusCompleted int8 = 3 // 已完成
    StatusCancelled int8 = 4 // 已取消
    StatusSettled   int8 = 5 // 已结算
)
```
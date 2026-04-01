package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"purchase-order/models"
	"purchase-order/repositories"
	"purchase-order/responses"
	"purchase-order/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 允许排序的字段白名单
var allowedOrderByFields = map[string]bool{
	"id":            true,
	"order_no":      true,
	"supplier_name": true,
	"order_date":    true,
	"total_amount":  true,
	"status":        true,
	"created_at":    true,
	"updated_at":    true,
}

type PurchaseOrderController struct {
	service services.PurchaseOrderService
}

func NewPurchaseOrderController(service services.PurchaseOrderService) *PurchaseOrderController {
	return &PurchaseOrderController{service: service}
}

// CreateRequest 创建采购订单请求
type CreateRequest struct {
	SupplierName string `json:"supplierName" binding:"required"`
	OrderDate    string `json:"orderDate" binding:"required"`
	TotalAmount  string `json:"totalAmount" binding:"required"`
	Status       int8   `json:"status"`
}

// GetList 获取采购订单列表
func (c *PurchaseOrderController) GetList(ctx *gin.Context) {
	// 解析分页参数并校验
	page := parseInt(ctx.DefaultQuery("page", "1"))
	pageSize := parseInt(ctx.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 校验排序字段
	orderBy := ctx.DefaultQuery("order_by", "id")
	if !allowedOrderByFields[orderBy] {
		orderBy = "id"
	}

	// 校验排序方向
	order := strings.ToLower(ctx.DefaultQuery("order", "asc"))
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	query := repositories.ListQuery{
		Page:         page,
		PageSize:     pageSize,
		OrderBy:      orderBy,
		Order:        order,
		OrderNo:      ctx.Query("order_no"),
		SupplierName: ctx.Query("supplier_name"),
		StartDate:    ctx.Query("start_date"),
		EndDate:      ctx.Query("end_date"),
	}

	// 处理 status 参数
	if statusStr := ctx.Query("status"); statusStr != "" {
		status := int8(parseInt(statusStr))
		// 校验 status 是否在有效范围内
		if status >= 1 && status <= 6 {
			query.Status = &status
		}
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
	idStr := ctx.Param("poId")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, responses.Error(400, "无效的ID"))
		return
	}

	order, err := c.service.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, responses.Error(404, err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, responses.Error(500, err.Error()))
		}
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
		// Count characters (runes) instead of bytes
		if len([]rune(req.SupplierName)) > 100 {
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
	// Validate decimal places - max 2 digits for currency
		// Extract fractional part and check length
		decimalStr := totalAmount.String()
		if strings.Contains(decimalStr, ".") {
		fractionalPart := strings.SplitN(decimalStr, ".", 2)[1]
		if len(fractionalPart) > 2 {
		ctx.JSON(http.StatusBadRequest, responses.Error(400, "金额最多两位小数"))
		return
		}
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

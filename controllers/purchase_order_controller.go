package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"purchase-order/models"
	"purchase-order/repositories"
	"purchase-order/responses"
	"purchase-order/services"
)

const (
	defaultRequestTimeout = 30 * time.Second
)

// 错误码定义
const (
	ErrCodeSuccess         = 0
	ErrCodeInvalidParams   = 40001
	ErrCodeNotFound        = 40401
	ErrCodeInternalError   = 50001
	ErrCodeConcurrentError = 50002
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

// CreateRequest 创建采购订单请求
type CreateRequest struct {
	SupplierName string `json:"supplierName" binding:"required"`
	OrderDate    string `json:"orderDate" binding:"required"`
	TotalAmount  string `json:"totalAmount" binding:"required"`
	Status       int8   `json:"status"`
}

// NewPurchaseOrderController 创建采购订单控制器
func NewPurchaseOrderController(service services.PurchaseOrderService) *PurchaseOrderController {
	return &PurchaseOrderController{service: service}
}

// GetList 获取采购订单列表
func (c *PurchaseOrderController) GetList(ctx *gin.Context) {
	// 设置超时上下文
	reqCtx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	// 解析并验证参数
	query, err := c.parseListQuery(ctx)
	if err != nil {
		responses.SendBadRequest(ctx, err.Error())
		return
	}

	// 调用服务
	result, err := c.service.GetList(reqCtx, query)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	responses.SendJSON(ctx, http.StatusOK, responses.SuccessWithPage(result.List, result.Total, result.Page, result.PageSize))
}

// GetByID 获取采购订单详情
func (c *PurchaseOrderController) GetByID(ctx *gin.Context) {
	idStr := ctx.Param("poId")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		responses.SendBadRequest(ctx, "无效的ID")
		return
	}

	// 设置超时上下文
	reqCtx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	order, err := c.service.GetByID(reqCtx, id)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	responses.SendSuccess(ctx, order)
}

// Create 创建采购订单
func (c *PurchaseOrderController) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		responses.SendBadRequest(ctx, "参数格式错误")
		return
	}

	// 验证请求
	if err := c.validateCreateRequest(&req); err != nil {
		responses.SendBadRequest(ctx, err.Error())
		return
	}

	// 构造订单
	order, err := c.buildOrderFromRequest(&req)
	if err != nil {
		responses.SendBadRequest(ctx, err.Error())
		return
	}

	// 设置超时上下文
	reqCtx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	// 调用服务
	result, err := c.service.Create(reqCtx, *order)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	responses.SendCreated(ctx, result)
}

// validateCreateRequest 验证创建请求
func (c *PurchaseOrderController) validateCreateRequest(req *CreateRequest) error {
	// 验证供应商名称
	if strings.TrimSpace(req.SupplierName) == "" {
		return errors.New("供应商名称不能为空")
	}
	if len([]rune(req.SupplierName)) > 100 {
		return errors.New("供应商名称最多100字符")
	}

	// 验证金额
	amount, err := decimal.NewFromString(req.TotalAmount)
	if err != nil {
		return errors.New("金额格式无效")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("金额必须大于0")
	}
	// 验证小数位数
	if strings.Contains(req.TotalAmount, ".") {
		fractionalPart := strings.SplitN(req.TotalAmount, ".", 2)[1]
		if len(fractionalPart) > 2 {
			return errors.New("金额最多两位小数")
		}
	}

	// 验证状态
	if req.Status != 0 && !models.IsValidStatus(req.Status) {
		return errors.New("状态值必须在1-6之间")
	}

	return nil
}

// buildOrderFromRequest 从请求构造订单对象
func (c *PurchaseOrderController) buildOrderFromRequest(req *CreateRequest) (*models.PurchaseOrder, error) {
	orderDate, err := time.Parse("2006-01-02", req.OrderDate)
	if err != nil {
		return nil, errors.New("日期格式无效，请使用YYYY-MM-DD格式")
	}

	totalAmount, err := decimal.NewFromString(req.TotalAmount)
	if err != nil {
		return nil, errors.New("金额格式无效")
	}

	status := req.Status
	if status == 0 {
		status = models.StatusPending
	}

	return &models.PurchaseOrder{
		SupplierName: req.SupplierName,
		OrderDate:    orderDate,
		TotalAmount:  totalAmount,
		Status:       status,
	}, nil
}

// parseListQuery 解析列表查询参数
func (c *PurchaseOrderController) parseListQuery(ctx *gin.Context) (repositories.ListQuery, error) {
	page := parseInt(ctx.DefaultQuery("page", "1"))
	pageSize := parseInt(ctx.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	orderBy := ctx.DefaultQuery("order_by", "id")
	if !isAllowedOrderByField(orderBy) {
		orderBy = "id"
	}

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

	if statusStr := ctx.Query("status"); statusStr != "" {
		status := int8(parseInt(statusStr))
		if models.IsValidStatus(status) {
			query.Status = &status
		}
	}

	return query, nil
}

// handleError 错误处理
func (c *PurchaseOrderController) handleError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrOrderNotFound):
		responses.SendNotFound(ctx, err.Error())
	case errors.Is(err, services.ErrConcurrentConflict):
		responses.SendConflict(ctx, err.Error())
	default:
		responses.SendInternalError(ctx, "内部服务错误")
	}
}

// parseInt 解析整数
func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

// isAllowedOrderByField 检查是否为允许的排序字段
func isAllowedOrderByField(field string) bool {
	return allowedOrderByFields[field]
}

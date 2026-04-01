package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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
		if status >= 1 && status <= 5 {
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

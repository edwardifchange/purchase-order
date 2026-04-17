package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"purchase-order/models"
	"purchase-order/repositories"

	"gorm.io/gorm"
)

const (
	seqSuffixLength = 4
	maxCreateRetries = 5
	defaultRetryDelay = 10 * time.Millisecond
)

// 业务错误定义
var (
	ErrOrderDateRequired  = errors.New("订单日期不能为空")
	ErrInvalidStatus      = errors.New("状态值必须在1-6之间")
	ErrAmountTooSmall     = errors.New("金额必须大于0")
	ErrConcurrentConflict = errors.New("系统繁忙，请稍后重试")
	ErrOrderNotFound      = errors.New("采购订单不存在")
)

// PurchaseOrderService 采购订单服务接口
type PurchaseOrderService interface {
	GetList(ctx context.Context, query repositories.ListQuery) (*ListResult, error)
	GetByID(ctx context.Context, id uint64) (*models.PurchaseOrder, error)
	Create(ctx context.Context, order models.PurchaseOrder) (*models.PurchaseOrder, error)
}

// ListResult 列表结果
type ListResult struct {
	List     []models.PurchaseOrder
	Total    int64
	Page     int
	PageSize int
}

type purchaseOrderService struct {
	repo repositories.PurchaseOrderRepository
}

// NewPurchaseOrderService 创建采购订单服务
func NewPurchaseOrderService(repo repositories.PurchaseOrderRepository) PurchaseOrderService {
	return &purchaseOrderService{repo: repo}
}

// GetList 获取采购订单列表
func (s *purchaseOrderService) GetList(ctx context.Context, query repositories.ListQuery) (*ListResult, error) {
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

	orders, total, err := s.repo.FindAll(ctx, query)
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

// GetByID 根据ID获取采购订单
func (s *purchaseOrderService) GetByID(ctx context.Context, id uint64) (*models.PurchaseOrder, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return order, nil
}

// Create 创建采购订单
func (s *purchaseOrderService) Create(ctx context.Context, order models.PurchaseOrder) (*models.PurchaseOrder, error) {
	// 验证订单
	if err := s.validateOrder(order); err != nil {
		return nil, err
	}

	// 重试机制
	var result *models.PurchaseOrder
	var lastErr error

	for attempt := 0; attempt < maxCreateRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(defaultRetryDelay)
		}

		// 生成订单号
		orderNo, err := s.generateOrderNo(ctx, order.OrderDate)
		if err != nil {
			lastErr = err
			continue
		}
		order.OrderNo = orderNo

		// 创建订单
		if err := s.repo.Create(ctx, &order); err != nil {
			if s.isDuplicateKeyError(err) {
				lastErr = err
				continue // 并发冲突，重试
			}
			return nil, err
		}

		result = &order
		return result, nil
	}

	// 所有重试都失败
	if s.isDuplicateKeyError(lastErr) {
		return nil, ErrConcurrentConflict
	}
	return nil, lastErr
}

// validateOrder 验证订单数据
func (s *purchaseOrderService) validateOrder(order models.PurchaseOrder) error {
	if order.OrderDate.IsZero() {
		return ErrOrderDateRequired
	}

	// 设置默认状态
	if order.Status == 0 {
		order.Status = models.StatusPending
	}

	// 验证状态
	if !models.IsValidStatus(order.Status) {
		return ErrInvalidStatus
	}

	// 验证金额
	if order.TotalAmount.LessThanOrEqual(decimal.Zero) {
		return ErrAmountTooSmall
	}

	return nil
}

// generateOrderNo 生成订单号
func (s *purchaseOrderService) generateOrderNo(ctx context.Context, date time.Time) (string, error) {
	dateStr := date.Format("20060102")
	prefix := "PO" + dateStr

	maxOrderNo, err := s.repo.GetMaxOrderNoByDate(ctx, dateStr)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	seq := 1
	if maxOrderNo != "" {
		seqStr := maxOrderNo[len(maxOrderNo)-seqSuffixLength:]
		s, err := strconv.Atoi(seqStr)
		if err != nil {
			return "", errors.New("解析订单号序号失败")
		}
		seq = s + 1
	}

	return fmt.Sprintf("%s%04d", prefix, seq), nil
}

// isDuplicateKeyError 判断是否为重复键错误
func (s *purchaseOrderService) isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "Duplicate entry") ||
		strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "UNIQUE constraint failed")
}

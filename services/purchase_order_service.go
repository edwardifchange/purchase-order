package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"purchase-order/models"
	"purchase-order/repositories"

	"gorm.io/gorm"
)

const seqSuffixLength = 4
const maxCreateRetries = 3

// generateOrderNo 生成采购订单号
// 格式: PO + YYYYMMDD + 4位序号 (例如: PO202604010001)
// 参数: date - 订单日期, repo - 仓库实例
// 返回: 生成的订单号, 错误信息
func generateOrderNo(date time.Time, repo repositories.PurchaseOrderRepository) (string, error) {
	dateStr := date.Format("20060102")
	prefix := "PO" + dateStr

	maxOrderNo, err := repo.GetMaxOrderNoByDate(dateStr)
	if err != nil && err != gorm.ErrRecordNotFound {
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

	orderNo := fmt.Sprintf("%s%04d", prefix, seq)
	return orderNo, nil
}

type PurchaseOrderService interface {
	GetList(query repositories.ListQuery) (*ListResult, error)
	GetByID(id uint64) (*models.PurchaseOrder, error)
	Create(order models.PurchaseOrder) (*models.PurchaseOrder, error)
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

// Create 创建采购订单
// 自动生成订单号并保存到数据库
// 参数: order - 采购订单对象 (OrderNo 会自动生成,不需要传入)
// 返回: 创建后的采购订单指针, 错误信息
func (s *purchaseOrderService) Create(order models.PurchaseOrder) (*models.PurchaseOrder, error) {
	// Validate OrderDate is provided
	if order.OrderDate.IsZero() {
		return nil, errors.New("订单日期不能为空")
	}

	// Set default status if not provided
	if order.Status == 0 {
		order.Status = models.StatusPending
	}

	// Validate status range
	if order.Status < 1 || order.Status > 6 {
		return nil, errors.New("状态值必须在1-6之间")
	}

	// Retry loop for handling concurrent order number conflicts
	for attempt := 0; attempt < maxCreateRetries; attempt++ {
		// Generate order number
		orderNo, err := generateOrderNo(order.OrderDate, s.repo)
		if err != nil {
			return nil, err
		}
		order.OrderNo = orderNo

		// Create in database
		err = s.repo.Create(&order)
		if err == nil {
			// Success - return the created order
			return &order, nil
		}

		// Check if error is duplicate key (concurrent conflict)
		// MySQL duplicate key error typically contains "Duplicate entry"
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "duplicate key") {
			// Conflict detected - retry and generate new order number
			continue
		}

		// Other error - don't retry
		return nil, err
	}

	// All retries exhausted
	return nil, errors.New("创建订单失败：并发冲突，请稍后重试")
}

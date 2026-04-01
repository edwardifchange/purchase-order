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
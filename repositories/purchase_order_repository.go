package repositories

import (
	"purchase-order/models"

	"gorm.io/gorm"
)

type PurchaseOrderRepository interface {
	FindAll(query ListQuery) ([]models.PurchaseOrder, int64, error)
	FindByID(id uint64) (*models.PurchaseOrder, error)
}

type ListQuery struct {
	Page         int
	PageSize     int
	OrderBy      string
	Order        string
	OrderNo      string
	SupplierName string
	Status       *int8
	StartDate    string
	EndDate      string
}

type purchaseOrderRepository struct {
	db *gorm.DB
}

func NewPurchaseOrderRepository(db *gorm.DB) PurchaseOrderRepository {
	return &purchaseOrderRepository{db: db}
}

func (r *purchaseOrderRepository) FindAll(query ListQuery) ([]models.PurchaseOrder, int64, error) {
	var orders []models.PurchaseOrder
	var total int64

	db := r.db.Model(&models.PurchaseOrder{})

	// 筛选条件
	if query.OrderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+query.OrderNo+"%")
	}
	if query.SupplierName != "" {
		db = db.Where("supplier_name LIKE ?", "%"+query.SupplierName+"%")
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.StartDate != "" {
		db = db.Where("order_date >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("order_date <= ?", query.EndDate)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderStr := query.OrderBy
	if query.Order == "desc" {
		orderStr += " DESC"
	} else {
		orderStr += " ASC"
	}

	// 分页
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order(orderStr).Offset(offset).Limit(query.PageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *purchaseOrderRepository) FindByID(id uint64) (*models.PurchaseOrder, error) {
	var order models.PurchaseOrder
	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}
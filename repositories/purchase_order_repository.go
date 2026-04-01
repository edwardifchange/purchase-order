package repositories

import (
	"purchase-order/models"

	"gorm.io/gorm"
)

type PurchaseOrderRepository interface {
	FindAll(query ListQuery) ([]models.PurchaseOrder, int64, error)
	FindByID(id uint64) (*models.PurchaseOrder, error)
	GetMaxOrderNoByDate(dateStr string) (string, error)
	Create(order *models.PurchaseOrder) error
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

// GetMaxOrderNoByDate 获取指定日期的最大订单号
// dateStr: 日期字符串，格式为 YYYYMMDD (例如: "20260401")
// 返回: 最大订单号，如果该日期没有订单则返回空字符串 (无错误)
// 错误: 仅在数据库查询出错时返回错误
func (r *purchaseOrderRepository) GetMaxOrderNoByDate(dateStr string) (string, error) {
	var maxOrderNo string
	prefix := "PO" + dateStr

	err := r.db.Model(&models.PurchaseOrder{}).
		Where("order_no LIKE ?", prefix+"%").
		Order("order_no DESC").
		Select("order_no").
		First(&maxOrderNo).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // No orders for this date yet
		}
		return "", err
	}

	return maxOrderNo, nil
}

// Create 创建采购订单记录
func (r *purchaseOrderRepository) Create(order *models.PurchaseOrder) error {
	return r.db.Create(order).Error
}

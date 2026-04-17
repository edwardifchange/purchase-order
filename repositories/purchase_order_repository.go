package repositories

import (
	"context"
	"fmt"
	"purchase-order/models"

	"gorm.io/gorm"
)

// PurchaseOrderRepository 采购订单仓储接口
type PurchaseOrderRepository interface {
	FindAll(ctx context.Context, query ListQuery) ([]models.PurchaseOrder, int64, error)
	FindByID(ctx context.Context, id uint64) (*models.PurchaseOrder, error)
	FindByOrderNo(ctx context.Context, orderNo string) (*models.PurchaseOrder, error)
	GetMaxOrderNoByDate(ctx context.Context, dateStr string) (string, error)
	Create(ctx context.Context, order *models.PurchaseOrder) error
	CreateBatch(ctx context.Context, orders []*models.PurchaseOrder) error
	Update(ctx context.Context, order *models.PurchaseOrder) error
	Delete(ctx context.Context, id uint64) error
	WithTx(tx *gorm.DB) PurchaseOrderRepository
}

// ListQuery 列表查询参数
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

// NewPurchaseOrderRepository 创建采购订单仓储
func NewPurchaseOrderRepository(db *gorm.DB) PurchaseOrderRepository {
	return &purchaseOrderRepository{db: db}
}

// WithTx 创建事务仓储
func (r *purchaseOrderRepository) WithTx(tx *gorm.DB) PurchaseOrderRepository {
	return &purchaseOrderRepository{db: tx}
}

// FindAll 查询采购订单列表
func (r *purchaseOrderRepository) FindAll(ctx context.Context, query ListQuery) ([]models.PurchaseOrder, int64, error) {
	var orders []models.PurchaseOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&models.PurchaseOrder{})

	// 应用筛选条件
	db = r.applyFilters(db, query)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 构建排序
	orderStr := fmt.Sprintf("%s %s", query.OrderBy, query.Order)

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order(orderStr).Offset(offset).Limit(query.PageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// FindByID 根据ID查询采购订单
func (r *purchaseOrderRepository) FindByID(ctx context.Context, id uint64) (*models.PurchaseOrder, error) {
	var order models.PurchaseOrder
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByOrderNo 根据订单号查询采购订单
func (r *purchaseOrderRepository) FindByOrderNo(ctx context.Context, orderNo string) (*models.PurchaseOrder, error) {
	var order models.PurchaseOrder
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetMaxOrderNoByDate 获取指定日期的最大订单号
func (r *purchaseOrderRepository) GetMaxOrderNoByDate(ctx context.Context, dateStr string) (string, error) {
	var maxOrderNo string
	prefix := "PO" + dateStr

	err := r.db.WithContext(ctx).
		Model(&models.PurchaseOrder{}).
		Where("order_no LIKE ?", prefix+"%").
		Order("order_no DESC").
		Limit(1).
		Select("order_no").
		First(&maxOrderNo).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}

	return maxOrderNo, nil
}

// Create 创建采购订单
func (r *purchaseOrderRepository) Create(ctx context.Context, order *models.PurchaseOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// CreateBatch 批量创建采购订单
func (r *purchaseOrderRepository) CreateBatch(ctx context.Context, orders []*models.PurchaseOrder) error {
	if len(orders) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(orders, 100).Error
}

// Update 更新采购订单
func (r *purchaseOrderRepository) Update(ctx context.Context, order *models.PurchaseOrder) error {
	return r.db.WithContext(ctx).
		Model(&models.PurchaseOrder{}).
		Where("id = ?", order.ID).
		Updates(order).Error
}

// Delete 删除采购订单（软删除）
func (r *purchaseOrderRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.PurchaseOrder{}, id).Error
}

// applyFilters 应用筛选条件
func (r *purchaseOrderRepository) applyFilters(db *gorm.DB, query ListQuery) *gorm.DB {
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
	return db
}

package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	TablePrefix = ""
	TableName   = "purchase_orders"
)

// PurchaseOrder 采购订单模型
type PurchaseOrder struct {
	ID           uint64          `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	OrderNo      string          `json:"orderNo" gorm:"uniqueIndex:idx_order_no;size:50;not null;comment:订单号"`
	SupplierName string          `json:"supplierName" gorm:"index:idx_supplier_name;size:100;not null;comment:供应商名称"`
	OrderDate    time.Time       `json:"orderDate" gorm:"index:idx_order_date;type:date;not null;comment:订单日期"`
	TotalAmount  decimal.Decimal `json:"totalAmount" gorm:"type:decimal(12,2);not null;comment:订单总金额"`
	Status       int8            `json:"status" gorm:"index:idx_status;not null;default:1;comment:订单状态"`

	// 审计字段
	CreatedBy uint64          `json:"createdBy,omitempty" gorm:"comment:创建人ID"`
	UpdatedBy uint64          `json:"updatedBy,omitempty" gorm:"comment:更新人ID"`
	CreatedAt time.Time       `json:"createdAt" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time       `json:"updatedAt" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt gorm.DeletedAt  `json:"-" gorm:"index;comment:删除时间"`
}

// TableName 指定表名
func (PurchaseOrder) TableName() string {
	return TablePrefix + TableName
}

// 状态常量
const (
	StatusPending   int8 = 1 // 待审批
	StatusApproved  int8 = 2 // 已审批
	StatusCompleted int8 = 3 // 已完成
	StatusCancelled int8 = 4 // 已取消
	StatusSettled   int8 = 5 // 已结算
	StatusDelivered int8 = 6 // 已到货
)

// StatusMap 状态映射
var StatusMap = map[int8]string{
	StatusPending:   "待审批",
	StatusApproved:  "已审批",
	StatusCompleted: "已完成",
	StatusCancelled: "已取消",
	StatusSettled:   "已结算",
	StatusDelivered: "已到货",
}

// StatusText 获取状态描述
func (p *PurchaseOrder) StatusText() string {
	if text, ok := StatusMap[p.Status]; ok {
		return text
	}
	return "未知"
}

// 状态转换规则
var statusTransitions = map[int8][]int8{
	StatusPending:   {StatusApproved, StatusCancelled},
	StatusApproved:  {StatusCompleted, StatusCancelled},
	StatusCompleted: {StatusSettled},
	StatusSettled:   {StatusDelivered},
}

// CanTransitionTo 检查状态转换是否合法
func (p *PurchaseOrder) CanTransitionTo(newStatus int8) bool {
	allowed, ok := statusTransitions[p.Status]
	if !ok {
		return false
	}

	for _, s := range allowed {
		if s == newStatus {
			return true
		}
	}
	return false
}

// IsValidStatus 检查状态是否有效
func IsValidStatus(status int8) bool {
	_, ok := StatusMap[status]
	return ok
}

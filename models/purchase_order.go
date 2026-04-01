package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type PurchaseOrder struct {
	ID           uint64          `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderNo      string          `json:"orderNo" gorm:"uniqueIndex;size:50;not null"`
	SupplierName string          `json:"supplierName" gorm:"size:100;not null"`
	OrderDate    time.Time       `json:"orderDate" gorm:"type:date;not null"`
	TotalAmount  decimal.Decimal `json:"totalAmount" gorm:"type:decimal(12,2);not null"`
	Status       int8            `json:"status" gorm:"not null;default:1"`
	CreatedAt    time.Time       `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

const (
	StatusPending   int8 = 1 // 待审批
	StatusApproved  int8 = 2 // 已审批
	StatusCompleted int8 = 3 // 已完成
	StatusCancelled int8 = 4 // 已取消
	StatusSettled   int8 = 5 // 已结算
	StatusDelivered int8 = 6 // 已到货
)

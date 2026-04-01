package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"purchase-order/models"
	"purchase-order/repositories"
	"purchase-order/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type purchaseOrderServiceStub struct {
	lastQuery repositories.ListQuery
}

func (s *purchaseOrderServiceStub) GetList(query repositories.ListQuery) (*services.ListResult, error) {
	s.lastQuery = query
	return &services.ListResult{
		List:     []models.PurchaseOrder{},
		Total:    0,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (s *purchaseOrderServiceStub) GetByID(id uint64) (*models.PurchaseOrder, error) {
	return &models.PurchaseOrder{ID: id}, nil
}

func (s *purchaseOrderServiceStub) Create(order models.PurchaseOrder) (*models.PurchaseOrder, error) {
	return &order, nil
}

func TestGetListAllowsDeliveredStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceStub := &purchaseOrderServiceStub{}
	controller := NewPurchaseOrderController(serviceStub)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/purchase-orders?status=6", nil)
	ctx.Request = req

	controller.GetList(ctx)

	if assert.NotNil(t, serviceStub.lastQuery.Status) {
		assert.Equal(t, int8(6), *serviceStub.lastQuery.Status)
	}
	assert.Equal(t, http.StatusOK, w.Code)
}

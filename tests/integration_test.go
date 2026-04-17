package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"purchase-order/config"
	"purchase-order/routers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	// 确保数据库已初始化
	if config.GetDB() == nil {
		cfg := config.DatabaseConfig{
			Host:         "localhost",
			Port:         3306,
			User:         "root",
			Password:     "123456",
			DBName:       "purchase_order_db",
			LogMode:      "silent",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
		}
		config.InitDatabase(cfg)
	}

	gin.SetMode(gin.TestMode)
	return routers.SetupRouter()
}

func TestGetList(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/purchase-orders", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetListWithParams(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/purchase-orders?page=1&page_size=10&order_by=id&order=desc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetByID(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/purchase-orders/1", nil)
	router.ServeHTTP(w, req)

	// 返回 404 因为数据库中没有数据
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPing(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

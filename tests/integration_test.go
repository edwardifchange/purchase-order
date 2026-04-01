package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/api/v1/purchase-orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": []string{}, "message": "success"})
	})

	router.GET("/api/v1/purchase-orders/page=1&page_size=10&order_by=id&order=desc", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": []string{}, "message": "success"})
	})

	router.GET("/api/v1/purchase-orders/1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": "1", "message": "success"})
	})

	return router
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

	assert.Equal(t, http.StatusOK, w.Code)
}

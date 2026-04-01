package routers

import (
	"net/http"
	"purchase-order/config"
	"purchase-order/controllers"
	"purchase-order/repositories"
	"purchase-order/services"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 初始化依赖
	db := config.GetDB()
	purchaseOrderRepo := repositories.NewPurchaseOrderRepository(db)
	purchaseOrderService := services.NewPurchaseOrderService(purchaseOrderRepo)
	purchaseOrderController := controllers.NewPurchaseOrderController(purchaseOrderService)

	// 注册路由
	api := r.Group("/api/v1")
	{
		purchaseOrders := api.Group("/purchase-orders")
		{
			purchaseOrders.GET("", purchaseOrderController.GetList)
			purchaseOrders.GET("/:poId", purchaseOrderController.GetByID)
			purchaseOrders.POST("", purchaseOrderController.Create)
		}

		// 测试接口
		api.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "ok",
				"data":    []interface{}{},
			})
		})

		api.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "ok",
				"data":    []interface{}{},
			})
		})
	}

	return r
}

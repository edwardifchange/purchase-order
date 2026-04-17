package routers

import (
	"net/http"
	"purchase-order/config"
	"purchase-order/controllers"
	"purchase-order/repositories"
	"purchase-order/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// requestLogger 请求日志中间件
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// 可以在这里记录日志
		_ = path
		_ = latency
		_ = status
	}
}

// errorRecovery 错误恢复中间件
func errorRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "内部服务错误",
					"data":    nil,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// requestTracker 请求追踪中间件
func requestTracker() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := uuid.New().String()
		c.Set("traceID", traceID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

// healthCheck 健康检查
func healthCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database ping failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	}
}

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// 全局中间件
	r.Use(errorRecovery())
	r.Use(requestLogger())
	r.Use(requestTracker())

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Trace-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Trace-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 初始化依赖
	db := config.GetDB()
	purchaseOrderRepo := repositories.NewPurchaseOrderRepository(db)
	purchaseOrderService := services.NewPurchaseOrderService(purchaseOrderRepo)
	purchaseOrderController := controllers.NewPurchaseOrderController(purchaseOrderService)

	// 健康检查端点
	r.GET("/health", healthCheck(db))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 采购订单路由
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

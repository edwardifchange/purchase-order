package main

import (
	"log"

	"purchase-order/config"
	"purchase-order/routers"
)

func main() {
	// 初始化数据库
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "123456",
		DBName:   "purchase_order_db",
	}

	if err := config.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 设置路由
	r := routers.SetupRouter()

	// 启动服务
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

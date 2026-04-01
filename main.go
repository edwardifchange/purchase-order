package main

import (
	"log"
	"os"

	"purchase-order/config"
	"purchase-order/routers"
)

func main() {
	// 初始化数据库
	cfg := config.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     3306,
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "purchase_order_db"),
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

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

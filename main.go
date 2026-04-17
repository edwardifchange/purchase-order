package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"purchase-order/config"
	"purchase-order/routers"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// 初始化数据库
	cfg := config.DatabaseConfig{
		Host:         getEnv("DB_HOST", "localhost"),
		Port:         3306,
		User:         getEnv("DB_USER", "root"),
		Password:     getEnv("DB_PASSWORD", ""),
		DBName:       getEnv("DB_NAME", "purchase_order_db"),
		LogMode:      getEnv("DB_LOG_MODE", "info"),
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		MaxLifetime:  time.Hour,
	}

	if err := config.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 设置路由
	r := routers.SetupRouter()

	// 启动服务
	port := getEnv("SERVER_PORT", "8080")
	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
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

package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"purchase-order/models"
)

var db *gorm.DB

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	LogMode      string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
}

// InitDatabase 初始化数据库连接
func InitDatabase(cfg DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	gormConfig := &gorm.Config{
		Logger: getLogLevel(cfg.LogMode),
	}

	var err error
	db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

	// 自动迁移
	err = db.AutoMigrate(&models.PurchaseOrder{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// getLogLevel 获取日志级别
func getLogLevel(mode string) logger.Interface {
	switch mode {
	case "silent":
		return logger.Default.LogMode(logger.Silent)
	case "error":
		return logger.Default.LogMode(logger.Error)
	case "warn":
		return logger.Default.LogMode(logger.Warn)
	case "info":
		return logger.Default.LogMode(logger.Info)
	default:
		return logger.Default.LogMode(logger.Info)
	}
}

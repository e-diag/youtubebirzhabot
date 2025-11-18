package db

import (
	"fmt"
	"os"
	"time"
	"youtube-market/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	config := &gorm.Config{}
	if os.Getenv("GIN_MODE") != "release" {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	// Настройка подключения с retry логикой
	var db *gorm.DB
	var err error
	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), config)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			fmt.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...\n", i+1, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Настройка connection pool для стабильности
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Настройки connection pool
	sqlDB.SetMaxIdleConns(10)                  // Максимум простаивающих соединений
	sqlDB.SetMaxOpenConns(100)                 // Максимум открытых соединений
	sqlDB.SetConnMaxLifetime(time.Hour)        // Максимальное время жизни соединения
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Максимальное время простоя соединения

	// Проверяем подключение
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Auto migrate models
	if err := db.AutoMigrate(&models.User{}, &models.Ad{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	DB = db
	return nil
}

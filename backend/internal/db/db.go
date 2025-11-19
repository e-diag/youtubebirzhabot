package db

import (
	"context"
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

	// Логируем DSN для отладки (без пароля)
	fmt.Printf("Connecting to database (DSN: %s)\n", maskDSN(dsn))

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
			fmt.Printf("Successfully connected to database on attempt %d\n", i+1)
			break
		}
		if i < maxRetries-1 {
			fmt.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...\n", i+1, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		} else {
			fmt.Printf("Failed to connect to database after %d attempts: %v\n", maxRetries, err)
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
	sqlDB.SetMaxIdleConns(5)                   // Максимум простаивающих соединений (уменьшено)
	sqlDB.SetMaxOpenConns(25)                 // Максимум открытых соединений (уменьшено)
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // Максимальное время жизни соединения (уменьшено)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)  // Максимальное время простоя соединения (уменьшено)

	// Проверяем подключение с таймаутом
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	
	fmt.Println("Database ping successful")

	// Auto migrate models
	if err := db.AutoMigrate(&models.User{}, &models.Ad{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	DB = db
	fmt.Println("Database initialized successfully")
	return nil
}

// maskDSN скрывает пароль в DSN для безопасного логирования
func maskDSN(dsn string) string {
	// Простая маскировка пароля в connection string
	// Формат: postgres://user:password@host:port/dbname
	if len(dsn) < 20 {
		return "***"
	}
	// Ищем позицию пароля (между : и @)
	start := 0
	end := len(dsn)
	for i := 0; i < len(dsn)-1; i++ {
		if dsn[i] == ':' && dsn[i+1] != '/' {
			start = i + 1
			break
		}
	}
	for i := start; i < len(dsn); i++ {
		if dsn[i] == '@' {
			end = i
			break
		}
	}
	if start > 0 && end > start {
		return dsn[:start] + "***" + dsn[end:]
	}
	return "***"
}

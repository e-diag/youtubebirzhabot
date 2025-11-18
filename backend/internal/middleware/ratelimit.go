package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"youtube-market/internal/metrics"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

// InitRedis инициализирует подключение к Redis
func InitRedis() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		// Используем дефолтные настройки для docker-compose
		rdb = redis.NewClient(&redis.Options{
			Addr:     "redis:6379",
			Password: "",
			DB:       0,
		})
	} else {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			return fmt.Errorf("failed to parse REDIS_URL: %w", err)
		}
		rdb = redis.NewClient(opt)
	}

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// RateLimitMiddleware ограничивает количество запросов: 60 запросов в минуту на IP
// Исключает статические ресурсы и метрики из лимита
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Исключаем статические ресурсы, метрики и health check из rate limiting
		path := c.Request.URL.Path
		if path == "/metrics" || path == "/health" || 
		   path == "/launch_512x512.svg" ||
		   path == "/terms" || path == "/privacy" ||
		   path == "/" || 
		   len(path) > 7 && path[:7] == "/static" ||
		   len(path) > 7 && path[:7] == "/assets" {
			c.Next()
			return
		}

		if rdb == nil {
			// Если Redis не инициализирован, пропускаем проверку
			c.Next()
			return
		}

		// Получаем IP адрес
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s", ip)

		ctx := context.Background()

		// Получаем текущее количество запросов
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Если ошибка Redis, пропускаем проверку
			c.Next()
			return
		}

		// Устанавливаем TTL для ключа (1 минута)
		if count == 1 {
			rdb.Expire(ctx, key, time.Minute)
		}

		// Проверяем лимит (60 запросов в минуту вместо 10)
		if count > 60 {
			// Собираем метрику rate limit
			metrics.RateLimitHits.WithLabelValues(ip).Inc()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

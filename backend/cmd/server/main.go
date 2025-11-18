package main

import (
	"log"
	"os"
	"time"
	"youtube-market/internal/db"
	"youtube-market/internal/handlers"
	"youtube-market/internal/logger"
	"youtube-market/internal/metrics"
	"youtube-market/internal/middleware"
	"youtube-market/internal/models"
	"youtube-market/internal/notifier"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// В контейнере .env может отсутствовать — переменные окружения передаются напрямую.
		if !os.IsNotExist(err) {
			log.Printf("Warning: could not load .env file: %v", err)
		}
	}

	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Printf("Warning: Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Initialize Telegram notifications
	if err := notifier.Init(); err != nil {
		log.Printf("Warning: Failed to initialize Telegram notifications: %v", err)
	}

	// Initialize database
	if err := db.Init(); err != nil {
		logger.Fatal("Failed to initialize database", err, nil)
	}

	// Initialize Redis for rate limiting
	if err := middleware.InitRedis(); err != nil {
		logger.Warning("Redis not available, rate limiting disabled", map[string]interface{}{
			"error": err.Error(),
		})
		notifier.NotifyWarning("Redis not available, rate limiting disabled", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Setup router
	r := setupRouter()

	// Start manager bot in background
	go handlers.RunManagerBot()

	// Start metrics collection in background
	go collectMetrics()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting", map[string]interface{}{
		"port": port,
	})
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", err, nil)
	}
}

func setupRouter() *gin.Engine {
	// Set release mode in production
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Global middleware
	r.Use(middleware.SafeLoggerMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware())
	r.Use(middleware.ErrorLoggerMiddleware())

	// Static files
	r.Static("/static", "./static")
	r.Static("/assets", "./static/assets")

	// Главная страница (поддержка GET и HEAD)
	serveIndex := func(c *gin.Context) {
		// Проверяем существование файла перед отправкой
		if _, err := os.Stat("./static/index.html"); os.IsNotExist(err) {
			log.Printf("Warning: static/index.html not found, serving 404")
			c.JSON(404, gin.H{"error": "index.html not found"})
			return
		}
		c.File("./static/index.html")
	}
	r.GET("/", serveIndex)
	r.HEAD("/", serveIndex)

	// Legal pages (поддержка GET и HEAD)
	serveTerms := func(c *gin.Context) {
		if _, err := os.Stat("./static/terms.html"); os.IsNotExist(err) {
			c.JSON(404, gin.H{"error": "terms.html not found"})
			return
		}
		c.File("./static/terms.html")
	}
	servePrivacy := func(c *gin.Context) {
		if _, err := os.Stat("./static/privacy.html"); os.IsNotExist(err) {
			c.JSON(404, gin.H{"error": "privacy.html not found"})
			return
		}
		c.File("./static/privacy.html")
	}
	r.GET("/terms", serveTerms)
	r.HEAD("/terms", serveTerms)
	r.GET("/privacy", servePrivacy)
	r.HEAD("/privacy", servePrivacy)

	// Metrics endpoint (без аутентификации для мониторинга)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes with TMA authentication
	api := r.Group("/api")
	api.Use(middleware.TMAuthMiddleware())
	{
		api.GET("/ads", handlers.GetAds)
		api.GET("/ads/:id/photo", handlers.GetAdPhoto)
		api.GET("/myads", handlers.GetMyAds)
		api.GET("/profile/:username", handlers.GetProfileAds)
		api.GET("/scammer/:username", handlers.CheckScammer)
		api.GET("/blacklist", handlers.GetBlacklist)
	}

	return r
}

// collectMetrics периодически обновляет бизнес-метрики
func collectMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		updateBusinessMetrics()
	}
}

// updateBusinessMetrics обновляет метрики из базы данных
func updateBusinessMetrics() {
	// Общее количество объявлений
	var totalAds int64
	db.DB.Model(&models.Ad{}).Count(&totalAds)
	metrics.AdsTotal.Set(float64(totalAds))

	// Активные объявления
	var activeAds int64
	now := time.Now()
	db.DB.Model(&models.Ad{}).Where("status = ? AND expires_at > ?", models.AdStatusActive, now).Count(&activeAds)
	metrics.AdsActive.Set(float64(activeAds))

	// Премиум объявления
	var premiumAds int64
	db.DB.Model(&models.Ad{}).Where("status = ? AND expires_at > ? AND is_premium = ?", models.AdStatusActive, now, true).Count(&premiumAds)
	metrics.AdsPremium.Set(float64(premiumAds))

	// Общее количество пользователей
	var totalUsers int64
	db.DB.Model(&models.User{}).Count(&totalUsers)
	metrics.UsersTotal.Set(float64(totalUsers))

	// Пользователи в чёрном списке
	var scammers int64
	db.DB.Model(&models.User{}).Where("is_scammer = ?", true).Count(&scammers)
	metrics.UsersScammers.Set(float64(scammers))
}

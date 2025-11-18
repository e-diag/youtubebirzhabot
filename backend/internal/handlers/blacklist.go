package handlers

import (
	"net/http"
	"strings"
	"time"
	"youtube-market/internal/db"
	"youtube-market/internal/metrics"
	"youtube-market/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CheckScammer(c *gin.Context) {
	start := time.Now()
	username := strings.TrimSpace(strings.TrimPrefix(c.Param("username"), "@"))
	if username == "" {
		metrics.APIRequestsTotal.WithLabelValues("scammer", "400").Inc()
		metrics.ErrorsTotal.WithLabelValues("validation", "scammer").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "username parameter is required"})
		return
	}

	queryStart := time.Now()
	var user models.User
	err := db.DB.Where("LOWER(username) = LOWER(?) AND is_scammer = true", username).First(&user).Error
	metrics.DatabaseQueryDuration.WithLabelValues("select").Observe(time.Since(queryStart).Seconds())

	if err == gorm.ErrRecordNotFound {
		metrics.APIRequestsTotal.WithLabelValues("scammer", "200").Inc()
		metrics.APIReponseTime.WithLabelValues("scammer").Observe(time.Since(start).Seconds())
		c.JSON(http.StatusOK, gin.H{
			"safe": true,
			"msg":  "Юзер не был замечен в мошеннических схемах",
		})
		return
	}

	if err != nil {
		metrics.APIRequestsTotal.WithLabelValues("scammer", "500").Inc()
		metrics.ErrorsTotal.WithLabelValues("database", "scammer").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check user"})
		return
	}

	metrics.APIRequestsTotal.WithLabelValues("scammer", "200").Inc()
	metrics.APIReponseTime.WithLabelValues("scammer").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{
		"safe": false,
		"msg":  "Осторожно! Мошенник",
	})
}

func GetBlacklist(c *gin.Context) {
	start := time.Now()
	queryStart := time.Now()
	var scammers []models.User
	if err := db.DB.
		Where("is_scammer = ?", true).
		Order("username ASC").
		Find(&scammers).Error; err != nil {
		metrics.APIRequestsTotal.WithLabelValues("blacklist", "500").Inc()
		metrics.ErrorsTotal.WithLabelValues("database", "blacklist").Inc()
		metrics.DatabaseQueryDuration.WithLabelValues("select").Observe(time.Since(queryStart).Seconds())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load blacklist"})
		return
	}
	metrics.DatabaseQueryDuration.WithLabelValues("select").Observe(time.Since(queryStart).Seconds())

	response := make([]gin.H, 0, len(scammers))
	for _, user := range scammers {
		response = append(response, gin.H{
			"username":   user.Username,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	// Собираем метрики
	metrics.APIRequestsTotal.WithLabelValues("blacklist", "200").Inc()
	metrics.APIReponseTime.WithLabelValues("blacklist").Observe(time.Since(start).Seconds())

	c.JSON(http.StatusOK, response)
}

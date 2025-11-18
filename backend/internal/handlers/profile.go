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

func GetProfileAds(c *gin.Context) {
	start := time.Now()
	username := strings.TrimSpace(c.Param("username"))
	if username == "" {
		metrics.APIRequestsTotal.WithLabelValues("profile", "400").Inc()
		metrics.ErrorsTotal.WithLabelValues("validation", "profile").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "username parameter is required"})
		return
	}

	username = strings.TrimPrefix(username, "@")

	queryStart := time.Now()
	var ads []models.Ad
	if err := db.DB.
		Where("LOWER(username) = LOWER(?)", username).
		Order(gorm.Expr("CASE WHEN status = ? THEN 0 WHEN status = ? THEN 1 ELSE 2 END, updated_at DESC", models.AdStatusActive, models.AdStatusExpired)).
		Find(&ads).Error; err != nil {
		metrics.APIRequestsTotal.WithLabelValues("profile", "500").Inc()
		metrics.ErrorsTotal.WithLabelValues("database", "profile").Inc()
		metrics.DatabaseQueryDuration.WithLabelValues("select").Observe(time.Since(queryStart).Seconds())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch profile ads"})
		return
	}
	metrics.DatabaseQueryDuration.WithLabelValues("select").Observe(time.Since(queryStart).Seconds())

	now := time.Now()
	response := make([]AdView, 0, len(ads))
	for _, ad := range ads {
		// Ensure status reflects current expiration
		if ad.Status == models.AdStatusActive && ad.ExpiresAt.Before(now) {
			ad.Status = models.AdStatusExpired
		}
		response = append(response, buildAdView(ad))
	}

	// Собираем метрики
	metrics.APIRequestsTotal.WithLabelValues("profile", "200").Inc()
	metrics.APIReponseTime.WithLabelValues("profile").Observe(time.Since(start).Seconds())

	c.JSON(http.StatusOK, response)
}

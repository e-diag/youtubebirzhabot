package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"youtube-market/internal/db"
	"youtube-market/internal/metrics"
	"youtube-market/internal/middleware"
	"youtube-market/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAdPhoto(c *gin.Context) {
	start := time.Now()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		metrics.APIRequestsTotal.WithLabelValues("ad_photo", "400").Inc()
		metrics.ErrorsTotal.WithLabelValues("validation", "ad_photo").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ad id"})
		return
	}

	queryStart := time.Now()
	var ad models.Ad
	if err := db.DB.First(&ad, id).Error; err != nil {
		metrics.APIRequestsTotal.WithLabelValues("ad_photo", "404").Inc()
		metrics.DatabaseQueryDuration.WithLabelValues("select").Observe(time.Since(queryStart).Seconds())
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "ad not found"})
		} else {
			middleware.CaptureError(c, err, map[string]string{
				"handler": "GetAdPhoto",
				"ad_id":   strconv.Itoa(id),
			})
			metrics.ErrorsTotal.WithLabelValues("database", "ad_photo").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ad"})
		}
		return
	}
	metrics.DatabaseQueryDuration.WithLabelValues("select").Observe(time.Since(queryStart).Seconds())

	if ad.PhotoPath == "" {
		metrics.APIRequestsTotal.WithLabelValues("ad_photo", "404").Inc()
		c.Status(http.StatusNotFound)
		return
	}

	token := getBotToken()
	if token == "" {
		metrics.APIRequestsTotal.WithLabelValues("ad_photo", "500").Inc()
		metrics.ErrorsTotal.WithLabelValues("config", "ad_photo").Inc()
		c.Status(http.StatusNotFound)
		return
	}

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, ad.PhotoPath)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		middleware.CaptureError(c, err, map[string]string{
			"handler":   "GetAdPhoto",
			"ad_id":     strconv.Itoa(id),
			"error_type": "external_api",
		})
		metrics.APIRequestsTotal.WithLabelValues("ad_photo", "502").Inc()
		metrics.ErrorsTotal.WithLabelValues("external_api", "ad_photo").Inc()
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch photo"})
		return
	}
	defer resp.Body.Close()

	for k, values := range resp.Header {
		if len(values) == 0 {
			continue
		}
		switch k {
		case "Content-Type", "Content-Length":
			c.Header(k, values[0])
		}
	}

	metrics.APIRequestsTotal.WithLabelValues("ad_photo", "200").Inc()
	metrics.APIReponseTime.WithLabelValues("ad_photo").Observe(time.Since(start).Seconds())
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, resp.Body)
}


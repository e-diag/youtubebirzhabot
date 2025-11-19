package handlers

import (
	"fmt"
	"io"
	"log"
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
		log.Printf("GetAdPhoto: invalid ad id: %v", err)
		metrics.APIRequestsTotal.WithLabelValues("ad_photo", "400").Inc()
		metrics.ErrorsTotal.WithLabelValues("validation", "ad_photo").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ad id"})
		return
	}
	
	log.Printf("GetAdPhoto: запрос фото для объявления ID=%d", id)

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
		log.Printf("GetAdPhoto: объявление ID=%d не имеет фото", id)
		metrics.APIRequestsTotal.WithLabelValues("ad_photo", "404").Inc()
		c.Status(http.StatusNotFound)
		return
	}

	token := getBotToken()
	if token == "" {
		log.Printf("GetAdPhoto: BOT_TOKEN не установлен")
		metrics.APIRequestsTotal.WithLabelValues("ad_photo", "500").Inc()
		metrics.ErrorsTotal.WithLabelValues("config", "ad_photo").Inc()
		c.Status(http.StatusNotFound)
		return
	}

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, ad.PhotoPath)
	log.Printf("GetAdPhoto: запрос фото из Telegram API: %s", url)
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("GetAdPhoto: ошибка при запросе к Telegram API: %v", err)
		middleware.CaptureError(c, err, map[string]string{
			"handler":    "GetAdPhoto",
			"ad_id":      strconv.Itoa(id),
			"error_type": "external_api",
			"photo_path": ad.PhotoPath,
		})
		metrics.APIRequestsTotal.WithLabelValues("ad_photo", "502").Inc()
		metrics.ErrorsTotal.WithLabelValues("external_api", "ad_photo").Inc()
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch photo"})
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("GetAdPhoto: Telegram API вернул статус %d для объявления ID=%d", resp.StatusCode, id)
		middleware.CaptureError(c, fmt.Errorf("telegram API returned status %d", resp.StatusCode), map[string]string{
			"handler":    "GetAdPhoto",
			"ad_id":      strconv.Itoa(id),
			"error_type": "external_api",
			"status_code": strconv.Itoa(resp.StatusCode),
			"photo_path": ad.PhotoPath,
		})
		metrics.APIRequestsTotal.WithLabelValues("ad_photo", "502").Inc()
		metrics.ErrorsTotal.WithLabelValues("external_api", "ad_photo").Inc()
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch photo"})
		return
	}

	// Устанавливаем заголовки перед копированием тела
	for k, values := range resp.Header {
		if len(values) == 0 {
			continue
		}
		switch k {
		case "Content-Type", "Content-Length":
			c.Header(k, values[0])
		}
	}

	// Устанавливаем статус и копируем тело ответа
	c.Status(http.StatusOK)
	
	// Копируем тело ответа
	bytesCopied, err := io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("GetAdPhoto: ошибка при копировании тела ответа: %v", err)
		// Не возвращаем ошибку клиенту, так как заголовки уже отправлены
		return
	}
	
	log.Printf("GetAdPhoto: успешно отправлено %d байт для объявления ID=%d", bytesCopied, id)
	metrics.APIRequestsTotal.WithLabelValues("ad_photo", "200").Inc()
	metrics.APIReponseTime.WithLabelValues("ad_photo").Observe(time.Since(start).Seconds())
}


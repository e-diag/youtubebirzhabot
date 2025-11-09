package handlers

import (
	"net/http"
	"strings"
	"youtube-market/internal/db"
	"youtube-market/internal/models"

	"github.com/gin-gonic/gin"
)

func GetProfileAds(c *gin.Context) {
	username := strings.TrimSpace(c.Param("username"))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username parameter is required"})
		return
	}

	// Remove @ if present
	username = strings.TrimPrefix(username, "@")

	var ads []models.Ad
	if err := db.DB.Where("LOWER(username) = LOWER(?) AND expires_at > NOW() AND deleted_at IS NULL", username).
		Order("is_premium DESC, created_at DESC").
		Find(&ads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile ads"})
		return
	}

	c.JSON(http.StatusOK, ads)
}

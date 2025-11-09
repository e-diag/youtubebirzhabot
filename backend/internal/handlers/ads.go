package handlers

import (
	"net/http"
	"strconv"
	"youtube-market/internal/db"
	"youtube-market/internal/models"

	"github.com/gin-gonic/gin"
)

func GetAds(c *gin.Context) {
	var ads []models.Ad
	query := db.DB.Where("expires_at > NOW() AND deleted_at IS NULL")

	// Filter by category
	if cat := c.Query("cat"); cat != "" {
		query = query.Where("category = ?", cat)
	}

	// Filter by filter1
	if f1 := c.Query("f1"); f1 != "" {
		query = query.Where("filter1 = ?", f1)
	}

	// Order by premium first, then by creation date
	query = query.Order("is_premium DESC, created_at DESC")

	// Execute query
	if err := query.Find(&ads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ads"})
		return
	}

	c.JSON(http.StatusOK, ads)
}

func GetMyAds(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id parameter is required"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	var ads []models.Ad
	if err := db.DB.Where("user_id = ? AND expires_at > NOW() AND deleted_at IS NULL", userID).Find(&ads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ads"})
		return
	}

	c.JSON(http.StatusOK, ads)
}

package handlers

import (
	"net/http"
	"strings"
	"youtube-market/internal/db"
	"youtube-market/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CheckScammer(c *gin.Context) {
	username := strings.TrimSpace(c.Param("username"))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username parameter is required"})
		return
	}

	// Remove @ if present
	username = strings.TrimPrefix(username, "@")

	var user models.User
	err := db.DB.Where("LOWER(username) = LOWER(?) AND is_scammer = true", username).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, gin.H{
			"safe": true,
			"msg":  "Юзер не был замечен в мошеннических схемах",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"safe": false,
		"msg":  "Осторожно! Мошенник",
	})
}

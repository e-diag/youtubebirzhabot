package middleware

import (
	"fmt"
	"youtube-market/internal/logger"
	"youtube-market/internal/metrics"
	"youtube-market/internal/notifier"

	"github.com/gin-gonic/gin"
)

// ErrorLoggerMiddleware логирует ошибки в файлы и отправляет уведомления
func ErrorLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Логируем только ошибки (статус >= 500)
		if c.Writer.Status() >= 500 {
			userID := ""
			if uid, exists := c.Get("user_id"); exists {
				userID = uid.(string)
			}

			username := ""
			if un, exists := c.Get("username"); exists {
				username = un.(string)
			}

			context := map[string]interface{}{
				"status_code": c.Writer.Status(),
				"username":    username,
			}

			err := fmt.Errorf("HTTP %d: %s %s", c.Writer.Status(), c.Request.Method, c.Request.URL.Path)

			logger.ErrorWithContext(
				"HTTP error",
				err,
				userID,
				c.Request.URL.Path,
				c.Request.Method,
				c.ClientIP(),
				context,
			)

			// Отправляем уведомление в Telegram для критических ошибок
			if c.Writer.Status() >= 500 {
				notifier.NotifyError(
					fmt.Sprintf("HTTP %d: %s %s", c.Writer.Status(), c.Request.Method, c.Request.URL.Path),
					err,
					context,
				)
			}

			// Увеличиваем метрику
			metrics.ErrorsTotal.WithLabelValues("http", c.Request.URL.Path).Inc()
		}
	}
}

// CaptureError логирует ошибку с контекстом
func CaptureError(c *gin.Context, err error, tags map[string]string) {
	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	username := ""
	if un, exists := c.Get("username"); exists {
		username = un.(string)
	}

	context := make(map[string]interface{})
	if tags != nil {
		for k, v := range tags {
			context[k] = v
		}
	}
	context["username"] = username

	logger.ErrorWithContext(
		"Application error",
		err,
		userID,
		c.Request.URL.Path,
		c.Request.Method,
		c.ClientIP(),
		context,
	)

	// Отправляем уведомление в Telegram
	notifier.NotifyError("Application error", err, context)

	// Увеличиваем метрику
	metrics.ErrorsTotal.WithLabelValues("application", c.Request.URL.Path).Inc()
}

// CaptureMessage логирует сообщение
func CaptureMessage(c *gin.Context, message string, level logger.LogLevel) {
	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	context := map[string]interface{}{
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
		"ip":     c.ClientIP(),
	}

	switch level {
	case logger.LevelError:
		logger.Error(message, nil, context)
		notifier.NotifyError(message, nil, context)
	case logger.LevelWarning:
		logger.Warning(message, context)
		notifier.NotifyWarning(message, context)
	default:
		logger.Info(message, context)
	}
}


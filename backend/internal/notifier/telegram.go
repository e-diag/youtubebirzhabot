package notifier

import (
	"fmt"
	"os"
	"time"
	"youtube-market/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot        *tgbotapi.BotAPI
	chatID     int64
	initialized bool
)

// Init инициализирует Telegram уведомления
func Init() error {
	botToken := os.Getenv("BOT_TOKEN")
	notifyChatID := os.Getenv("NOTIFY_CHAT_ID")

	if botToken == "" || notifyChatID == "" {
		// Уведомления не обязательны
		return nil
	}

	var err error
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("failed to initialize telegram bot for notifications: %w", err)
	}

	// Парсим chat ID
	if _, err := fmt.Sscanf(notifyChatID, "%d", &chatID); err != nil {
		return fmt.Errorf("invalid NOTIFY_CHAT_ID format: %w", err)
	}

	initialized = true
	logger.Info("Telegram notifications initialized", map[string]interface{}{
		"chat_id": chatID,
	})

	return nil
}

// NotifyError отправляет уведомление об ошибке в Telegram
func NotifyError(message string, err error, context map[string]interface{}) {
	if !initialized {
		return
	}

	text := fmt.Sprintf("🚨 *Ошибка в приложении*\n\n")
	text += fmt.Sprintf("*Сообщение:* %s\n", message)

	if err != nil {
		text += fmt.Sprintf("*Ошибка:* `%s`\n", err.Error())
	}

	if context != nil {
		text += "\n*Контекст:*\n"
		for k, v := range context {
			text += fmt.Sprintf("• %s: `%v`\n", k, v)
		}
	}

	text += fmt.Sprintf("\n*Время:* %s", time.Now().Format("2006-01-02 15:04:05"))

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	if _, sendErr := bot.Send(msg); sendErr != nil {
		logger.Error("Failed to send telegram notification", sendErr, nil)
	}
}

// NotifyWarning отправляет предупреждение в Telegram
func NotifyWarning(message string, context map[string]interface{}) {
	if !initialized {
		return
	}

	text := fmt.Sprintf("⚠️ *Предупреждение*\n\n")
	text += fmt.Sprintf("*Сообщение:* %s\n", message)

	if context != nil {
		text += "\n*Контекст:*\n"
		for k, v := range context {
			text += fmt.Sprintf("• %s: `%v`\n", k, v)
		}
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	if _, err := bot.Send(msg); err != nil {
		logger.Error("Failed to send telegram notification", err, nil)
	}
}

// NotifyInfo отправляет информационное сообщение в Telegram
func NotifyInfo(message string, context map[string]interface{}) {
	if !initialized {
		return
	}

	text := fmt.Sprintf("ℹ️ *Информация*\n\n")
	text += fmt.Sprintf("*Сообщение:* %s\n", message)

	if context != nil {
		text += "\n*Контекст:*\n"
		for k, v := range context {
			text += fmt.Sprintf("• %s: `%v`\n", k, v)
		}
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	if _, err := bot.Send(msg); err != nil {
		logger.Error("Failed to send telegram notification", err, nil)
	}
}


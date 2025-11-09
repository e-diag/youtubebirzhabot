package handlers

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"youtube-market/internal/db"
	"youtube-market/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// mustInt64 — безопасное преобразование строки в int64
func mustInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Printf("Invalid MANAGER_ID: %s, using 0", s)
		return 0
	}
	return i
}

func RunManagerBot() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Println("BOT_TOKEN not set, manager bot disabled")
		return
	}

	managerID := mustInt64(os.Getenv("MANAGER_ID"))

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Bot init failed:", err)
	}

	log.Printf("Manager bot started for user ID: %d", managerID)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		msg := update.Message

		// Только менеджер
		if msg.From.ID != managerID {
			continue
		}

		// Автоудаление сообщения
		go func(m tgbotapi.Message) {
			time.Sleep(30 * time.Second)
			bot.Request(tgbotapi.NewDeleteMessage(m.Chat.ID, m.MessageID))
		}(*msg)

		// === Команды ===
		text := msg.Text

		// Добавить в чёрный список
		if strings.HasPrefix(text, "/addscam") {
			username := strings.TrimSpace(strings.TrimPrefix(text, "/addscam"))
			username = strings.TrimPrefix(username, "@")
			username = strings.TrimSpace(username)
			if username == "" {
				sendReply(bot, msg.Chat.ID, msg.MessageID, "Использование: /addscam @username")
				continue
			}
			db.DB.FirstOrCreate(&models.User{}, models.User{Username: username}).Updates(map[string]interface{}{
				"IsScammer": true,
			})
			sendReply(bot, msg.Chat.ID, msg.MessageID, "✅ Добавлен в чёрный список: @"+username)
			continue
		}

		// Убрать из чёрного списка
		if strings.HasPrefix(text, "/remscam") {
			username := strings.TrimSpace(strings.TrimPrefix(text, "/remscam"))
			username = strings.TrimPrefix(username, "@")
			username = strings.TrimSpace(username)
			if username == "" {
				sendReply(bot, msg.Chat.ID, msg.MessageID, "Использование: /remscam @username")
				continue
			}
			result := db.DB.Where("username = ?", username).Updates(&models.User{IsScammer: false})
			if result.RowsAffected > 0 {
				sendReply(bot, msg.Chat.ID, msg.MessageID, "✅ Удалён из чёрного списка: @"+username)
			} else {
				sendReply(bot, msg.Chat.ID, msg.MessageID, "❌ Пользователь @"+username+" не найден в чёрном списке")
			}
			continue
		}

		// Показать меню
		if text == "/start" || text == "/menu" {
			showMenu(bot, msg.Chat.ID)
		}
	}
}

// Вспомогательная функция отправки ответа
func sendReply(bot *tgbotapi.BotAPI, chatID int64, replyTo int, text string) {
	reply := tgbotapi.NewMessage(chatID, text)
	reply.ReplyToMessageID = replyTo
	bot.Send(reply)
}

// Главное меню
func showMenu(bot *tgbotapi.BotAPI, chatID int64) {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Скамеры"),
			tgbotapi.NewKeyboardButton("Объявления"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Весь список"),
		),
	)
	msg := tgbotapi.NewMessage(chatID, "Меню менеджера:")
	msg.ReplyMarkup = kb
	bot.Send(msg)
}

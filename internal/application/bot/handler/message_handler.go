package handler

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	helpMessage = `
/track — начать отслеживание ссылки. Опционально можно указать теги.
/untrack — прекратить отслеживание ссылки.
/list — вывести список всех отслеживаемых ссылок.
`
	greetingMessage = "Добро пожаловать! Используйте /help, чтобы посмотреть доступные команды."
	unknownMessage  = "Неизвестная команда. Воспользуйтесь /help, чтобы посмотреть список доступных команд."
)

type MessageHandler interface {
	SendMessage(chatID int64, message string) error
}

type TelegramHandler struct {
	MsgHandler MessageHandler
}

func (h TelegramHandler) HandleCommand(update tgbotapi.Update) {
	var text string

	switch update.Message.Command() {
	case "start":
		text = greetingMessage
	case "help":
		text = helpMessage
	default:
		text = unknownMessage
	}

	if err := h.MsgHandler.SendMessage(update.Message.Chat.ID, text); err != nil {
		slog.Error("error while sending message", slog.String("error", err.Error()))
	}
}

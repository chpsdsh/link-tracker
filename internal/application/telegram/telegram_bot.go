package telegram

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	updateOffset  = 0
	updateTimeout = 60
)

type TelegramBot struct {
	Bot *tgbotapi.BotAPI
}

func (t TelegramBot) StartMainLoop() {
	u := tgbotapi.NewUpdate(updateOffset)
	u.Timeout = updateTimeout

	updates := t.Bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.IsCommand() {
			t.handleCommand(update)
		} else {
			t.handleMessage(update)
		}
	}
}

func (t TelegramBot) handleCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	switch update.Message.Command() {
	case "start":
		msg.Text = "Hello from LinkTracker"
	case "help":
		msg.Text = "Help will be updated soon"
	default:
		msg.Text = "I don't know that command"
	}

	if _, err := t.Bot.Send(msg); err != nil {
		slog.Error("message send error", "err", err.Error())
	}
}

func (t TelegramBot) handleMessage(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

	if _, err := t.Bot.Send(msg); err != nil {
		slog.Error("message send error", "err", err.Error())
	}
}

package bot

import (
	"context"
	"log/slog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
)

const (
	telegramWorkersNum = 5
	updateOffset       = 0
	updateTimeout      = 60
	startCommand       = "Старт"
	helpCommand        = "Помощь"
	trackCommand       = "начать отслеживание ссылки"
	untrackCommand     = "закончить отслеживание ссылки"
	listCommand        = "вывести список всех отслеживаемых ссылок"
)

type TelegramBot struct {
	Bot     *tgbotapi.BotAPI
	Handler handler.TelegramHandler
}

func (t TelegramBot) StartMainLoop(ctx context.Context, wg *sync.WaitGroup) {
	if err := t.setupBotCommands(); err != nil {
		slog.Error("error loading commands", slog.String("error", err.Error()))
	} else {
		slog.Info("all commands loaded")
	}

	u := tgbotapi.NewUpdate(updateOffset)
	u.Timeout = updateTimeout
	updates := t.Bot.GetUpdatesChan(u)

	for range telegramWorkersNum {
		wg.Go(func() {
			t.telegramWorker(ctx, updates)
		})
	}
}

func (t TelegramBot) telegramWorker(ctx context.Context, updateChan tgbotapi.UpdatesChannel) {
	for {
		select {
		case update, ok := <-updateChan:
			if !ok {
				return
			}
			t.Handler.HandleCommand(update)
		case <-ctx.Done():
			return
		}
	}
}

func (t TelegramBot) SendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	if _, err := t.Bot.Send(msg); err != nil {
		slog.Error("message send error", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (t TelegramBot) setupBotCommands() error {
	commands := []tgbotapi.BotCommand{
		{"start", startCommand},
		{"help", helpCommand},
		{"track", trackCommand},
		{"untrack", untrackCommand},
		{"list", listCommand},
	}
	conf := tgbotapi.NewSetMyCommands(commands...)
	if _, err := t.Bot.Request(conf); err != nil {
		return err
	}
	return nil
}

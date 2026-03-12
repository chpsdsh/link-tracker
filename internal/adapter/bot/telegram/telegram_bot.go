//go:generate mockgen -source telegram_bot.go -destination=../mocks/telegram_bot_mocks.go -package=mocks
package telegram

import (
	"context"
	"log/slog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	telegramWorkersNum = 5
	updateOffset       = 0
	updateTimeout      = 60
	startCommand       = "Старт"
	helpCommand        = "Помощь"
	trackCommand       = "Начать отслеживание ссылки"
	untrackCommand     = "Закончить отслеживание ссылки"
	listCommand        = "Вывести список всех отслеживаемых ссылок"
)

type BotHandler interface {
	HandleUpdate(update tgbotapi.Update)
}

type TelegramAPI interface {
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

type TelegramBot struct {
	Bot        TelegramAPI
	Handler    BotHandler
	BaseLogger *slog.Logger
}

func (t TelegramBot) StartMainLoop(ctx context.Context, wg *sync.WaitGroup) {
	if err := t.setupBotCommands(); err != nil {
		t.BaseLogger.Error("error loading commands", slog.String("error", err.Error()))
	} else {
		t.BaseLogger.Info("all commands loaded")
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
			t.Handler.HandleUpdate(update)
		case <-ctx.Done():
			return
		}
	}
}

func (t TelegramBot) SendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	if _, err := t.Bot.Send(msg); err != nil {
		t.BaseLogger.Error("message send error", slog.String("err", err.Error()))
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

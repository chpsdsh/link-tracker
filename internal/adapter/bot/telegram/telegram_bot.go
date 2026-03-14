//go:generate mockgen -source telegram_bot.go -destination=../mocks/telegram_bot_mocks.go -package=mocks
package telegram

import (
	"context"
	"fmt"
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

type TgAPI interface {
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

type Bot struct {
	BotAPI     TgAPI
	Handler    BotHandler
	BaseLogger *slog.Logger
}

func (t Bot) StartMainLoop(ctx context.Context, wg *sync.WaitGroup) {
	if err := t.setupBotCommands(); err != nil {
		t.BaseLogger.Error("error loading commands", slog.String("error", err.Error()))
	} else {
		t.BaseLogger.Info("all commands loaded")
	}

	u := tgbotapi.NewUpdate(updateOffset)
	u.Timeout = updateTimeout
	updates := t.BotAPI.GetUpdatesChan(u)

	for range telegramWorkersNum {
		wg.Go(func() {
			t.telegramWorker(ctx, updates)
		})
	}
}

func (t Bot) SendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	if _, err := t.BotAPI.Send(msg); err != nil {
		t.BaseLogger.Error("message send error", slog.String("err", err.Error()))
		return fmt.Errorf("error sending message: %w", err)
	}
	return nil
}

func (t Bot) telegramWorker(ctx context.Context, updateChan tgbotapi.UpdatesChannel) {
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

func (t Bot) setupBotCommands() error {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: startCommand},
		{Command: "help", Description: helpCommand},
		{Command: "track", Description: trackCommand},
		{Command: "untrack", Description: untrackCommand},
		{Command: "list", Description: listCommand},
	}
	conf := tgbotapi.NewSetMyCommands(commands...)
	if _, err := t.BotAPI.Request(conf); err != nil {
		return fmt.Errorf("error setting up bot commands: %w", err)
	}
	return nil
}

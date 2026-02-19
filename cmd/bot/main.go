package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
)

const envFilename = ".env"

func main() {
	if err := godotenv.Load(envFilename); err != nil {
		slog.Error("Error loading .env file", slog.String("err", err.Error()), slog.String("file", envFilename))
		os.Exit(1)
	}

	api, err := tgbotapi.NewBotAPI(os.Getenv("APP_TELEGRAM_TOKEN"))
	if err != nil {
		slog.Error("Error loading .env file", slog.String("err", err.Error()))
		os.Exit(1)
	}

	api.Debug = true
	telegramBot := bot.TelegramBot{Bot: api}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	wg := &sync.WaitGroup{}

	telegramHandler := handler.TelegramHandler{MsgHandler: telegramBot}
	telegramBot.Handler = telegramHandler

	telegramBot.StartMainLoop(ctx, wg)

	<-ctx.Done()
	wg.Wait()
}

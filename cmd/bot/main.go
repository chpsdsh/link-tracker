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
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/logger"
)

const (
	envFilename    = ".env"
	telegramApiKey = "APP_TELEGRAM_TOKEN"
)

func main() {
	baseLogger := logger.NewLogger(os.Stdout, logger.OutputFormatJson, slog.LevelInfo)

	if err := godotenv.Load(envFilename); err != nil {
		baseLogger.Error("error loading file", slog.String("file", envFilename), slog.String("err", err.Error()))
		os.Exit(1)
	}

	api, err := tgbotapi.NewBotAPI(os.Getenv(telegramApiKey))
	if err != nil {
		baseLogger.Error("error creating telegram bot", slog.String("err", err.Error()))
		os.Exit(1)
	}
	api.Debug = true

	telegramBot := bot.TelegramBot{Bot: api, BaseLogger: baseLogger}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	wg := &sync.WaitGroup{}

	telegramHandler := handler.TelegramHandler{MsgSender: telegramBot, BaseLogger: baseLogger}
	telegramBot.Handler = telegramHandler
	telegramBot.StartMainLoop(ctx, wg)

	<-ctx.Done()
	wg.Wait()
}

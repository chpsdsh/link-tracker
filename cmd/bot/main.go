package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/botclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/botserver"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/logger"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/statestorage"
)

const (
	envFilename      = "bot.env"
	botServerAddr    = "localhost:8080"
	shutdownDuration = 10 * time.Second
	clientTimeout    = 15 * time.Second
)

func main() {
	baseLogger := logger.NewLogger(os.Stdout, logger.OutputFormatJson, slog.LevelInfo)

	if err := godotenv.Load(envFilename); err != nil {
		baseLogger.Error("error loading file", slog.String("file", envFilename), slog.String("err", err.Error()))
		os.Exit(1)
	}

	conf, err := config.ParseConfig()
	if err != nil {
		baseLogger.Error("error parsing environment variables", slog.String("err", err.Error()))
		os.Exit(1)
	}

	api, err := tgbotapi.NewBotAPI(conf.TelegramToken)
	if err != nil {
		baseLogger.Error("failed to initialize telegram bot API", slog.String("err", err.Error()))
		os.Exit(1)
	}
	api.Debug = true

	telegramBot := bot.TelegramBot{Bot: api, BaseLogger: baseLogger}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	wg := &sync.WaitGroup{}

	client := &http.Client{Timeout: clientTimeout}

	telegramHandler := handler.TelegramHandler{MsgSender: telegramBot,
		Session:    statestorage.NewStateStorage(),
		BaseLogger: baseLogger,
		Client:     botclient.Client{Client: client}}

	telegramBot.Handler = telegramHandler

	updatesServer := botserver.UpdatesServer{BaseLogger: baseLogger, Handler: telegramHandler}
	mux := http.NewServeMux()
	h := botserver.HandlerFromMux(&updatesServer, mux)

	server := &http.Server{
		Handler: h,
		Addr:    botServerAddr,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			baseLogger.Error("error starting bot http server", slog.String("err", err.Error()))
		}
	}()

	telegramBot.StartMainLoop(ctx, wg)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		baseLogger.Error("error shutting down bot http server", slog.String("err", err.Error()))
	}

	wg.Wait()
}

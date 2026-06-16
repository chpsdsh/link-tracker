package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/botclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/botmetrics"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/receiver"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/telegram"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/database"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/statestorage"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/integration"
)

const (
	envFilename      = "deploy/bot.env"
	shutdownDuration = time.Second * 10
)

func StartBot(baseLogger *slog.Logger) error {
	if err := godotenv.Load(envFilename); err != nil {
		baseLogger.Error("error loading file", slog.String("file", envFilename), slog.String("err", err.Error()))
		return fmt.Errorf("error loading file: %w", err)
	}

	conf, err := config.ParseConfig()
	if err != nil {
		baseLogger.Error("error parsing config", slog.String("err", err.Error()))
		return fmt.Errorf("error parsing config: %w", err)
	}
	var telegramBot telegram.Bot
	if conf.WithTelegramAPI {
		api, errTgAPI := tgbotapi.NewBotAPI(conf.TelegramToken)
		if errTgAPI != nil {
			baseLogger.Error("failed to initialize telegram bot API", slog.String("err", errTgAPI.Error()))
			return fmt.Errorf("error initializing telegram bot API: %w", errTgAPI)
		}
		api.Debug = true
		telegramBot = telegram.Bot{BotAPI: api, BaseLogger: baseLogger}
	} else {
		integrationTgaAPI := integration.NewIntegrationTgAPI()
		telegramBot = telegram.Bot{BotAPI: integrationTgaAPI, BaseLogger: baseLogger}
	}

	wg := &sync.WaitGroup{}

	client := botclient.NewBotClient(conf)

	telegramHandler := handler.TelegramHandler{MsgSender: telegramBot,
		Session:    statestorage.NewStateStorage(),
		BaseLogger: baseLogger,
		Client:     client}

	telegramBot.Handler = telegramHandler

	db, err := database.NewDB(conf.PostgresConfig)
	if err != nil {
		baseLogger.Error("error connecting to database", slog.String("err", err.Error()))
		return fmt.Errorf("connecting to database: %w", err)
	}
	inboxRepo := repository.NewInboxRepository(db.GetDBPool())

	botReceiver, err := receiver.NewReceiver(conf, telegramHandler, baseLogger, inboxRepo)
	if err != nil {
		baseLogger.Error("error creating telegram receiver", slog.String("err", err.Error()))
		return fmt.Errorf("error creating telegram receiver: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if receiverStartErr := botReceiver.Start(ctx); receiverStartErr != nil {
		baseLogger.Error("error starting receiver", slog.String("err", receiverStartErr.Error()))
		cancel()
	}

	telegramBot.StartMainLoop(ctx, wg)
	metricsServer := botmetrics.NewMetricsServer(baseLogger)
	metricsServer.Start()

	botmetrics.RegisterBotMetrics()

	<-ctx.Done()

	defer cancel()
	shutdown(baseLogger, botReceiver, wg, db, metricsServer)
	return nil
}

func shutdown(baseLogger *slog.Logger, receiver receiver.Receiver, wg *sync.WaitGroup, db *database.DB, server botmetrics.Server) {
	if shutdownErr := receiver.Shutdown(); shutdownErr != nil {
		baseLogger.Error("shutdown error", slog.String("err", shutdownErr.Error()))
	}
	wg.Wait()

	db.CloseConnectionPool()
	ctx, cancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		baseLogger.Error("error shutting down metrics server", slog.String("err", err.Error()))
	}
}

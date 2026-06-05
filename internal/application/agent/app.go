package agent

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/consumer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/grouper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/prioritizer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/producer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/summarizer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/database"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/agent/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

const (
	envFilename             = "deploy/agent.env"
	notificationChanBufSize = 10
)

func StartAgent(baseLogger *slog.Logger) error {
	if err := godotenv.Load(envFilename); err != nil {
		baseLogger.Error("error loading file", slog.String("file", envFilename), slog.String("err", err.Error()))
		return fmt.Errorf("error loading file: %w", err)
	}

	conf, err := config.ParseAIAgentConfig()
	if err != nil {
		baseLogger.Error("error parsing config", slog.String("err", err.Error()))
		return fmt.Errorf("error parsing config: %w", err)
	}

	db, err := database.NewDB(conf.PostgresConfig)
	if err != nil {
		baseLogger.Error("error connecting to database", slog.String("err", err.Error()))
		return fmt.Errorf("connecting to database: %w", err)
	}

	inboxRepo := repository.NewInboxRepository(db.GetDBPool())

	notificationsChan := make(chan pkg.KafkaLinkUpdate, notificationChanBufSize)

	updatesProducer, err := producer.NewKafkaProducer(conf.KafkaConfig, baseLogger, notificationsChan)
	if err != nil {
		baseLogger.Error("error creating kafka producer", slog.String("err", err.Error()))
		return fmt.Errorf("creating kafka producer: %w", err)
	}

	agentSummarizer := summarizer.NewLinksSummarizer(conf)

	agentService := &service.AgentService{
		Summarizer:  agentSummarizer,
		Sender:      updatesProducer,
		BaseLogger:  baseLogger,
		Prioritizer: prioritizer.Prioritizer{Config: conf},
	}
	updatesGrouper := grouper.NewUpdatesGrouper(baseLogger, agentService)
	agentService.Grouper = updatesGrouper

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	scheduler, err := service.NewGroupScheduler(updatesGrouper)
	if err != nil {
		baseLogger.Error("error creating group scheduler", slog.String("err", err.Error()))
		return fmt.Errorf("creating group scheduler: %w", err)
	}
	scheduler.StartAgentScheduler(conf)
	updatesProducer.StartProducerLoop(ctx)

	updatesConsumer, err := consumer.NewNotificationsConsumer(conf.KafkaConfig, baseLogger, agentService, inboxRepo)
	if err != nil {
		baseLogger.Error("error creating kafka consumer", slog.String("err", err.Error()))
		return fmt.Errorf("error creating kafka notifications consumer: %w", err)
	}

	if err = updatesConsumer.Start(ctx); err != nil {
		return fmt.Errorf("error starting kafka consumer: %w", err)
	}

	<-ctx.Done()
	shutdown(baseLogger, db, updatesProducer, notificationsChan, updatesConsumer, scheduler)

	return nil
}

func shutdown(baseLogger *slog.Logger, db *database.DB, updatesProducer *producer.KafkaProducer, notificationsChan chan pkg.KafkaLinkUpdate, updatesConsumer *consumer.NotificationsConsumer, scheduler service.GroupScheduler) {
	db.CloseConnectionPool()

	updatesProducer.Close()

	close(notificationsChan)

	if err := updatesConsumer.Shutdown(); err != nil {
		baseLogger.Error("error shutting down kafka consumer", slog.String("err", err.Error()))
	}

	if err := scheduler.ShutdownScheduler(); err != nil {
		baseLogger.Error("error shutting down scheduler", slog.String("err", err.Error()))
	}
}

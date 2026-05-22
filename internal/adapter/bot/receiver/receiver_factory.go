package receiver

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/receiver/botserver"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/receiver/consumer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/receiver/fallbackreceiver"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
)

const (
	kafkaReceiverType    = "kafka"
	httpReceiverType     = "http"
	fallbackReceiverType = "fallback"
)

type Receiver interface {
	Start(ctx context.Context) error
	Shutdown() error
}

func NewReceiver(conf config.BotConfig, telegramHandler handler.TelegramHandler, logger *slog.Logger, repository consumer.InboxRepository) (Receiver, error) {
	switch conf.UpdatesReceiveType {
	case kafkaReceiverType:
		notificationConsumer, err := consumer.NewBotNotificationsConsumer(conf, logger, telegramHandler, repository)
		if err != nil {
			return nil, fmt.Errorf("error creating kafka notifications consumer: %w", err)
		}
		return notificationConsumer, nil
	case httpReceiverType:
		server := botserver.NewBotHTTPServer(logger, telegramHandler, conf)
		return server, nil
	case fallbackReceiverType:
		notificationConsumer, err := consumer.NewBotNotificationsConsumer(conf, logger, telegramHandler, repository)
		if err != nil {
			return nil, fmt.Errorf("error creating kafka notifications consumer: %w", err)
		}
		botServer := botserver.NewBotHTTPServer(logger, telegramHandler, conf)
		return fallbackreceiver.NewFallbackReceiver(notificationConsumer, botServer), nil
	}
	return nil, fmt.Errorf("unknown receiver type: %s", conf.UpdatesReceiveType)
}

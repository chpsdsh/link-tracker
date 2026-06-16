package fallbackreceiver

import (
	"context"
	"errors"
	"fmt"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/receiver/botserver"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/receiver/consumer"
)

type FallbackReceiver struct {
	kafkaConsumer *consumer.NotificationsConsumer
	botServer     botserver.BotHTTPServer
}

func NewFallbackReceiver(kafkaConsumer *consumer.NotificationsConsumer, botServer botserver.BotHTTPServer) FallbackReceiver {
	return FallbackReceiver{kafkaConsumer: kafkaConsumer, botServer: botServer}
}

func (r FallbackReceiver) Start(ctx context.Context) error {
	if err := r.kafkaConsumer.Start(ctx); err != nil {
		return fmt.Errorf("start kafka consumer: %w", err)
	}
	if err := r.botServer.Start(ctx); err != nil {
		return fmt.Errorf("start bot server: %w", err)
	}
	return nil
}

func (r FallbackReceiver) Shutdown() error {
	consumerCloseErr := r.kafkaConsumer.Shutdown()
	botServerCloseErr := r.botServer.Shutdown()
	return errors.Join(consumerCloseErr, botServerCloseErr)
}

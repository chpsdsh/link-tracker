package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/metrics"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/dlqproducer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
)

const (
	kafkaScope = "kafka"
)

type NotificationsConsumer struct {
	notificationTopic string
	consumer          sarama.ConsumerGroup
	wg                sync.WaitGroup
	logger            *slog.Logger
	dlqProducer       dlqproducer.DlqProducer
	groupHandler      GroupHandler
}

func NewBotNotificationsConsumer(conf config.BotConfig, logger *slog.Logger, tgHandler handler.TelegramBotHandler, inboxRepo InboxRepository) (*NotificationsConsumer, error) {
	saramaConf := sarama.NewConfig()

	saramaConf.Version = sarama.V3_6_0_0

	saramaConf.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConf.Consumer.Offsets.AutoCommit.Enable = false
	saramaConf.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(conf.KafkaConfig.Brokers, conf.KafkaConfig.ConsumerGroup, saramaConf)
	if err != nil {
		return nil, fmt.Errorf("creating consumer group connection: %w", err)
	}

	dlqProducer, err := dlqproducer.NewDLQProducer(conf.KafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("creating dlq producer: %w", err)
	}

	groupHandler := NewGroupHandler(dlqProducer, tgHandler, inboxRepo, logger, conf.KafkaConfig.ConsumerGroup)

	return &NotificationsConsumer{consumer: consumer,
		notificationTopic: conf.KafkaConfig.ProcessedNotificationsTopic,
		wg:                sync.WaitGroup{},
		logger:            logger,
		dlqProducer:       dlqProducer,
		groupHandler:      groupHandler}, nil
}

func (c *NotificationsConsumer) Start(ctx context.Context) error {
	c.wg.Go(func() {
		defer func() {
			if closeErr := c.dlqProducer.Close(); closeErr != nil {
				c.logger.Error("error closing dlq producer", slog.String("error", closeErr.Error()))
			}
		}()

		for {
			if consumeErr := c.consumer.Consume(ctx, []string{c.notificationTopic}, c.groupHandler); consumeErr != nil {
				c.logger.Error("error consuming", slog.String("err", consumeErr.Error()))
				continue
			}
			if ctx.Err() != nil {
				return
			}
		}
	})

	c.wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case groupErr, ok := <-c.consumer.Errors():
				if !ok {
					return
				}
				metrics.ErrorsCounterTotal.WithLabelValues(kafkaScope, "consumer-error").Inc()
				c.logger.Error("error consuming", slog.String("err", groupErr.Error()))
			}
		}
	})
	return nil
}

func (c *NotificationsConsumer) Shutdown() error {
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("closing consumer group: %w", err)
	}
	c.wg.Wait()
	return nil
}

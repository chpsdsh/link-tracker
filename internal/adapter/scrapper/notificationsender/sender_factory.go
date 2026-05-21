package notificationsender

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/notificationsender/fallbacksender"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/notificationsender/httpsender"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/notificationsender/producer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

var ErrNotValidSenderType = errors.New("not valid kafka sender type")

const (
	kafkaSender    = "kafka"
	httpSender     = "http"
	fallbackSender = "fallback"
)

func NewSender(ctx context.Context, conf config.ScrapperConfig, logger *slog.Logger, updatesChan chan pkg.KafkaLinkUpdate) (service.Sender, error) {
	switch conf.UpdatesSendType {
	case kafkaSender:
		notificationProducer, err := producer.NewKafkaProducer(conf, logger, updatesChan)
		if err != nil {
			return nil, fmt.Errorf("error creating kafka producer: %w", err)
		}
		notificationProducer.StartProducerLoop(ctx)
		return notificationProducer, nil
	case httpSender:
		return httpsender.NewUpdatesClient(conf), nil
	case fallbackSender:
		notificationProducer, err := producer.NewKafkaProducer(conf, logger, updatesChan)
		if err != nil {
			return nil, fmt.Errorf("error creating kafka producer: %w", err)
		}
		notificationProducer.StartProducerLoop(ctx)
		httpClient := httpsender.NewUpdatesClient(conf)
		return fallbacksender.NewFallbackSender(notificationProducer, httpClient), nil
	default:
		return nil, ErrNotValidSenderType
	}
}

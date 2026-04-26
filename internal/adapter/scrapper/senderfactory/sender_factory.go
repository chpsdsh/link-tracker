package senderfactory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/producer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/scrapperclient"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

var ErrNotValidSenderType = errors.New("not valid kafka sender type")

const (
	kafkaSender = "kafka"
	httpSender  = "http"
)

func NewSender(ctx context.Context, conf config.Config, logger *slog.Logger, updatesChan chan pkg.KafkaLinkUpdate) (service.Sender, error) {
	switch conf.UpdatesSendType {
	case kafkaSender:
		producer, err := producer.NewKafkaProducer(conf, logger, updatesChan)
		if err != nil {
			return nil, fmt.Errorf("error creating kafka producer: %w", err)
		}
		producer.StartProducerLoop(ctx)
		return producer, nil
	case httpSender:
		return scrapperclient.NewUpdatesClient(conf), nil
	default:
		return nil, ErrNotValidSenderType
	}
}

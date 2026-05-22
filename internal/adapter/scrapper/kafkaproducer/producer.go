package kafkaproducer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

const (
	producerRetryMax       = 10
	producerFlushMessages  = 100
	producerFlushFrequency = 500 * time.Millisecond
	eventIDKeyName         = "event_id"
)

type KafkaProducer struct {
	producer          sarama.AsyncProducer
	NotificationTopic string
	wg                sync.WaitGroup
	closeOnce         sync.Once
	logger            *slog.Logger
	linkUpdatesChan   chan pkg.KafkaLinkUpdate
}

func NewKafkaProducer(conf config.KafkaConfig, logger *slog.Logger, linkUpdatesChan chan pkg.KafkaLinkUpdate) (*KafkaProducer, error) {
	saramaConf := sarama.NewConfig()

	saramaConf.Version = sarama.V3_6_0_0
	saramaConf.Producer.Partitioner = sarama.NewHashPartitioner
	saramaConf.Producer.Retry.Max = producerRetryMax
	saramaConf.Producer.RequiredAcks = sarama.WaitForAll
	saramaConf.Producer.Flush.Messages = producerFlushMessages
	saramaConf.Producer.Flush.Frequency = producerFlushFrequency

	producer, err := sarama.NewAsyncProducer(conf.Brokers, saramaConf)
	if err != nil {
		return &KafkaProducer{}, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaProducer{producer: producer,
		NotificationTopic: conf.RawNotificationsTopic,
		linkUpdatesChan:   linkUpdatesChan,
		wg:                sync.WaitGroup{},
		closeOnce:         sync.Once{},
		logger:            logger}, nil
}

func (p *KafkaProducer) StartProducerLoop(ctx context.Context) {
	p.wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				p.closeProducerOnes()
				return
			case update, ok := <-p.linkUpdatesChan:
				if !ok {
					p.closeProducerOnes()
					return
				}

				msg := &sarama.ProducerMessage{
					Topic: p.NotificationTopic,
					Key:   sarama.StringEncoder(update.Key),
					Value: sarama.ByteEncoder(update.Value),
					Headers: []sarama.RecordHeader{{
						Key:   []byte(eventIDKeyName),
						Value: []byte(update.EventID),
					},
					},
				}

				select {
				case p.producer.Input() <- msg:
				case <-ctx.Done():
					p.closeProducerOnes()
					return
				}
			}
		}
	})

	p.wg.Go(func() {
		for err := range p.producer.Errors() {
			p.logger.Error("producer error",
				slog.String("topic", p.NotificationTopic),
				slog.String("err", err.Error()),
			)
		}
	})
}

func (p *KafkaProducer) SendLinkUpdate(update pkg.LinkUpdate, eventID string) error {
	bytes, err := json.Marshal(update)
	if err != nil {
		p.logger.Error("marshalling error:", slog.String("error", err.Error()))
		return fmt.Errorf("marshalling error: %w", err)
	}

	linkUpdate := pkg.KafkaLinkUpdate{Key: update.URL, Value: bytes, EventID: eventID}

	p.linkUpdatesChan <- linkUpdate
	return nil
}

func (p *KafkaProducer) Close() {
	p.wg.Wait()
}

func (p *KafkaProducer) closeProducerOnes() {
	p.closeOnce.Do(func() {
		p.producer.AsyncClose()
	})
}

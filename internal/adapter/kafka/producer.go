package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/kafka/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

const (
	producerRetryMax       = 10
	producerFlushMessages  = 100
	producerFLushFrequency = 500 * time.Millisecond
)

type Producer struct {
	producer          sarama.AsyncProducer
	NotificationTopic string
	wg                sync.WaitGroup
	closeOnce         sync.Once
	logger            slog.Logger
}

func NewKafkaProducer(kafkaConf config.KafkaConfig, logger slog.Logger) (Producer, error) {
	saramaConf := sarama.NewConfig()

	saramaConf.Version = sarama.V3_6_0_0
	saramaConf.Producer.Partitioner = sarama.NewHashPartitioner
	saramaConf.Producer.Retry.Max = producerRetryMax
	saramaConf.Producer.RequiredAcks = sarama.WaitForAll
	saramaConf.Producer.Flush.Messages = producerFlushMessages
	saramaConf.Producer.Flush.Frequency = producerFLushFrequency

	producer, err := sarama.NewAsyncProducer(kafkaConf.Brokers, saramaConf)
	if err != nil {
		return Producer{}, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return Producer{producer: producer,
		NotificationTopic: kafkaConf.NotificationsTopic,
		wg:                sync.WaitGroup{},
		closeOnce:         sync.Once{},
		logger:            logger}, nil
}

func (p *Producer) StartProducerLoop(ctx context.Context, linkUpdatesChan <-chan pkg.LinkUpdate) {
	p.wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				p.closeProducerOnes()
				return
			case update, ok := <-linkUpdatesChan:
				if !ok {
					p.closeProducerOnes()
					return
				}

				bytes, err := json.Marshal(update)
				if err != nil {
					p.logger.Error("marshalling error:", slog.String("error", err.Error()))
					continue
				}

				msg := &sarama.ProducerMessage{
					Topic: p.NotificationTopic,
					Key:   sarama.StringEncoder(update.URL),
					Value: sarama.ByteEncoder(bytes),
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

func (p *Producer) closeProducerOnes() {
	p.closeOnce.Do(func() {
		p.producer.AsyncClose()
	})
}

func (p *Producer) Close() {
	p.wg.Wait()
}

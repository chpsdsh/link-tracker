package dlqproducer

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/metrics"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
)

const (
	producerRetryMax     = 10
	kafkaScope           = "kafkaScope"
	dlqProducerScopeType = "dlq_producer"
)

type DlqProducer struct {
	producer sarama.SyncProducer
	dLQTopic string
}

func NewDLQProducer(conf config.KafkaConfig) (DlqProducer, error) {
	saramaConf := sarama.NewConfig()

	saramaConf.Version = sarama.V3_6_0_0
	saramaConf.Producer.Partitioner = sarama.NewHashPartitioner
	saramaConf.Producer.Return.Successes = true
	saramaConf.Producer.Retry.Max = producerRetryMax
	saramaConf.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(conf.Brokers, saramaConf)
	if err != nil {
		return DlqProducer{}, fmt.Errorf("creating DLQ producer: %w", err)
	}
	return DlqProducer{producer: producer, dLQTopic: conf.DLQTopic}, nil
}

func (p DlqProducer) SendToDLQ(msg *sarama.ConsumerMessage, reason error) error {
	payload := map[string]interface{}{
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"key":       string(msg.Key),
		"value":     string(msg.Value),
		"error":     reason.Error(),
		"ts":        time.Now().UTC(),
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal dlq: %w", err)
	}
	producerMsg := &sarama.ProducerMessage{
		Topic: p.dLQTopic,
		Key:   sarama.ByteEncoder(msg.Key),
		Value: sarama.ByteEncoder(bytes),
	}

	startTime := time.Now()
	_, _, err = p.producer.SendMessage(producerMsg)
	metrics.ObserveCommandDuration(startTime, kafkaScope, dlqProducerScopeType)

	if err != nil {
		return fmt.Errorf("send to dlq: %w", err)
	}

	return nil
}

func (p DlqProducer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("close dlq producer: %w", err)
	}
	return nil
}

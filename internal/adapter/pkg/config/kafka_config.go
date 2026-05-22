package config

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrNoKafkaUser           = errors.New("kafka user should be set with " + kafkaUser + " environment variable")
	ErrNoKafkaPassword       = errors.New("kafka password should be set with " + kafkaPassword + " environment variable")
	ErrNoKafkaTopic          = errors.New("kafka raw topic should be set with " + kafkaRawTopic + " environment variable")
	ErrNoKafkaDLQTopic       = errors.New("kafka DLQ topic should be set with " + kafkaDLQTopic + " environment variable")
	ErrNoKafkaBrokers        = errors.New("kafka brokers should be set with " + kafkaBrokers + " environment variable")
	ErrNoKafkaConsumerGroup  = errors.New("kafka consumer group should be set with " + kafkaConsumerGroup + " environment variable")
	ErrNoKafkaProcessedTopic = errors.New("kafka processed topic should be set with " + kafkaProcessedTopic + " environment variable")
)

const (
	kafkaUser           = "KAFKA_USER"
	kafkaPassword       = "KAFKA_PASSWORD"
	kafkaRawTopic       = "KAFKA_RAW_TOPIC"
	kafkaProcessedTopic = "KAFKA_PROCESSED_TOPIC"
	kafkaDLQTopic       = "KAFKA_DLQ_TOPIC"
	kafkaConsumerGroup  = "KAFKA_CONSUMER_GROUP"
	kafkaBrokers        = "KAFKA_BROKERS"
)

type KafkaConfig struct {
	Brokers                     []string
	RawNotificationsTopic       string
	ProcessedNotificationsTopic string
	DLQTopic                    string
	ConsumerGroup               string
	User                        string
	Password                    string
}

func ParseKafkaConfig() (KafkaConfig, error) {
	user := os.Getenv(kafkaUser)
	if user == "" {
		return KafkaConfig{}, ErrNoKafkaUser
	}

	password := os.Getenv(kafkaPassword)
	if password == "" {
		return KafkaConfig{}, ErrNoKafkaPassword
	}

	topic := os.Getenv(kafkaRawTopic)
	if topic == "" {
		return KafkaConfig{}, ErrNoKafkaTopic
	}

	dlqTopic := os.Getenv(kafkaDLQTopic)
	if dlqTopic == "" {
		return KafkaConfig{}, ErrNoKafkaDLQTopic
	}

	brokers := os.Getenv(kafkaBrokers)
	if brokers == "" {
		return KafkaConfig{}, ErrNoKafkaBrokers
	}

	consumerGroup := os.Getenv(kafkaConsumerGroup)
	if consumerGroup == "" {
		return KafkaConfig{}, ErrNoKafkaConsumerGroup
	}

	processedNotificationsTopic := os.Getenv(kafkaProcessedTopic)
	if processedNotificationsTopic == "" {
		return KafkaConfig{}, ErrNoKafkaProcessedTopic
	}

	return KafkaConfig{
		Brokers:                     strings.Split(brokers, ","),
		RawNotificationsTopic:       topic,
		ProcessedNotificationsTopic: processedNotificationsTopic,
		User:                        user,
		Password:                    password,
		DLQTopic:                    dlqTopic,
		ConsumerGroup:               consumerGroup,
	}, nil
}

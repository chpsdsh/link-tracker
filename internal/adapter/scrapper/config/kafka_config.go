package config

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrNoKafkaUser     = errors.New("kafka user should be set with " + kafkaUser + " environment variable")
	ErrNoKafkaPassword = errors.New("kafka password should be set with " + kafkaPassword + " environment variable")
	ErrNoKafkaTopic    = errors.New("kafka topic should be set with " + kafkaTopic + " environment variable")
	ErrNoKafkaBrokers  = errors.New("kafka brokers should be set with " + kafkaBrokers + " environment variable")
)

const (
	kafkaUser     = "KAFKA_USER"
	kafkaPassword = "KAFKA_PASSWORD"
	kafkaTopic    = "KAFKA_TOPIC"
	kafkaBrokers  = "KAFKA_BROKERS"
)

type KafkaConfig struct {
	Brokers            []string
	NotificationsTopic string
	User               string
	Password           string
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

	topic := os.Getenv(kafkaTopic)
	if topic == "" {
		return KafkaConfig{}, ErrNoKafkaTopic
	}

	brokers := os.Getenv(kafkaBrokers)
	if brokers == "" {
		return KafkaConfig{}, ErrNoKafkaBrokers
	}

	return KafkaConfig{
		Brokers:            strings.Split(brokers, ","),
		NotificationsTopic: topic,
		User:               user,
		Password:           password,
	}, nil
}

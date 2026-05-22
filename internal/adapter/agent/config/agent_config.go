package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
)

const (
	aiStopWords              = "AI_STOP_WORDS"
	aiExcludedAuthors        = "AI_EXCLUDED_AUTHORS"
	aiMinLength              = "AI_MIN_LENGTH"
	aiSummarizationThreshold = "AI_SUMMARIZATION_THRESHOLD"
)

var (
	ErrNoAIMinLength                   = errors.New("ai min length should be set with " + aiMinLength + " environment variable")
	ErrInvalidAIMinLength              = errors.New(aiMinLength + " should be non-negative integer")
	ErrNoAISummarizationThreshold      = errors.New("ai summarization threshold should be set with " + aiSummarizationThreshold + " environment variable")
	ErrInvalidAISummarizationThreshold = errors.New(aiSummarizationThreshold + " should be positive integer")
)

type AIAgentConfig struct {
	StopWords       []string
	ExcludedAuthors []string
	MinLength       int
	Threshold       int
	KafkaConfig     config.KafkaConfig
	PostgresConfig  config.PostgresConfig
}

func ParseAIAgentConfig() (AIAgentConfig, error) {
	minLengthStr := os.Getenv(aiMinLength)
	if minLengthStr == "" {
		return AIAgentConfig{}, ErrNoAIMinLength
	}

	minLength, err := strconv.Atoi(minLengthStr)
	if err != nil || minLength < 0 {
		return AIAgentConfig{}, ErrInvalidAIMinLength
	}

	thresholdStr := os.Getenv(aiSummarizationThreshold)
	if thresholdStr == "" {
		return AIAgentConfig{}, ErrNoAISummarizationThreshold
	}

	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil || threshold <= 0 {
		return AIAgentConfig{}, ErrInvalidAISummarizationThreshold
	}
	kafkaConfig, err := config.ParseKafkaConfig()
	if err != nil {
		return AIAgentConfig{}, fmt.Errorf("failed to parse kafka config: %w", err)
	}

	postgresConfig, err := config.ParsePostgresConfig()
	if err != nil {
		return AIAgentConfig{}, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	return AIAgentConfig{
		StopWords:       parseStringList(os.Getenv(aiStopWords)),
		ExcludedAuthors: parseStringList(os.Getenv(aiExcludedAuthors)),
		MinLength:       minLength,
		Threshold:       threshold,
		KafkaConfig:     kafkaConfig,
		PostgresConfig:  postgresConfig,
	}, nil
}

func parseStringList(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}

		result = append(result, item)
	}

	return result
}

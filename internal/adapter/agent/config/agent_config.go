package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
)

const (
	aiStopWords              = "AI_STOP_WORDS"
	aiExcludedAuthors        = "AI_EXCLUDED_AUTHORS"
	aiMinLength              = "AI_MIN_LENGTH"
	aiSummarizationThreshold = "AI_SUMMARIZATION_THRESHOLD"
	aiHighPriorityKeyWords   = "AI_HIGH_PRIORITY_KEY_WORDS"

	aiLowPriorityKeyWords = "AI_LOW_PRIORITY_KEY_WORDS"
	groupWindowMS         = "GROUP_WINDOW_MS"
)

var (
	ErrNoAIMinLength                   = errors.New("ai min length should be set with " + aiMinLength + " environment variable")
	ErrInvalidAIMinLength              = errors.New(aiMinLength + " should be non-negative integer")
	ErrNoAISummarizationThreshold      = errors.New("ai summarization threshold should be set with " + aiSummarizationThreshold + " environment variable")
	ErrInvalidAISummarizationThreshold = errors.New(aiSummarizationThreshold + " should be positive integer")
	ErrNoAIHighPriorityKeyWords        = errors.New("ai high priority key words should be set with " + aiHighPriorityKeyWords + " environment variable")
	ErrNoAILowPriorityKeyWords         = errors.New("ai low priority key words should be set with " + aiLowPriorityKeyWords + " environment variable")
	ErrNoGroupWindowMS                 = errors.New("group window ms should be set with " + groupWindowMS + " environment variable")
	ErrInvalidGroupWindowMS            = errors.New(groupWindowMS + " should be positive integer")
)

type AIAgentConfig struct {
	StopWords            []string
	ExcludedAuthors      []string
	MinLength            int
	Threshold            int
	HighPriorityKeyWords []string
	LowPriorityKeyWords  []string
	GroupWindow          time.Duration
	KafkaConfig          config.KafkaConfig
	PostgresConfig       config.PostgresConfig
	YandexAgentConfig    YandexAgentConfig
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

	yandexAgentConfig, err := ParseYandexConfig()
	if err != nil {
		return AIAgentConfig{}, fmt.Errorf("failed to parse yandex agent config: %w", err)
	}

	highPriorityKeyWordsStr := os.Getenv(aiHighPriorityKeyWords)

	if highPriorityKeyWordsStr == "" {
		return AIAgentConfig{}, ErrNoAIHighPriorityKeyWords
	}
	highPriorityKeyWords := parseStringList(highPriorityKeyWordsStr)
	if len(highPriorityKeyWords) == 0 {
		return AIAgentConfig{}, ErrNoAIHighPriorityKeyWords
	}
	lowPriorityKeyWordsStr := os.Getenv(aiLowPriorityKeyWords)
	if lowPriorityKeyWordsStr == "" {
		return AIAgentConfig{}, ErrNoAILowPriorityKeyWords
	}
	lowPriorityKeyWords := parseStringList(lowPriorityKeyWordsStr)
	if len(lowPriorityKeyWords) == 0 {
		return AIAgentConfig{}, ErrNoAILowPriorityKeyWords
	}
	groupWindowMSStr := os.Getenv(groupWindowMS)
	if groupWindowMSStr == "" {
		return AIAgentConfig{}, ErrNoGroupWindowMS
	}
	groupWindowMSValue, err := strconv.Atoi(groupWindowMSStr)
	if err != nil || groupWindowMSValue <= 0 {
		return AIAgentConfig{}, ErrInvalidGroupWindowMS
	}

	return AIAgentConfig{
		StopWords:            parseStringList(os.Getenv(aiStopWords)),
		ExcludedAuthors:      parseStringList(os.Getenv(aiExcludedAuthors)),
		MinLength:            minLength,
		Threshold:            threshold,
		HighPriorityKeyWords: highPriorityKeyWords,
		LowPriorityKeyWords:  lowPriorityKeyWords,
		GroupWindow:          time.Duration(groupWindowMSValue) * time.Millisecond,
		KafkaConfig:          kafkaConfig,
		PostgresConfig:       postgresConfig,
		YandexAgentConfig:    yandexAgentConfig,
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

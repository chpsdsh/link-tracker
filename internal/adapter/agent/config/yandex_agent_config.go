package config

import (
	"errors"
	"os"
)

const (
	yandexAPIKey   = "YANDEX_API_KEY"
	yandexFolderID = "YANDEX_FOLDER_ID"
	yandexModel    = "YANDEX_MODEL"
	yandexBaseURL  = "YANDEX_BASE_URL"
)

var (
	ErrNoYandexAPIKey   = errors.New("yandex api key should be set with " + yandexAPIKey + " environment variable")
	ErrNoYandexFolderID = errors.New("yandex folder id should be set with " + yandexFolderID + " environment variable")
	ErrNoYandexModel    = errors.New("yandex model should be set with " + yandexModel + " environment variable")
	ErrNoYandexBaseURL  = errors.New("yandex base url should be set with " + yandexBaseURL + " environment variable")
)

type YandexAgentConfig struct {
	APIKey   string
	FolderID string
	Model    string
	BaseURL  string
}

func ParseYandexConfig() (YandexAgentConfig, error) {
	apiKey := os.Getenv(yandexAPIKey)
	if apiKey == "" {
		return YandexAgentConfig{}, ErrNoYandexAPIKey
	}

	folderID := os.Getenv(yandexFolderID)
	if folderID == "" {
		return YandexAgentConfig{}, ErrNoYandexFolderID
	}

	model := os.Getenv(yandexModel)
	if model == "" {
		return YandexAgentConfig{}, ErrNoYandexModel
	}

	baseURL := os.Getenv(yandexBaseURL)
	if baseURL == "" {
		return YandexAgentConfig{}, ErrNoYandexBaseURL
	}

	return YandexAgentConfig{
		APIKey:   apiKey,
		FolderID: folderID,
		Model:    model,
		BaseURL:  baseURL,
	}, nil
}

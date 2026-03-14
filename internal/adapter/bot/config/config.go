package config

import (
	"errors"
	"os"
)

const (
	telegramAPIKey        = "APP_TELEGRAM_TOKEN"
	scrapperServerAddress = "SCRAPPER_SERVER_ADDRESS"
)

var (
	ErrNoTelegramToken   = errors.New("telegram token should be set up with " + telegramAPIKey + " environment variable")
	ErrNoScrapperAddress = errors.New("scrapper server address should be set up with " + scrapperServerAddress + " environment variable")
)

type Config struct {
	TelegramToken         string
	ScrapperServerAddress string
}

func ParseConfig() (Config, error) {
	token := os.Getenv(telegramAPIKey)
	if token == "" {
		return Config{}, ErrNoTelegramToken
	}
	scrapperAddr := os.Getenv(scrapperServerAddress)
	if scrapperAddr == "" {
		return Config{}, ErrNoScrapperAddress
	}
	return Config{TelegramToken: token, ScrapperServerAddress: scrapperAddr}, nil
}

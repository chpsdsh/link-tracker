package config

import (
	"errors"
	"os"
)

const (
	telegramApiKey        = "APP_TELEGRAM_TOKEN"
	scrapperServerAddress = "SCRAPPER_SERVER_ADDRESS"
)

var (
	NoTelegramTokenError   = errors.New("telegram token should be set up with " + telegramApiKey + " environment variable")
	NoScrapperAddressError = errors.New("scrapper server address should be set up with " + scrapperServerAddress + " environment variable")
)

type Config struct {
	TelegramToken         string
	ScrapperServerAddress string
}

func ParseConfig() (Config, error) {
	token := os.Getenv(telegramApiKey)
	if token == "" {
		return Config{}, NoTelegramTokenError
	}
	scrapperAddr := os.Getenv(scrapperServerAddress)
	if scrapperAddr == "" {
		return Config{}, NoScrapperAddressError
	}
	return Config{TelegramToken: token, ScrapperServerAddress: scrapperAddr}, nil
}

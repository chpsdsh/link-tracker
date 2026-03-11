package config

import (
	"errors"
	"os"
)

const telegramApiKey = "APP_TELEGRAM_TOKEN"

type Config struct {
	TelegramToken string
}

func ParseConfig() (Config, error) {
	token := os.Getenv(telegramApiKey)
	if token == "" {
		return Config{}, errors.New("telegram token should be set up with " + telegramApiKey + " environment variable")
	}
	return Config{TelegramToken: token}, nil
}

package main

import (
	"log/slog"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/telegram"
)

const envFilename = ".env"

func main() {
	err := godotenv.Load(envFilename)
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("APP_TELEGRAM_TOKEN"))
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}

	bot.Debug = true

	telegramBot := telegram.TelegramBot{Bot: bot}
	telegramBot.StartMainLoop()
}

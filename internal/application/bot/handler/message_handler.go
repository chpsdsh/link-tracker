package handler

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/statestorage"
)

const (
	helpMessage = `
/track — начать отслеживание ссылки. Опционально можно указать теги.
/untrack — прекратить отслеживание ссылки.
/list — вывести список всех отслеживаемых ссылок.
`
	greetingMessage       = "Добро пожаловать! Используйте /help, чтобы посмотреть доступные команды."
	unknownMessage        = "Воспользуйтесь /help, чтобы посмотреть список доступных команд."
	cancelMessage         = "Операция отменена"
	trackMessage          = "Введите ссылку для отслеживания"
	untrackMessage        = "Введите ссылку для прекращения отслеживания"
	tagsMessage           = "Введи список тегов через запятую"
	trackConfirmMessage   = "Ссылка добавлена для отслеживания"
	untrackConfirmMessage = "Ссылка удалена из отслеживания"
)

type Sender interface {
	SendMessage(chatID int64, message string) error
}

type StateStorage interface {
	GetState(chatId int64) statestorage.State
	SetState(chatId int64, state statestorage.State)
}

type TelegramHandler struct {
	MsgSender  Sender
	Session    StateStorage
	BaseLogger *slog.Logger
}

func (h TelegramHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message.IsCommand() {
		h.handleCommand(update)
	} else {
		h.handleMessage(update)
	}
}

func (h TelegramHandler) handleCommand(update tgbotapi.Update) {
	var text string
	switch update.Message.Command() {
	case "start":
		text = greetingMessage
		h.Session.SetState(update.Message.Chat.ID, statestorage.InitialState)
	case "help":
		text = helpMessage
	case "track":
		text = trackMessage
		h.Session.SetState(update.Message.Chat.ID, statestorage.WaitingForTrackUrlState)
	case "untrack":
		text = untrackMessage
		h.Session.SetState(update.Message.Chat.ID, statestorage.WaitingForUnTrackUrlState)
	case "list":

	case "cancel":
		text = cancelMessage
		h.Session.SetState(update.Message.Chat.ID, statestorage.InitialState)
	default:
		text = unknownMessage
	}

	if err := h.MsgSender.SendMessage(update.Message.Chat.ID, text); err != nil {
		h.BaseLogger.Error("error while sending message", slog.String("error", err.Error()))
	}
}

func (h TelegramHandler) handleMessage(update tgbotapi.Update) {
	var text string
	switch h.Session.GetState(update.Message.Chat.ID) {
	case statestorage.WaitingForTrackUrlState:
		text = tagsMessage
		h.Session.SetState(update.Message.Chat.ID, statestorage.WaitingForTagsState)
	case statestorage.WaitingForTagsState:
		text = trackConfirmMessage
		h.Session.SetState(update.Message.Chat.ID, statestorage.InitialState)
	case statestorage.WaitingForUnTrackUrlState:
		text = untrackConfirmMessage
		h.Session.SetState(update.Message.Chat.ID, statestorage.InitialState)
	default:
		text = unknownMessage
	}

	if err := h.MsgSender.SendMessage(update.Message.Chat.ID, text); err != nil {
		h.BaseLogger.Error("error while sending message", slog.String("error", err.Error()))
	}
}

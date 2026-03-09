package handler

import (
	"errors"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/statestorage"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

var (
	ErrChatNotFound               = errors.New("chat not found")
	ErrLinkExists                 = errors.New("link already tracked")
	ErrIncorrectRequestParameters = errors.New("incorrect request parameters")
	ErrChatAlreadyExists          = errors.New("chat already exists")
	ErrLinkNotExists              = errors.New("link not exists")
)

const (
	helpMessage = `
/track — начать отслеживание ссылки. Опционально можно указать теги.
/untrack — прекратить отслеживание ссылки.
/list — вывести список всех отслеживаемых ссылок.
`
	greetingMessage            = "Добро пожаловать! Используйте /help, чтобы посмотреть доступные команды."
	incorrectRequestParameters = "Некорректные параметры запроса"
	chatAlreadyExists          = "Вы уже зарегестирировали чат, чтобы посмотреть лоступные команды используйте /help"
	unknownMessage             = "Воспользуйтесь /help, чтобы посмотреть список доступных команд."
	cancelMessage              = "Операция отменена"
	trackMessage               = "Введите ссылку для отслеживания"
	untrackMessage             = "Введите ссылку для прекращения отслеживания"
	tagsMessage                = "Введи список тегов через запятую"
	trackConfirmMessage        = "Ссылка добавлена для отслеживания"
	untrackConfirmMessage      = "Ссылка удалена из отслеживания"
	noTrackedLinks             = "Нет отслеживаемых ссылок"
)

type Sender interface {
	SendMessage(chatID int64, message string) error
}

type StateStorage interface {
	GetState(chatId int64) statestorage.State
	SetState(chatId int64, state statestorage.State)
}

type NetworkClient interface {
	RegisterChat(chatId int64) error
	GetLinks(chatId int64) (bot.ListLinksResponse, error)
	AddLink(chatId int64)
}

type TelegramHandler struct {
	MsgSender  Sender
	Session    StateStorage
	BaseLogger *slog.Logger
	Client     NetworkClient
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
		err := h.Client.RegisterChat(update.Message.Chat.ID)
		switch {
		case errors.Is(err, ErrIncorrectRequestParameters):
			text = incorrectRequestParameters
			h.BaseLogger.Error("incorrect request parameters", slog.String("error", err.Error()))
		case errors.Is(err, ErrChatAlreadyExists):
			text = chatAlreadyExists
			h.BaseLogger.Error("chat already exists", slog.String("error", err.Error()))
		default:
			text = greetingMessage
			h.Session.SetState(update.Message.Chat.ID, statestorage.InitialState)
		}

	case "help":
		text = helpMessage
	case "track":
		text = trackMessage
		h.Session.SetState(update.Message.Chat.ID, statestorage.WaitingForTrackUrlState)
	case "untrack":
		text = untrackMessage
		h.Session.SetState(update.Message.Chat.ID, statestorage.WaitingForUnTrackUrlState)
	case "list":
		linksArr, err := h.Client.GetLinks(update.Message.Chat.ID)
		if err != nil {
			h.BaseLogger.Error("error getting links", slog.String("error", err.Error()),
				slog.Int64("chatId", update.Message.Chat.ID))
			return
		}
		h.handleLinks(update.Message.Chat.ID, linksArr)
		return
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

func (h TelegramHandler) handleLinks(chatID int64, links bot.ListLinksResponse) {
	if links.Size == 0 {
		if err := h.MsgSender.SendMessage(chatID, noTrackedLinks); err != nil {
			h.BaseLogger.Error("error while sending message", slog.String("error", err.Error()))
		}
	}

	for _, link := range links.Links {
		if err := h.MsgSender.SendMessage(chatID, link.Url); err != nil {
			h.BaseLogger.Error("error while sending message", slog.String("error", err.Error()))
		}
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
		//TODO: Сделать созранение ссылки в stateStore
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

func (h TelegramHandler) HandleLinkUpdate(linkUpdate shared.LinkUpdate) {
	for _, id := range linkUpdate.TgChatIds {
		if err := h.MsgSender.SendMessage(id, linkUpdate.Description+" '"+linkUpdate.Url); err != nil {
			h.BaseLogger.Error("error while sending message", slog.String("error", err.Error()))
		}
	}
}

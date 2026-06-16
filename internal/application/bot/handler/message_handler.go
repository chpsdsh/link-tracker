//go:generate mockgen -source message_handler.go -destination=../mocks/message_handler_mocks.go -package=mocks
package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/botmetrics"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/statestorage"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
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
	isAlreadyTracked           = "Ссылка уже отслеживается"
	chatNotFond                = "Для начала использования бота введите /start"
	notAuthorizedMessage       = "Данная команда недоступна. Для начала использования бота введите /start"
	linkNotFound               = "Данная ссылка не отслеживается"
	notURLToTrack              = "Вы ввели не ссылку, для добавления ссылки введить /track"
	notURLToUntrack            = "Вы ввели не ссылку, для удаления ссылки введить /untrack"
)

type TelegramBotHandler interface {
	HandleLinkUpdate(linkUpdate pkg.ProcessedLinkUpdate) error
}

type Sender interface {
	SendMessage(chatID int64, message string) error
}

type StateStorage interface {
	GetState(chatID int64) statestorage.State
	SetState(chatID int64, state statestorage.State)
	SetLinkAndUpdateState(chatID int64, link string, state statestorage.State)
	GetLink(chatID int64) string
	ClearLinkAndUpdateState(chatID int64, state statestorage.State)
}

type NetworkClient interface {
	RegisterChat(chatID int64) error
	GetLinks(chatID int64) (bot.ListLinksResponse, error)
	AddLink(chatID int64, linkRequest pkg.AddLinkRequest) (bot.LinkResponse, error)
	RemoveLink(chatID int64, removeRequest bot.RemoveLinkRequest) (bot.LinkResponse, error)
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

func (h TelegramHandler) HandleLinkUpdate(linkUpdate pkg.ProcessedLinkUpdate) error {
	for _, id := range linkUpdate.TgChatIDs {
		if err := h.MsgSender.SendMessage(id, linkUpdate.Description+"Приоритет: "+linkUpdate.Priority); err != nil {
			h.BaseLogger.Error("error while sending message", slog.String("error", err.Error()))
			return fmt.Errorf("telegram send message: %w", err)
		}
		botmetrics.SentNotificationsTotal.Inc()
	}
	return nil
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
		if h.Session.GetState(update.Message.Chat.ID) == statestorage.NoState {
			text = notAuthorizedMessage
		} else {
			text = trackMessage
			h.Session.SetState(update.Message.Chat.ID, statestorage.WaitingForTrackURLState)
		}
	case "untrack":
		if h.Session.GetState(update.Message.Chat.ID) == statestorage.NoState {
			text = notAuthorizedMessage
		} else {
			text = untrackMessage
			h.Session.SetState(update.Message.Chat.ID, statestorage.WaitingForUntrackURLState)
		}
	case "list":
		if h.Session.GetState(update.Message.Chat.ID) == statestorage.NoState {
			text = notAuthorizedMessage
		} else {
			linksArr, err := h.Client.GetLinks(update.Message.Chat.ID)
			if err != nil {
				h.BaseLogger.Error("error getting links", slog.String("error", err.Error()),
					slog.Int64("chatId", update.Message.Chat.ID))
				return
			}
			h.handleLinks(update.Message.Chat.ID, linksArr)
			return
		}
	case "cancel":
		if h.Session.GetState(update.Message.Chat.ID) == statestorage.NoState {
			text = notAuthorizedMessage
		} else {
			text = cancelMessage
			h.Session.ClearLinkAndUpdateState(update.Message.Chat.ID, statestorage.InitialState)
		}
	default:
		text = unknownMessage
	}

	if err := h.MsgSender.SendMessage(update.Message.Chat.ID, text); err != nil {
		h.BaseLogger.Error("error while sending message", slog.String("error", err.Error()))
	}
	botmetrics.CommandRequestTotal.WithLabelValues(text).Inc()
}

func (h TelegramHandler) handleLinks(chatID int64, links bot.ListLinksResponse) {
	if links.Size == 0 {
		if err := h.MsgSender.SendMessage(chatID, noTrackedLinks); err != nil {
			h.BaseLogger.Error("error while sending message", slog.String("error", err.Error()))
		}
		return
	}

	for _, link := range links.Links {
		if err := h.MsgSender.SendMessage(chatID, link.URL); err != nil {
			h.BaseLogger.Error("error while sending message", slog.String("error", err.Error()))
		}
	}
}

func (h TelegramHandler) handleMessage(update tgbotapi.Update) {
	var text string
	switch h.Session.GetState(update.Message.Chat.ID) {
	case statestorage.WaitingForTrackURLState:
		text = h.handleTrackURL(update)
	case statestorage.WaitingForTagsState:
		text = h.handleTrack(update)
		h.Session.ClearLinkAndUpdateState(update.Message.Chat.ID, statestorage.InitialState)
	case statestorage.WaitingForUntrackURLState:
		text = h.handleUntrack(update)
		h.Session.ClearLinkAndUpdateState(update.Message.Chat.ID, statestorage.InitialState)
	case statestorage.InitialState:
		text = unknownMessage
	case statestorage.NoState:
		text = unknownMessage
	default:
		text = unknownMessage
	}

	if err := h.MsgSender.SendMessage(update.Message.Chat.ID, text); err != nil {
		h.BaseLogger.Error("error while sending message", slog.String("error", err.Error()))
	}
}

func (h TelegramHandler) handleTrackURL(update tgbotapi.Update) string {
	var text string
	if _, err := url.ParseRequestURI(update.Message.Text); err != nil {
		h.Session.SetState(update.Message.Chat.ID, statestorage.InitialState)
		return notURLToTrack
	}
	h.Session.SetLinkAndUpdateState(update.Message.Chat.ID, update.Message.Text, statestorage.WaitingForTagsState)
	text = tagsMessage
	return text
}

func (h TelegramHandler) handleUntrack(update tgbotapi.Update) string {
	var text string
	if _, err := url.ParseRequestURI(update.Message.Text); err != nil {
		return notURLToUntrack
	}
	linkResp, err := h.Client.RemoveLink(update.Message.Chat.ID, bot.RemoveLinkRequest{Link: update.Message.Text})
	switch {
	case errors.Is(err, ErrIncorrectRequestParameters):
		text = incorrectRequestParameters
	case errors.Is(err, ErrChatNotFound):
		text = chatNotFond
	case errors.Is(err, ErrLinkNotExists):
		text = linkNotFound
	default:
		text = untrackConfirmMessage + " " + linkResp.URL + " " + tagsToString(linkResp.Tags)
	}
	return text
}

func (h TelegramHandler) handleTrack(update tgbotapi.Update) string {
	var text string
	tags := parseTags(update.Message.Text)
	linkResp, err := h.Client.AddLink(update.Message.Chat.ID, pkg.AddLinkRequest{Link: h.Session.GetLink(update.Message.Chat.ID), Tags: tags})
	switch {
	case errors.Is(err, ErrIncorrectRequestParameters):
		text = incorrectRequestParameters
	case errors.Is(err, ErrChatNotFound):
		text = chatNotFond
	case errors.Is(err, ErrLinkExists):
		text = isAlreadyTracked
	default:
		text = trackConfirmMessage + " " + linkResp.URL + " " + tagsToString(linkResp.Tags)
	}
	return text
}

func parseTags(tags string) []string {
	return strings.Split(tags, ",")
}

func tagsToString(tags []string) string {
	return strings.Join(tags, ",")
}

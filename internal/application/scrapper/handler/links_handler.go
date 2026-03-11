package handler

import (
	"errors"
	"log/slog"
	"net/url"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

var (
	ErrChatNotFound               = errors.New("chat not found")
	ErrLinkExists                 = errors.New("link already tracked")
	ErrIncorrectRequestParameters = errors.New("incorrect request parameters")
	ErrChatAlreadyExists          = errors.New("chat already exists")
	ErrLinkNotExists              = errors.New("link not exists")
)

type Repository interface {
	ChatExists(chatId int64) bool
	GetLinks(chatId int64) []shared.LinkInfo
	AddLink(chatId int64, link shared.LinkInfo)
	DeleteLink(chatId int64, link string) (shared.LinkInfo, bool)
	DeleteChat(chatId int64)
	AddChat(chatId int64)
	GetAllLinks() []shared.LinkInfo
	GetChatIdsByLink(link string) []int64
	UpdateLinksTime(newTime time.Time, linkToUpdate shared.LinkInfo)
}

type LinksHandler struct {
	Repo       Repository
	BaseLogger *slog.Logger
}

func (h LinksHandler) AddChatId(chatId int64) error {
	if h.Repo.ChatExists(chatId) {
		return ErrChatAlreadyExists
	}
	h.Repo.AddChat(chatId)
	return nil
}

func (h LinksHandler) DeleteChat(chatId int64) error {
	if !h.Repo.ChatExists(chatId) {
		return ErrChatNotFound
	}
	h.Repo.DeleteChat(chatId)
	return nil
}

func (h LinksHandler) GetLinks(chatId int64) ([]shared.LinkInfo, error) {
	if !h.Repo.ChatExists(chatId) {
		return nil, ErrChatNotFound
	}
	links := h.Repo.GetLinks(chatId)
	return links, nil
}

func (h LinksHandler) AddLink(chatId int64, linkRequest shared.AddLinkRequest) error {
	if !h.Repo.ChatExists(chatId) {
		return ErrChatNotFound
	}
	link := linkRequest.Link
	if _, err := url.Parse(linkRequest.Link); err != nil {
		return ErrIncorrectRequestParameters
	}

	links := h.Repo.GetLinks(chatId)
	for _, l := range links {
		if l.Link == link {
			return ErrLinkExists
		}
	}

	trackedLink := shared.LinkInfo{Link: link, Tags: linkRequest.Tags, LastUpdateTime: time.Now()}
	h.Repo.AddLink(chatId, trackedLink)

	return nil
}

func (h LinksHandler) DeleteLink(chatId int64, link string) (shared.LinkInfo, error) {
	if !h.Repo.ChatExists(chatId) {
		return shared.LinkInfo{}, ErrChatNotFound
	}
	linkInfo, ok := h.Repo.DeleteLink(chatId, link)
	if !ok {
		return shared.LinkInfo{}, ErrLinkNotExists
	}
	return linkInfo, nil
}

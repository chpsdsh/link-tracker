//go:generate mockgen -source links_handler.go -destination=../mocks/links_handler_mocks.go -package=mocks
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
	ChatExists(chatID int64) bool
	GetLinks(chatID int64) []shared.LinkInfo
	AddLink(chatID int64, link shared.LinkInfo)
	DeleteLink(chatID int64, link string) (shared.LinkInfo, bool)
	DeleteChat(chatID int64)
	AddChat(chatID int64)
	GetAllLinks() []shared.LinkInfo
	GetChatIDsByLink(link string) []int64
	UpdateLinksTime(newTime time.Time, linkToUpdate shared.LinkInfo)
}

type LinksHandler struct {
	Repo       Repository
	BaseLogger *slog.Logger
}

func (h LinksHandler) AddChatID(chatID int64) error {
	if h.Repo.ChatExists(chatID) {
		return ErrChatAlreadyExists
	}
	h.Repo.AddChat(chatID)
	return nil
}

func (h LinksHandler) DeleteChat(chatID int64) error {
	if !h.Repo.ChatExists(chatID) {
		return ErrChatNotFound
	}
	h.Repo.DeleteChat(chatID)
	return nil
}

func (h LinksHandler) GetLinks(chatID int64) ([]shared.LinkInfo, error) {
	if !h.Repo.ChatExists(chatID) {
		return nil, ErrChatNotFound
	}
	links := h.Repo.GetLinks(chatID)
	return links, nil
}

func (h LinksHandler) AddLink(chatID int64, linkRequest shared.AddLinkRequest) error {
	if !h.Repo.ChatExists(chatID) {
		return ErrChatNotFound
	}
	link := linkRequest.Link
	if _, err := url.Parse(linkRequest.Link); err != nil {
		return ErrIncorrectRequestParameters
	}

	links := h.Repo.GetLinks(chatID)
	for _, l := range links {
		if l.Link == link {
			return ErrLinkExists
		}
	}

	trackedLink := shared.LinkInfo{Link: link, Tags: linkRequest.Tags, LastUpdateTime: time.Now()}
	h.Repo.AddLink(chatID, trackedLink)

	return nil
}

func (h LinksHandler) DeleteLink(chatID int64, link string) (shared.LinkInfo, error) {
	if !h.Repo.ChatExists(chatID) {
		return shared.LinkInfo{}, ErrChatNotFound
	}
	linkInfo, ok := h.Repo.DeleteLink(chatID, link)
	if !ok {
		return shared.LinkInfo{}, ErrLinkNotExists
	}
	return linkInfo, nil
}

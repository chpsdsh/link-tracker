package handler

import (
	"net/http"
	"net/url"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

type Repository interface {
	ChatExists(chatId int64) bool
	GetLinks(chatId int64) []shared.TrackedLink
	AddLink(chatId int64, link shared.TrackedLink)
	DeleteLink(chatId int64, link string)
}

type LinksHandler struct {
	Repo Repository
}

func (s LinksHandler) AddLink(chatId int64, linkRequest scrapper.AddLinkRequest) int {
	if !s.Repo.ChatExists(chatId) {
		return http.StatusNotFound
	}
	link := *linkRequest.Link
	if _, err := url.Parse(*linkRequest.Link); err != nil {
		return http.StatusBadRequest
	}

	links := s.Repo.GetLinks(chatId)
	for _, l := range links {
		if l.Link == link {
			return http.StatusConflict
		}
	}

	trackedLink := shared.TrackedLink{Link: link, Tags: *linkRequest.Tags}
	s.Repo.AddLink(chatId, trackedLink)

	return http.StatusOK
}

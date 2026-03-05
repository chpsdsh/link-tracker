package scrapperserver

import (
	"net/http"
	"net/url"
)

type Repository interface {
	GetLinks(chatId int64) []url.URL
	AddLink(chatId int64, link url.URL)
	DeleteLink(chatId int64, link url.URL)
}

type Scrapper struct {
	repo Repository
}

func (s Scrapper) DeleteLinks(w http.ResponseWriter, r *http.Request, params DeleteLinksParams) {

}

func (s Scrapper) GetLinks(w http.ResponseWriter, r *http.Request, params GetLinksParams) {

}

func (s Scrapper) PostLinks(w http.ResponseWriter, r *http.Request, params PostLinksParams) {

}

func (s Scrapper) DeleteTgChatId(w http.ResponseWriter, r *http.Request, id int64) {

}

func (s Scrapper) PostTgChatId(w http.ResponseWriter, r *http.Request, id int64) {

}

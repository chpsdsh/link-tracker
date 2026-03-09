package botclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

const (
	scrapperUrl         = "http://localhost:8081"
	tgHeaderKey         = "Tg-Chat-Id"
	contentTypeKey      = "Content-Type"
	typeApplicationJSON = "application/json"
)

type Client struct {
	Client *http.Client
}

func (c Client) doRequest(chatId int64, method, url string) (*http.Response, error) {
	stringId := strconv.FormatInt(chatId, 10)
	req, err := http.NewRequest(method, url+stringId, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c Client) doRequestWithHeader(chatId int64, method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	stringId := strconv.FormatInt(chatId, 10)
	req.Header.Set(tgHeaderKey, stringId)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c Client) RegisterChat(chatId int64) error {
	resp, err := c.doRequest(chatId, http.MethodPost, scrapperUrl+"/tg-chat/")
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return handler.ErrIncorrectRequestParameters
	case http.StatusConflict:
		return handler.ErrChatAlreadyExists
	}

	return nil
}

func (c Client) UnregisterChat(chatId int64) error {
	resp, err := c.doRequest(chatId, http.MethodDelete, scrapperUrl+"/tg-chat/")
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return handler.ErrIncorrectRequestParameters
	case http.StatusNotFound:
		return handler.ErrChatNotFound
	}

	return nil
}

func (c Client) GetLinks(chatId int64) (bot.ListLinksResponse, error) {
	resp, err := c.doRequestWithHeader(chatId, http.MethodGet, scrapperUrl+"/links")
	if err != nil {
		return bot.ListLinksResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return bot.ListLinksResponse{}, handler.ErrIncorrectRequestParameters
	case http.StatusNotFound:
		return bot.ListLinksResponse{}, handler.ErrChatNotFound
	default:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return bot.ListLinksResponse{}, err
		}
		linksResponse := bot.ListLinksResponse{}
		if err := json.Unmarshal(data, &linksResponse); err != nil {
			return bot.ListLinksResponse{}, err
		}
		return linksResponse, nil
	}
}

func (c Client) AddLink(chatId int64, linkRequest shared.AddLinkRequest) (bot.LinkResponse, error) {
	data, err := json.Marshal(linkRequest)
	if err != nil {
		return bot.LinkResponse{}, err
	}
	req, err := http.NewRequest(http.MethodPost, scrapperUrl+"/links", bytes.NewBuffer(data))
	if err != nil {
		return bot.LinkResponse{}, err
	}

	stringId := strconv.FormatInt(chatId, 10)
	req.Header.Set(tgHeaderKey, stringId)
	req.Header.Set(contentTypeKey, typeApplicationJSON)

	resp, err := c.Client.Do(req)
	if err != nil {
		return bot.LinkResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusBadRequest:
		return bot.LinkResponse{}, handler.ErrIncorrectRequestParameters
	case http.StatusNotFound:
		return bot.LinkResponse{}, handler.ErrChatNotFound
	case http.StatusConflict:
		return bot.LinkResponse{}, handler.ErrLinkExists
	default:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return bot.LinkResponse{}, err
		}
		linksResponse := bot.LinkResponse{}
		if err := json.Unmarshal(data, &linksResponse); err != nil {
			return bot.LinkResponse{}, err
		}
		return linksResponse, nil
	}
}

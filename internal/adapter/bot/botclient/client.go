package botclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

const (
	tgHeaderKey         = "Tg-Chat-ID"
	contentTypeKey      = "Content-Type"
	typeApplicationJSON = "application/json"
)

type Client struct {
	Client *http.Client
	Config config.Config
}

func (c Client) RegisterChat(chatID int64) error {
	resp, err := c.doRequest(chatID, http.MethodPost, c.Config.ScrapperServerAddress+"/tg-chat/")
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

func (c Client) UnregisterChat(chatID int64) error {
	resp, err := c.doRequest(chatID, http.MethodDelete, c.Config.ScrapperServerAddress+"/tg-chat/")
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

func (c Client) GetLinks(chatID int64) (bot.ListLinksResponse, error) {
	resp, err := c.doRequestWithHeader(chatID, http.MethodGet, c.Config.ScrapperServerAddress+"/links")
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
		data, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			return bot.ListLinksResponse{}, fmt.Errorf("error reading response body: %w", errRead)
		}
		linksResponse := bot.ListLinksResponse{}
		if errUnmarshall := json.Unmarshal(data, &linksResponse); errUnmarshall != nil {
			return bot.ListLinksResponse{}, fmt.Errorf("error unmarshalling JSON: %w", errUnmarshall)
		}
		return linksResponse, nil
	}
}

func (c Client) AddLink(chatID int64, linkRequest shared.AddLinkRequest) (bot.LinkResponse, error) {
	data, err := json.Marshal(linkRequest)
	if err != nil {
		return bot.LinkResponse{}, fmt.Errorf("error marshalling JSON: %w", err)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.Config.ScrapperServerAddress+"/links", bytes.NewBuffer(data))
	if err != nil {
		return bot.LinkResponse{}, fmt.Errorf("error creating request: %w", err)
	}

	stringID := strconv.FormatInt(chatID, 10)
	req.Header.Set(tgHeaderKey, stringID)
	req.Header.Set(contentTypeKey, typeApplicationJSON)

	resp, err := c.Client.Do(req)
	if err != nil {
		return bot.LinkResponse{}, fmt.Errorf("error doing request: %w", err)
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
		dataRead, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			return bot.LinkResponse{}, fmt.Errorf("error reading response: %w", errRead)
		}
		linksResponse := bot.LinkResponse{}
		if errUnmarshall := json.Unmarshal(dataRead, &linksResponse); errUnmarshall != nil {
			return bot.LinkResponse{}, fmt.Errorf("error unmarshalling JSON: %w", errUnmarshall)
		}
		return linksResponse, nil
	}
}

func (c Client) RemoveLink(chatID int64, removeRequest bot.RemoveLinkRequest) (bot.LinkResponse, error) {
	data, err := json.Marshal(removeRequest)
	if err != nil {
		return bot.LinkResponse{}, fmt.Errorf("error marshalling JSON: %w", err)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, c.Config.ScrapperServerAddress+"/links", bytes.NewBuffer(data))
	if err != nil {
		return bot.LinkResponse{}, fmt.Errorf("error creating request: %w", err)
	}

	stringID := strconv.FormatInt(chatID, 10)
	req.Header.Set(tgHeaderKey, stringID)
	req.Header.Set(contentTypeKey, typeApplicationJSON)

	resp, err := c.Client.Do(req)
	if err != nil {
		return bot.LinkResponse{}, fmt.Errorf("error doing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return bot.LinkResponse{}, handler.ErrIncorrectRequestParameters
	case http.StatusNotFound:
		return bot.LinkResponse{}, handler.ErrLinkNotExists
	default:
		respData, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			return bot.LinkResponse{}, fmt.Errorf("error reading response: %w", errRead)
		}
		linksResponse := bot.LinkResponse{}
		if errUnmarshall := json.Unmarshal(respData, &linksResponse); errUnmarshall != nil {
			return bot.LinkResponse{}, fmt.Errorf("error unmarshalling JSON: %w", errUnmarshall)
		}
		return linksResponse, nil
	}
}

func (c Client) doRequest(chatID int64, method, url string) (*http.Response, error) {
	stringID := strconv.FormatInt(chatID, 10)
	req, err := http.NewRequestWithContext(context.Background(), method, url+stringID, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}

	return resp, nil
}

func (c Client) doRequestWithHeader(chatID int64, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	stringID := strconv.FormatInt(chatID, 10)
	req.Header.Set(tgHeaderKey, stringID)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}

	return resp, nil
}

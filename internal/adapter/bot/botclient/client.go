package botclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/sony/gobreaker"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/botmetrics"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

const (
	tgHeaderKey         = "Tg-Chat-ID"
	contentTypeKey      = "Content-Type"
	typeApplicationJSON = "application/json"
	botBreakerNameKey   = "bot-breaker"
	httpClientScope     = "http_client"
)

var ErrIncorrectCastType = errors.New("incorrect type cast")

type Client struct {
	Client  *http.Client
	Config  config.BotConfig
	Retrier *retry.Retrier
	Breaker *gobreaker.CircuitBreaker
}

func NewBotClient(conf config.BotConfig) Client {
	client := &http.Client{Timeout: conf.HTTPClientConfig.Timeout}
	retrier := retry.New(
		retry.Attempts(conf.RetryConfig.MaxAttempts),
		retry.Delay(conf.RetryConfig.Delay),
		retry.DelayType(retry.FixedDelay),
	)
	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        botBreakerNameKey,
		MaxRequests: conf.CircuitBreakerConfig.MaxRequests,
		Interval:    conf.CircuitBreakerConfig.Interval,
		Timeout:     conf.CircuitBreakerConfig.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			total := counts.Requests
			if total == 0 {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(total)
			return failureRatio >= conf.CircuitBreakerConfig.FailureRatio
		},
	})
	return Client{Client: client, Config: conf, Retrier: retrier, Breaker: breaker}
}

func (c Client) RegisterChat(chatID int64) error {
	botmetrics.CommandRequestTotal.WithLabelValues("register_chat").Inc()
	httpResult, err := c.doRequestWithMethodAndRetry(chatID, http.MethodPost, c.Config.ScrapperServerAddress+"/tg-chat/", c.doRequest)
	if err != nil {
		botmetrics.ErrorsCounterTotal.WithLabelValues(httpClientScope, "register_chat").Inc()
		return fmt.Errorf("error doing request: %w", err)
	}
	switch httpResult.StatusCode {
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
	botmetrics.CommandRequestTotal.WithLabelValues("unregister_chat").Inc()
	startTime := time.Now()
	httpResult, err := c.doRequestWithMethodAndRetry(chatID, http.MethodDelete, c.Config.ScrapperServerAddress+"/tg-chat/", c.doRequest)
	botmetrics.ObserveCommandDuration(startTime, httpClientScope, "unregister_chat")

	if err != nil {
		botmetrics.ErrorsCounterTotal.WithLabelValues(httpClientScope, "unregister_chat").Inc()
		return fmt.Errorf("error doing request: %w", err)
	}

	switch httpResult.StatusCode {
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
	botmetrics.CommandRequestTotal.WithLabelValues("get_links").Inc()
	startTime := time.Now()
	httpResult, err := c.doRequestWithMethodAndRetry(chatID, http.MethodGet, c.Config.ScrapperServerAddress+"/links", c.doRequestWithHeader)
	botmetrics.ObserveCommandDuration(startTime, httpClientScope, "get_links")
	if err != nil {
		botmetrics.ErrorsCounterTotal.WithLabelValues(httpClientScope, "get_links").Inc()
		return bot.ListLinksResponse{}, fmt.Errorf("error doing request: %w", err)
	}

	switch httpResult.StatusCode {
	case http.StatusBadRequest:
		return bot.ListLinksResponse{}, handler.ErrIncorrectRequestParameters
	case http.StatusNotFound:
		return bot.ListLinksResponse{}, handler.ErrChatNotFound
	default:
		linksResponse := bot.ListLinksResponse{}
		if errUnmarshall := json.Unmarshal(httpResult.Body, &linksResponse); errUnmarshall != nil {
			return bot.ListLinksResponse{}, fmt.Errorf("error unmarshalling JSON: %w", errUnmarshall)
		}
		return linksResponse, nil
	}
}

func (c Client) AddLink(chatID int64, linkRequest pkg.AddLinkRequest) (bot.LinkResponse, error) {
	botmetrics.CommandRequestTotal.WithLabelValues("add_link").Inc()
	linkRequestJSON, err := json.Marshal(linkRequest)
	if err != nil {
		return bot.LinkResponse{}, fmt.Errorf("error marshalling JSON: %w", err)
	}

	startTime := time.Now()
	httpResult, err := c.doRequestWithRetry(chatID, linkRequestJSON, c.doAddLinkRequest)
	botmetrics.ObserveCommandDuration(startTime, httpClientScope, "add_link")

	if err != nil {
		botmetrics.ErrorsCounterTotal.WithLabelValues(httpClientScope, "add_link").Inc()
		return bot.LinkResponse{}, fmt.Errorf("error adding link: %w", err)
	}

	switch httpResult.StatusCode {
	case http.StatusBadRequest:
		return bot.LinkResponse{}, handler.ErrIncorrectRequestParameters
	case http.StatusNotFound:
		return bot.LinkResponse{}, handler.ErrChatNotFound
	case http.StatusConflict:
		return bot.LinkResponse{}, handler.ErrLinkExists
	default:
		linksResponse := bot.LinkResponse{}
		if errUnmarshall := json.Unmarshal(httpResult.Body, &linksResponse); errUnmarshall != nil {
			return bot.LinkResponse{}, fmt.Errorf("error unmarshalling JSON: %w", errUnmarshall)
		}
		return linksResponse, nil
	}
}

func (c Client) RemoveLink(chatID int64, removeRequest bot.RemoveLinkRequest) (bot.LinkResponse, error) {
	botmetrics.CommandRequestTotal.WithLabelValues("remove_link").Inc()
	removeLinkJSON, err := json.Marshal(removeRequest)
	if err != nil {
		return bot.LinkResponse{}, fmt.Errorf("error marshalling JSON: %w", err)
	}

	startTime := time.Now()
	httpResult, err := c.doRequestWithRetry(chatID, removeLinkJSON, c.doRemoveLinkRequest)
	botmetrics.ObserveCommandDuration(startTime, httpClientScope, "remove_link")

	if err != nil {
		botmetrics.ErrorsCounterTotal.WithLabelValues(httpClientScope, "remove_link").Inc()
		return bot.LinkResponse{}, err
	}

	switch httpResult.StatusCode {
	case http.StatusBadRequest:
		return bot.LinkResponse{}, handler.ErrIncorrectRequestParameters
	case http.StatusNotFound:
		return bot.LinkResponse{}, handler.ErrLinkNotExists
	default:
		linksResponse := bot.LinkResponse{}
		if errUnmarshall := json.Unmarshal(httpResult.Body, &linksResponse); errUnmarshall != nil {
			return bot.LinkResponse{}, fmt.Errorf("error unmarshalling JSON: %w", errUnmarshall)
		}
		return linksResponse, nil
	}
}

func (c Client) doAddLinkRequest(chatID int64, data []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.Config.ScrapperServerAddress+"/links", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	stringID := strconv.FormatInt(chatID, 10)
	req.Header.Set(tgHeaderKey, stringID)
	req.Header.Set(contentTypeKey, typeApplicationJSON)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}

	return resp, nil
}

func (c Client) doRequestWithMethodAndRetry(chatID int64, method, url string, requestFunc func(chatID int64, method, url string) (*http.Response, error)) (pkg.HTTPResult, error) {
	result, err := c.Breaker.Execute(func() (any, error) {
		var httpResult pkg.HTTPResult
		retryErr := c.Retrier.Do(func() error {
			resp, reqErr := requestFunc(chatID, method, url)
			if reqErr != nil {
				if resp != nil && resp.Body != nil {
					_ = resp.Body.Close()
				}
				return fmt.Errorf("error sending request: %w", reqErr)
			}

			defer func() { _ = resp.Body.Close() }()

			if c.isRetryableStatus(resp.StatusCode) {
				return fmt.Errorf("retriable status code %d", resp.StatusCode)
			}

			data, errRead := io.ReadAll(resp.Body)
			if errRead != nil {
				return fmt.Errorf("error reading response body: %w", errRead)
			}

			httpResult.StatusCode = resp.StatusCode
			httpResult.Body = data
			return nil
		})

		if retryErr != nil {
			return pkg.HTTPResult{}, fmt.Errorf("all retries returned errors: %w", retryErr)
		}
		return httpResult, nil
	})
	if err != nil {
		return pkg.HTTPResult{}, fmt.Errorf("error doing request: %w", err)
	}

	httpResult, ok := result.(pkg.HTTPResult)
	if !ok {
		return pkg.HTTPResult{}, ErrIncorrectCastType
	}
	return httpResult, nil
}

func (c Client) doRequestWithRetry(chatID int64, dataJSON []byte, requestFunc func(chatID int64, data []byte) (*http.Response, error)) (pkg.HTTPResult, error) {
	result, err := c.Breaker.Execute(func() (any, error) {
		var httpResult pkg.HTTPResult
		retryErr := c.Retrier.Do(func() error {
			resp, reqErr := requestFunc(chatID, dataJSON)
			if reqErr != nil {
				if resp != nil && resp.Body != nil {
					_ = resp.Body.Close()
				}
				return fmt.Errorf("error sending request: %w", reqErr)
			}

			defer func() { _ = resp.Body.Close() }()

			if c.isRetryableStatus(resp.StatusCode) {
				return fmt.Errorf("retriable status code %d", resp.StatusCode)
			}

			data, errRead := io.ReadAll(resp.Body)
			if errRead != nil {
				return fmt.Errorf("error reading response body: %w", errRead)
			}

			httpResult.StatusCode = resp.StatusCode
			httpResult.Body = data
			return nil
		})

		if retryErr != nil {
			return pkg.HTTPResult{}, fmt.Errorf("all retries returned errors: %w", retryErr)
		}
		return httpResult, nil
	})

	if err != nil {
		return pkg.HTTPResult{}, fmt.Errorf("error doing request: %w", err)
	}

	httpResult, ok := result.(pkg.HTTPResult)
	if !ok {
		return pkg.HTTPResult{}, ErrIncorrectCastType
	}
	return httpResult, nil
}

func (c Client) doRemoveLinkRequest(chatID int64, data []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, c.Config.ScrapperServerAddress+"/links", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	stringID := strconv.FormatInt(chatID, 10)
	req.Header.Set(tgHeaderKey, stringID)
	req.Header.Set(contentTypeKey, typeApplicationJSON)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}
	return resp, nil
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
func (c Client) isRetryableStatus(status int) bool {
	for _, retryableStatus := range c.Config.RetryConfig.RetryableStatuses {
		if status == retryableStatus {
			return true
		}
	}
	return false
}

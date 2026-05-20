package scrapperclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/avast/retry-go/v5"
	"github.com/goccy/go-json"
	"github.com/sony/gobreaker"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

const updatesBreakerName = "updates-breaker"

type UpdatesClient struct {
	Client  *http.Client
	Config  config.Config
	Breaker *gobreaker.CircuitBreaker
	Retrier *retry.Retrier
}

func NewUpdatesClient(conf config.Config) *UpdatesClient {
	client := &http.Client{Timeout: conf.HTTPClientConfig.Timeout}
	retrier := retry.New(
		retry.Attempts(conf.RetryConfig.MaxAttempts),
		retry.Delay(conf.RetryConfig.Delay),
		retry.DelayType(retry.FixedDelay),
	)
	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        updatesBreakerName,
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
	return &UpdatesClient{Client: client, Config: conf, Retrier: retrier, Breaker: breaker}
}

func (c UpdatesClient) SendLinkUpdate(update pkg.LinkUpdate, _ string) error {
	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	result, err := c.Breaker.Execute(func() (any, error) {
		var status int
		retryErr := c.Retrier.Do(func() error {
			resp, reqErr := c.doLinkUpdateRequest(data)
			if reqErr != nil {
				if resp != nil && resp.Body != nil {
					_ = resp.Body.Close()
				}
				return fmt.Errorf("error sending link update: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if c.isRetryableStatus(resp.StatusCode) {
				return fmt.Errorf("retriable status code %d", resp.StatusCode)
			}
			status = resp.StatusCode
			return nil
		})
		if retryErr != nil {
			return nil, fmt.Errorf("all retries failed: %w", retryErr)
		}

		return status, nil
	})
	if err != nil {
		return fmt.Errorf("error doing request: %w", err)
	}

	status, ok := result.(int)
	if !ok {
		return fmt.Errorf("error retrieving response: %w", err)
	}
	switch status {
	case http.StatusOK:
		return nil
	default:
		return scrapper.ErrIncorrectRequestParameters
	}
}

func (c UpdatesClient) Close() {
	c.Client.CloseIdleConnections()
}

func (c UpdatesClient) doLinkUpdateRequest(data []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.Config.BotServerAddr+botPostEndpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Add(contentTypeKey, typeApplicationJSON)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	return resp, nil
}

func (c UpdatesClient) isRetryableStatus(status int) bool {
	for _, retryableStatus := range c.Config.RetryConfig.RetryableStatuses {
		if status == retryableStatus {
			return true
		}
	}
	return false
}

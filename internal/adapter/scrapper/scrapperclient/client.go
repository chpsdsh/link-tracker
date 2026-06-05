package scrapperclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/sony/gobreaker"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/scrappermetrics"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

var ErrUnmarshallingJSON = errors.New("json unmarshalling error")

const (
	applicationType     = "application/vnd.github.v3+json"
	version             = "2026-03-10"
	stackOverflowKey    = "&key="
	gitBreakerName      = "git-breaker"
	stackBreakerName    = "stack-overflow-breaker"
	stackOverflowSource = "stackOverflow"
	githubSource        = "github"
	httpClientSource    = "http_client"
)

type Client struct {
	Client               *http.Client
	Config               config.ScrapperConfig
	Retrier              *retry.Retrier
	GithubBreaker        *gobreaker.CircuitBreaker
	StackOverflowBreaker *gobreaker.CircuitBreaker
}

func NewScrapperClient(conf config.ScrapperConfig) Client {
	client := &http.Client{Timeout: conf.HTTPClientConfig.Timeout}
	retrier := retry.New(
		retry.Attempts(conf.RetryConfig.MaxAttempts),
		retry.Delay(conf.RetryConfig.Delay),
		retry.DelayType(retry.FixedDelay),
	)
	gitBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        gitBreakerName,
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
	stackOverflowBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        stackBreakerName,
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
	return Client{Client: client, Config: conf, Retrier: retrier, GithubBreaker: gitBreaker, StackOverflowBreaker: stackOverflowBreaker}
}

func (c Client) DoGithubRequest(url string) (scrapper.GitHubRepositoryResponse, error) {
	startTime := time.Now()
	body, err := c.doGithubRequest(url, c.doGithubAPIRequest)
	scrappermetrics.ObserveRequestDuration(startTime, httpClientSource, githubSource)
	if err != nil {
		scrappermetrics.APIErrorsTotal.WithLabelValues(githubSource).Inc()
		return scrapper.GitHubRepositoryResponse{}, fmt.Errorf("github request error: %w", err)
	}

	scrappermetrics.APIRequestsTotal.WithLabelValues(githubSource).Inc()

	gitUpdate := scrapper.GitHubRepositoryResponse{}
	if err = json.Unmarshal(body, &gitUpdate); err != nil {
		return scrapper.GitHubRepositoryResponse{}, errors.Join(err, ErrUnmarshallingJSON)
	}
	return gitUpdate, nil
}

func (c Client) DoGithubIssueRequest(url string) ([]scrapper.GithubIssue, error) {
	startTime := time.Now()
	body, err := c.doGithubRequest(url, c.doGithubAPIRequest)
	scrappermetrics.ObserveRequestDuration(startTime, httpClientSource, githubSource)
	if err != nil {
		scrappermetrics.APIErrorsTotal.WithLabelValues(githubSource).Inc()
		return nil, fmt.Errorf("error doing github api request: %w", err)
	}

	scrappermetrics.APIRequestsTotal.WithLabelValues(githubSource).Inc()

	var issueResponse []scrapper.GithubIssue
	if err = json.Unmarshal(body, &issueResponse); err != nil {
		return nil, errors.Join(err, ErrUnmarshallingJSON)
	}
	return issueResponse, nil
}

func (c Client) DoGithubPullRequestRequest(url string) ([]scrapper.GithubPullRequest, error) {
	startTime := time.Now()
	body, err := c.doGithubRequest(url, c.doGithubAPIRequest)
	scrappermetrics.ObserveRequestDuration(startTime, httpClientSource, githubSource)
	if err != nil {
		scrappermetrics.APIErrorsTotal.WithLabelValues(githubSource).Inc()
		return nil, fmt.Errorf("error requesting github API: %w", err)
	}

	scrappermetrics.APIRequestsTotal.WithLabelValues(githubSource).Inc()

	var pullRequestsResponse []scrapper.GithubPullRequest
	if err = json.Unmarshal(body, &pullRequestsResponse); err != nil {
		return nil, errors.Join(err, ErrUnmarshallingJSON)
	}
	return pullRequestsResponse, nil
}

func (c Client) DoStackOverflowQuestionRequest(url string) (scrapper.StackOverflowQuestionResponse, error) {
	startTime := time.Now()
	body, err := c.doStackOverflowRequest(url, c.doStackoverflowAPIRequest)
	scrappermetrics.ObserveRequestDuration(startTime, httpClientSource, stackOverflowSource)
	if err != nil {
		scrappermetrics.APIErrorsTotal.WithLabelValues(stackOverflowSource).Inc()
		return scrapper.StackOverflowQuestionResponse{}, fmt.Errorf("error requesting stackOverflow API: %w", err)
	}

	scrappermetrics.APIRequestsTotal.WithLabelValues(stackOverflowSource).Inc()

	stackOverflowUpdate := scrapper.StackOverflowQuestionResponse{}
	if err = json.Unmarshal(body, &stackOverflowUpdate); err != nil {
		return scrapper.StackOverflowQuestionResponse{}, errors.Join(err, ErrUnmarshallingJSON)
	}

	return stackOverflowUpdate, nil
}

func (c Client) DoStackOverflowAnswersRequest(url string) (scrapper.StackOverflowAnswersResponse, error) {
	startTime := time.Now()
	body, err := c.doStackOverflowRequest(url, c.doStackoverflowAPIRequest)
	scrappermetrics.ObserveRequestDuration(startTime, httpClientSource, stackOverflowSource)

	if err != nil {
		scrappermetrics.APIErrorsTotal.WithLabelValues(stackOverflowSource).Inc()
		return scrapper.StackOverflowAnswersResponse{}, fmt.Errorf("error requesting stackOverflow API: %w", err)
	}
	scrappermetrics.APIRequestsTotal.WithLabelValues(stackOverflowSource).Inc()
	stackOverflowAnswers := scrapper.StackOverflowAnswersResponse{}
	if err = json.Unmarshal(body, &stackOverflowAnswers); err != nil {
		return scrapper.StackOverflowAnswersResponse{}, errors.Join(err, ErrUnmarshallingJSON)
	}

	return stackOverflowAnswers, nil
}

func (c Client) DoStackOverflowCommentsRequest(url string) (scrapper.StackOverflowCommentsResponse, error) {
	startTime := time.Now()
	body, err := c.doStackOverflowRequest(url, c.doStackoverflowAPIRequest)
	scrappermetrics.ObserveRequestDuration(startTime, httpClientSource, stackOverflowSource)

	if err != nil {
		scrappermetrics.APIErrorsTotal.WithLabelValues(stackOverflowSource).Inc()
		return scrapper.StackOverflowCommentsResponse{}, fmt.Errorf("error requesting stackOverflow API: %w", err)
	}

	scrappermetrics.APIRequestsTotal.WithLabelValues(stackOverflowSource).Inc()

	stackOverflowComments := scrapper.StackOverflowCommentsResponse{}
	if err = json.Unmarshal(body, &stackOverflowComments); err != nil {
		return scrapper.StackOverflowCommentsResponse{}, errors.Join(err, ErrUnmarshallingJSON)
	}

	return stackOverflowComments, nil
}

func (c Client) doGithubRequest(url string, queryFunc func(url string) (*http.Response, error)) ([]byte, error) {
	result, err := c.GithubBreaker.Execute(func() (any, error) {
		return c.doRequestWithRetry(url, queryFunc)
	})
	if err != nil {
		return nil, fmt.Errorf("github circuit breaker request failed: %w", err)
	}

	resp, ok := result.([]byte)
	if !ok {
		return nil, errors.New("invalid circuit breaker response type")
	}
	return resp, nil
}

func (c Client) doStackOverflowRequest(url string, queryFunc func(url string) (*http.Response, error)) ([]byte, error) {
	result, err := c.StackOverflowBreaker.Execute(func() (any, error) {
		return c.doRequestWithRetry(url, queryFunc)
	})
	if err != nil {
		return nil, fmt.Errorf("stack overflow circuit breaker request failed: %w", err)
	}

	body, ok := result.([]byte)
	if !ok {
		return nil, errors.New("invalid circuit breaker response type")
	}
	return body, nil
}

func (c Client) doRequestWithRetry(url string, queryFunc func(url string) (*http.Response, error)) ([]byte, error) {
	var body []byte

	err := c.Retrier.Do(func() error {
		resp, err := queryFunc(url)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if c.isRetryableStatus(resp.StatusCode) {
			return fmt.Errorf("retriable status code %d", resp.StatusCode)
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return retry.Unrecoverable(
				fmt.Errorf("unexpected status code %d", resp.StatusCode),
			)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return retry.Unrecoverable(
				fmt.Errorf("error reading response body: %w", err),
			)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("request with retry failed: %w", err)
	}

	return body, nil
}

func (c Client) doGithubAPIRequest(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Accept", applicationType)
	req.Header.Add("Authorization", c.Config.GithubToken)
	req.Header.Add("X-GitHub-Api-Version", version)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}
	return resp, nil
}

func (c Client) doStackoverflowAPIRequest(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url+stackOverflowKey+c.Config.StackoverflowToken, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

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

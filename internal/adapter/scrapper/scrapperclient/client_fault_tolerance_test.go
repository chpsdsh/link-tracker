package scrapperclient

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	config2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
)

const githubRepoResponse = `{
	"id": 1,
	"name": "go",
	"full_name": "golang/go",
	"updated_at": "2026-01-01T00:00:00Z"
}`

func testScrapperConfig(timeout time.Duration) config.ScrapperConfig {
	return config.ScrapperConfig{
		GithubToken:        "test-github-token",
		StackoverflowToken: "test-stack-token",
		HTTPClientConfig: config2.HTTPClientConfig{
			Timeout: timeout,
		},
		RetryConfig: config2.RetryConfig{
			MaxAttempts:       3,
			Delay:             50 * time.Millisecond,
			RetryableStatuses: []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout},
		},
		CircuitBreakerConfig: config2.CircuitBreakerConfig{
			Interval:     time.Second,
			Timeout:      100 * time.Millisecond,
			MaxRequests:  2,
			FailureRatio: 0.5,
		},
	}
}

func callGithub(client Client, url string) error {
	_, err := client.DoGithubRequest(url)
	return err
}

func TestClient_DoGithubRequest_Timeout(t *testing.T) {
	serverDelay := 300 * time.Millisecond
	clientTimeout := 50 * time.Millisecond

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(serverDelay)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(githubRepoResponse))
	}))
	defer server.Close()

	conf := testScrapperConfig(clientTimeout)
	conf.RetryConfig.MaxAttempts = 1

	client := NewScrapperClient(conf)

	start := time.Now()
	_, err := client.DoGithubRequest(server.URL)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, serverDelay)
}

func TestClient_DoGithubRequest_RetryOn5xx(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		call := calls.Add(1)

		switch call {
		case 1, 2:
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(githubRepoResponse))
		}
	}))
	defer server.Close()

	conf := testScrapperConfig(time.Second)
	conf.RetryConfig.MaxAttempts = 3
	conf.RetryConfig.Delay = 10 * time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	client := NewScrapperClient(conf)

	_, err := client.DoGithubRequest(server.URL)

	require.NoError(t, err)
	assert.Equal(t, int32(3), calls.Load())
}

func TestClient_DoGithubRequest_NoRetryOn4xx(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"bad request"}`))
	}))
	defer server.Close()

	conf := testScrapperConfig(time.Second)
	conf.RetryConfig.MaxAttempts = 3
	conf.RetryConfig.Delay = 10 * time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	client := NewScrapperClient(conf)

	_, err := client.DoGithubRequest(server.URL)

	require.Error(t, err)
	assert.Equal(t, int32(1), calls.Load())
	assert.Contains(t, err.Error(), "unexpected status code 400")
}

func TestClient_DoGithubRequest_UsesConstantBackoff(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		call := calls.Add(1)

		if call < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(githubRepoResponse))
	}))
	defer server.Close()

	retryDelay := 80 * time.Millisecond

	conf := testScrapperConfig(time.Second)
	conf.RetryConfig.MaxAttempts = 3
	conf.RetryConfig.Delay = retryDelay
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	client := NewScrapperClient(conf)

	start := time.Now()
	_, err := client.DoGithubRequest(server.URL)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, int32(3), calls.Load())

	assert.GreaterOrEqual(t, elapsed, 2*retryDelay)
	assert.Less(t, elapsed, 2*retryDelay+200*time.Millisecond)
}

func TestClient_DoGithubRequest_CircuitBreakerGoesOpen(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	conf := testScrapperConfig(time.Second)
	conf.RetryConfig.MaxAttempts = 1
	conf.RetryConfig.Delay = time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	conf.CircuitBreakerConfig.Interval = time.Second
	conf.CircuitBreakerConfig.Timeout = 200 * time.Millisecond
	conf.CircuitBreakerConfig.MaxRequests = 1
	conf.CircuitBreakerConfig.FailureRatio = 0.5

	client := NewScrapperClient(conf)

	errCall1 := callGithub(client, server.URL)
	require.Error(t, errCall1)

	require.Equal(t, gobreaker.StateOpen, client.GithubBreaker.State())

	beforeOpenCalls := calls.Load()

	start := time.Now()
	errCall2 := callGithub(client, server.URL)
	elapsed := time.Since(start)

	require.Error(t, errCall2)
	assert.Contains(t, errCall2.Error(), gobreaker.ErrOpenState.Error())
	assert.Less(t, elapsed, 50*time.Millisecond)
	assert.Equal(t, beforeOpenCalls, calls.Load())
}

func TestClient_DoGithubRequest_CircuitBreakerHalfOpenToClosed(t *testing.T) {
	var shouldFail atomic.Bool
	shouldFail.Store(true)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if shouldFail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(githubRepoResponse))
	}))
	defer server.Close()

	conf := testScrapperConfig(time.Second)
	conf.RetryConfig.MaxAttempts = 1
	conf.RetryConfig.Delay = time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	conf.CircuitBreakerConfig.Interval = time.Second
	conf.CircuitBreakerConfig.Timeout = 100 * time.Millisecond
	conf.CircuitBreakerConfig.MaxRequests = 2
	conf.CircuitBreakerConfig.FailureRatio = 0.5

	client := NewScrapperClient(conf)

	err := callGithub(client, server.URL)
	require.Error(t, err)

	require.Equal(t, gobreaker.StateOpen, client.GithubBreaker.State())

	time.Sleep(150 * time.Millisecond)

	shouldFail.Store(false)

	err = callGithub(client, server.URL)
	require.NoError(t, err)

	err = callGithub(client, server.URL)
	require.NoError(t, err)

	assert.Equal(t, gobreaker.StateClosed, client.GithubBreaker.State())
}

func TestClient_DoGithubRequest_CircuitBreakerHalfOpenToOpen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	conf := testScrapperConfig(time.Second)
	conf.RetryConfig.MaxAttempts = 1
	conf.RetryConfig.Delay = time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	conf.CircuitBreakerConfig.Interval = time.Second
	conf.CircuitBreakerConfig.Timeout = 100 * time.Millisecond
	conf.CircuitBreakerConfig.MaxRequests = 1
	conf.CircuitBreakerConfig.FailureRatio = 0.5

	client := NewScrapperClient(conf)

	err := callGithub(client, server.URL)
	require.Error(t, err)

	require.Equal(t, gobreaker.StateOpen, client.GithubBreaker.State())

	time.Sleep(150 * time.Millisecond)

	err = callGithub(client, server.URL)
	require.Error(t, err)

	assert.Equal(t, gobreaker.StateOpen, client.GithubBreaker.State())
}

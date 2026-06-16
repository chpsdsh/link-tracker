package botclient

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/config"
	config2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
)

func testBotConfig(scrapperURL string) config.BotConfig {
	return config.BotConfig{
		ScrapperServerAddress: scrapperURL,
		HTTPClientConfig: config2.HTTPClientConfig{
			Timeout: time.Second,
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

func TestClient_RegisterChat_Timeout(t *testing.T) {
	serverDelay := 300 * time.Millisecond
	clientTimeout := 50 * time.Millisecond

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(serverDelay)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	conf := testBotConfig(server.URL)
	conf.HTTPClientConfig.Timeout = clientTimeout
	conf.RetryConfig.MaxAttempts = 1

	client := NewBotClient(conf)

	start := time.Now()
	err := client.RegisterChat(123)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, serverDelay)
}

func TestClient_RegisterChat_RetryOn5xx(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := calls.Add(1)

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/tg-chat/123", r.URL.Path)

		switch call {
		case 1, 2:
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	conf := testBotConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 3
	conf.RetryConfig.Delay = 10 * time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	client := NewBotClient(conf)

	err := client.RegisterChat(123)

	require.NoError(t, err)
	assert.Equal(t, int32(3), calls.Load())
}

func TestClient_RegisterChat_NoRetryOn4xx(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	conf := testBotConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 3
	conf.RetryConfig.Delay = 10 * time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	client := NewBotClient(conf)

	err := client.RegisterChat(123)

	require.Error(t, err)
	require.ErrorIs(t, err, handler.ErrIncorrectRequestParameters)
	assert.Equal(t, int32(1), calls.Load())
}

func TestClient_RegisterChat_UsesConstantBackoff(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		call := calls.Add(1)

		if call < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	retryDelay := 80 * time.Millisecond

	conf := testBotConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 3
	conf.RetryConfig.Delay = retryDelay
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	client := NewBotClient(conf)

	start := time.Now()
	err := client.RegisterChat(123)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, int32(3), calls.Load())

	assert.GreaterOrEqual(t, elapsed, 2*retryDelay)
	assert.Less(t, elapsed, 2*retryDelay+200*time.Millisecond)
}

func TestClient_RegisterChat_CircuitBreakerGoesOpen(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	conf := testBotConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 1
	conf.RetryConfig.Delay = time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	conf.CircuitBreakerConfig.Interval = time.Second
	conf.CircuitBreakerConfig.Timeout = 200 * time.Millisecond
	conf.CircuitBreakerConfig.MaxRequests = 1
	conf.CircuitBreakerConfig.FailureRatio = 0.5

	client := NewBotClient(conf)

	err := client.RegisterChat(123)
	require.Error(t, err)

	require.Equal(t, gobreaker.StateOpen, client.Breaker.State())

	beforeOpenCalls := calls.Load()

	start := time.Now()
	err = client.RegisterChat(123)
	elapsed := time.Since(start)

	require.Error(t, err)
	require.ErrorIs(t, err, gobreaker.ErrOpenState)
	assert.Less(t, elapsed, 50*time.Millisecond)
	assert.Equal(t, beforeOpenCalls, calls.Load())
}

func TestClient_RegisterChat_CircuitBreakerHalfOpenToClosed(t *testing.T) {
	var shouldFail atomic.Bool
	shouldFail.Store(true)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if shouldFail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	conf := testBotConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 1
	conf.RetryConfig.Delay = time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	conf.CircuitBreakerConfig.Interval = time.Second
	conf.CircuitBreakerConfig.Timeout = 100 * time.Millisecond
	conf.CircuitBreakerConfig.MaxRequests = 2
	conf.CircuitBreakerConfig.FailureRatio = 0.5

	client := NewBotClient(conf)

	err := client.RegisterChat(123)
	require.Error(t, err)

	require.Equal(t, gobreaker.StateOpen, client.Breaker.State())

	time.Sleep(150 * time.Millisecond)

	shouldFail.Store(false)

	err = client.RegisterChat(123)
	require.NoError(t, err)

	err = client.RegisterChat(123)
	require.NoError(t, err)

	assert.Equal(t, gobreaker.StateClosed, client.Breaker.State())
}

func TestClient_RegisterChat_CircuitBreakerHalfOpenToOpen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	conf := testBotConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 1
	conf.RetryConfig.Delay = time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	conf.CircuitBreakerConfig.Interval = time.Second
	conf.CircuitBreakerConfig.Timeout = 100 * time.Millisecond
	conf.CircuitBreakerConfig.MaxRequests = 1
	conf.CircuitBreakerConfig.FailureRatio = 0.5

	client := NewBotClient(conf)

	err := client.RegisterChat(123)
	require.Error(t, err)

	require.Equal(t, gobreaker.StateOpen, client.Breaker.State())

	time.Sleep(150 * time.Millisecond)

	err = client.RegisterChat(123)
	require.Error(t, err)

	assert.Equal(t, gobreaker.StateOpen, client.Breaker.State())
}

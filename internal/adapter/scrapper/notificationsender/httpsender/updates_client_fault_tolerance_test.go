package httpsender

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"
	config2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

func testUpdatesConfig(botServerAddr string) config2.ScrapperConfig {
	return config2.ScrapperConfig{
		BotServerAddr: botServerAddr,
		HTTPClientConfig: config.HTTPClientConfig{
			Timeout: time.Second,
		},
		RetryConfig: config.RetryConfig{
			MaxAttempts:       3,
			Delay:             50 * time.Millisecond,
			RetryableStatuses: []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout},
		},
		CircuitBreakerConfig: config.CircuitBreakerConfig{
			Interval:     time.Second,
			Timeout:      100 * time.Millisecond,
			MaxRequests:  2,
			FailureRatio: 0.5,
		},
	}
}

func testUpdate() pkg.LinkUpdate {
	return pkg.LinkUpdate{
		ID:          1,
		URL:         "https://github.com/golang/go",
		Description: "new update",
		TgChatIDs:   []int64{123},
	}
}

func TestUpdatesClient_SendLinkUpdate_Timeout(t *testing.T) {
	serverDelay := 300 * time.Millisecond
	clientTimeout := 50 * time.Millisecond

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(serverDelay)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	conf := testUpdatesConfig(server.URL)
	conf.HTTPClientConfig.Timeout = clientTimeout
	conf.RetryConfig.MaxAttempts = 1

	client := NewUpdatesClient(conf)

	start := time.Now()
	err := client.SendLinkUpdate(testUpdate(), "")
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, serverDelay)
}

func TestUpdatesClient_SendLinkUpdate_RetryOn5xx(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := calls.Add(1)

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, botPostEndpoint, r.URL.Path)

		switch call {
		case 1, 2:
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	conf := testUpdatesConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 3
	conf.RetryConfig.Delay = 10 * time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	client := NewUpdatesClient(conf)

	err := client.SendLinkUpdate(testUpdate(), "")

	require.NoError(t, err)
	assert.Equal(t, int32(3), calls.Load())
}

func TestUpdatesClient_SendLinkUpdate_NoRetryOn4xx(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	conf := testUpdatesConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 3
	conf.RetryConfig.Delay = 10 * time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	client := NewUpdatesClient(conf)

	err := client.SendLinkUpdate(testUpdate(), "")

	require.Error(t, err)
	require.ErrorIs(t, err, scrapper.ErrIncorrectRequestParameters)
	assert.Equal(t, int32(1), calls.Load())
}

func TestUpdatesClient_SendLinkUpdate_ConstantBackoff(t *testing.T) {
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

	conf := testUpdatesConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 3
	conf.RetryConfig.Delay = retryDelay
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	client := NewUpdatesClient(conf)

	start := time.Now()
	err := client.SendLinkUpdate(testUpdate(), "")
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, int32(3), calls.Load())

	assert.GreaterOrEqual(t, elapsed, 2*retryDelay)
	assert.Less(t, elapsed, 2*retryDelay+200*time.Millisecond)
}

func TestUpdatesClient_SendLinkUpdate_CircuitBreakerGoesOpen(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	conf := testUpdatesConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 1
	conf.RetryConfig.Delay = time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	conf.CircuitBreakerConfig.Interval = time.Second
	conf.CircuitBreakerConfig.Timeout = 200 * time.Millisecond
	conf.CircuitBreakerConfig.MaxRequests = 1
	conf.CircuitBreakerConfig.FailureRatio = 0.5

	client := NewUpdatesClient(conf)

	err := client.SendLinkUpdate(testUpdate(), "")
	require.Error(t, err)

	err = client.SendLinkUpdate(testUpdate(), "")
	require.Error(t, err)

	require.Equal(t, gobreaker.StateOpen, client.Breaker.State())

	beforeOpenCalls := calls.Load()

	start := time.Now()
	err = client.SendLinkUpdate(testUpdate(), "")
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), gobreaker.ErrOpenState.Error())
	assert.Less(t, elapsed, 50*time.Millisecond)
	assert.Equal(t, beforeOpenCalls, calls.Load())
}

func TestUpdatesClient_SendLinkUpdate_CircuitBreakerHalfOpenToClosed(t *testing.T) {
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

	conf := testUpdatesConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 1
	conf.RetryConfig.Delay = time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	conf.CircuitBreakerConfig.Interval = time.Second
	conf.CircuitBreakerConfig.Timeout = 100 * time.Millisecond
	conf.CircuitBreakerConfig.MaxRequests = 2
	conf.CircuitBreakerConfig.FailureRatio = 0.5

	client := NewUpdatesClient(conf)

	err := client.SendLinkUpdate(testUpdate(), "")
	require.Error(t, err)

	err = client.SendLinkUpdate(testUpdate(), "")
	require.Error(t, err)

	require.Equal(t, gobreaker.StateOpen, client.Breaker.State())

	time.Sleep(150 * time.Millisecond)

	shouldFail.Store(false)

	err = client.SendLinkUpdate(testUpdate(), "")
	require.NoError(t, err)

	err = client.SendLinkUpdate(testUpdate(), "")
	require.NoError(t, err)

	assert.Equal(t, gobreaker.StateClosed, client.Breaker.State())
}

func TestUpdatesClient_SendLinkUpdate_CircuitBreakerHalfOpenToOpen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	conf := testUpdatesConfig(server.URL)
	conf.RetryConfig.MaxAttempts = 1
	conf.RetryConfig.Delay = time.Millisecond
	conf.RetryConfig.RetryableStatuses = []int{http.StatusInternalServerError}

	conf.CircuitBreakerConfig.Interval = time.Second
	conf.CircuitBreakerConfig.Timeout = 100 * time.Millisecond
	conf.CircuitBreakerConfig.MaxRequests = 1
	conf.CircuitBreakerConfig.FailureRatio = 0.5

	client := NewUpdatesClient(conf)

	err := client.SendLinkUpdate(testUpdate(), "")
	require.Error(t, err)

	err = client.SendLinkUpdate(testUpdate(), "")
	require.Error(t, err)

	require.Equal(t, gobreaker.StateOpen, client.Breaker.State())

	time.Sleep(150 * time.Millisecond)

	err = client.SendLinkUpdate(testUpdate(), "")
	require.Error(t, err)

	assert.Equal(t, gobreaker.StateOpen, client.Breaker.State())
}

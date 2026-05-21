package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIPRateLimiter_Returns429WhenLimitExceeded(t *testing.T) {
	calls := 0

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})

	limiter := NewIPRateLimiter(1, 1)
	defer limiter.Close()

	handler := limiter.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/links", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, calls)

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusTooManyRequests, rec.Code)
	require.Equal(t, 1, calls)
}

func TestIPRateLimiter_UsesSeparateLimitersPerIP(t *testing.T) {
	calls := 0

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})

	limiter := NewIPRateLimiter(1, 1)
	defer limiter.Close()

	handler := limiter.Middleware(next)

	req1 := httptest.NewRequest(http.MethodGet, "/links", nil)
	req1.RemoteAddr = "127.0.0.1:12345"

	req2 := httptest.NewRequest(http.MethodGet, "/links", nil)
	req2.RemoteAddr = "127.0.0.2:12345"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req1)
	require.Equal(t, http.StatusOK, rec.Code)

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req2)
	require.Equal(t, http.StatusOK, rec.Code)

	require.Equal(t, 2, calls)
}

func TestIPRateLimiter_AllowsRequestAfterRefill(t *testing.T) {
	calls := 0

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})

	limiter := NewIPRateLimiter(10, 1)
	defer limiter.Close()

	handler := limiter.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/links", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusTooManyRequests, rec.Code)

	time.Sleep(150 * time.Millisecond)

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	require.Equal(t, 2, calls)
}

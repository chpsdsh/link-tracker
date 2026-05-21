package botclient

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/config"
	conf2 "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/config"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

func TestClientRegisterChat(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		expectedErr    error
		expectedMethod string
		expectedPath   string
	}{
		{
			name:           "success",
			status:         http.StatusOK,
			expectedErr:    nil,
			expectedMethod: http.MethodPost,
			expectedPath:   "/tg-chat/123",
		},
		{
			name:           "bad request",
			status:         http.StatusBadRequest,
			expectedErr:    handler.ErrIncorrectRequestParameters,
			expectedMethod: http.MethodPost,
			expectedPath:   "/tg-chat/123",
		},
		{
			name:           "chat already exists",
			status:         http.StatusConflict,
			expectedErr:    handler.ErrChatAlreadyExists,
			expectedMethod: http.MethodPost,
			expectedPath:   "/tg-chat/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.expectedMethod {
					t.Fatalf("expected method %s, got %s", tt.expectedMethod, r.Method)
				}
				if r.URL.Path != tt.expectedPath {
					t.Fatalf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
				}
				w.WriteHeader(tt.status)
			}))
			defer server.Close()
			conf := config.BotConfig{ScrapperServerAddress: server.URL, RetryConfig: conf2.RetryConfig{MaxAttempts: 3,
				Delay: 10 * time.Second, RetryableStatuses: []int{500}},
				CircuitBreakerConfig: conf2.CircuitBreakerConfig{Interval: 10 * time.Second, Timeout: 10 * time.Second, MaxRequests: 10, FailureRatio: 0.5}}
			client := NewBotClient(conf)

			err := client.RegisterChat(123)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestClientUnregisterChat(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		expectedErr    error
		expectedMethod string
		expectedPath   string
	}{
		{
			name:           "success",
			status:         http.StatusOK,
			expectedErr:    nil,
			expectedMethod: http.MethodDelete,
			expectedPath:   "/tg-chat/123",
		},
		{
			name:           "bad request",
			status:         http.StatusBadRequest,
			expectedErr:    handler.ErrIncorrectRequestParameters,
			expectedMethod: http.MethodDelete,
			expectedPath:   "/tg-chat/123",
		},
		{
			name:           "chat not found",
			status:         http.StatusNotFound,
			expectedErr:    handler.ErrChatNotFound,
			expectedMethod: http.MethodDelete,
			expectedPath:   "/tg-chat/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.expectedMethod {
					t.Fatalf("expected method %s, got %s", tt.expectedMethod, r.Method)
				}
				if r.URL.Path != tt.expectedPath {
					t.Fatalf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
				}
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			conf := config.BotConfig{ScrapperServerAddress: server.URL, RetryConfig: conf2.RetryConfig{MaxAttempts: 3,
				Delay: 10 * time.Second, RetryableStatuses: []int{500}},
				CircuitBreakerConfig: conf2.CircuitBreakerConfig{Interval: 10 * time.Second, Timeout: 10 * time.Second, MaxRequests: 10, FailureRatio: 0.5}}
			client := NewBotClient(conf)

			err := client.UnregisterChat(123)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestClientGetLinks(t *testing.T) {
	expectedResp := bot.ListLinksResponse{
		Links: []bot.LinkResponse{
			{URL: "https://github.com/golang/go", Tags: []string{"work"}},
			{URL: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}},
		},
		Size: 2,
	}

	tests := []struct {
		name         string
		status       int
		responseBody any
		expectedErr  error
		expectedSize int32
		expectedPath string
		expectedChat string
	}{
		{
			name:         "success",
			status:       http.StatusOK,
			responseBody: expectedResp,
			expectedErr:  nil,
			expectedSize: 2,
			expectedPath: "/links",
			expectedChat: "123",
		},
		{
			name:         "bad request",
			status:       http.StatusBadRequest,
			expectedErr:  handler.ErrIncorrectRequestParameters,
			expectedPath: "/links",
			expectedChat: "123",
		},
		{
			name:         "chat not found",
			status:       http.StatusNotFound,
			expectedErr:  handler.ErrChatNotFound,
			expectedPath: "/links",
			expectedChat: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != tt.expectedPath {
					t.Fatalf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
				}
				if r.Header.Get(tgHeaderKey) != tt.expectedChat {
					t.Fatalf("expected header %s=%s, got %s", tgHeaderKey, tt.expectedChat, r.Header.Get(tgHeaderKey))
				}

				w.WriteHeader(tt.status)
				if tt.responseBody != nil {
					if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {
						t.Fatal(err)
					}
				}
			}))
			defer server.Close()

			conf := config.BotConfig{ScrapperServerAddress: server.URL, RetryConfig: conf2.RetryConfig{MaxAttempts: 3,
				Delay: 10 * time.Second, RetryableStatuses: []int{500}},
				CircuitBreakerConfig: conf2.CircuitBreakerConfig{Interval: 10 * time.Second, Timeout: 10 * time.Second, MaxRequests: 10, FailureRatio: 0.5}}
			client := NewBotClient(conf)

			resp, err := client.GetLinks(123)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil && resp.Size != tt.expectedSize {
				t.Fatalf("expected size %d, got %d", tt.expectedSize, resp.Size)
			}
		})
	}
}

func TestClientAddLink(t *testing.T) {
	linkReq := pkg.AddLinkRequest{
		Link: "https://github.com/golang/go",
		Tags: []string{"work"},
	}
	expectedResp := bot.LinkResponse{
		URL:  "https://github.com/golang/go",
		Tags: []string{"work"},
	}

	tests := []struct {
		name         string
		status       int
		responseBody any
		expectedErr  error
	}{
		{
			name:         "success",
			status:       http.StatusOK,
			responseBody: expectedResp,
			expectedErr:  nil,
		},
		{
			name:        "bad request",
			status:      http.StatusBadRequest,
			expectedErr: handler.ErrIncorrectRequestParameters,
		},
		{
			name:        "chat not found",
			status:      http.StatusNotFound,
			expectedErr: handler.ErrChatNotFound,
		},
		{
			name:        "link exists",
			status:      http.StatusConflict,
			expectedErr: handler.ErrLinkExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assertAddLinkRequest(t, r, linkReq)

				w.WriteHeader(tt.status)
				writeJSONResponse(t, w, tt.responseBody)
			}))
			defer server.Close()

			conf := config.BotConfig{ScrapperServerAddress: server.URL, RetryConfig: conf2.RetryConfig{MaxAttempts: 3,
				Delay: 10 * time.Second, RetryableStatuses: []int{500}},
				CircuitBreakerConfig: conf2.CircuitBreakerConfig{Interval: 10 * time.Second, Timeout: 10 * time.Second, MaxRequests: 10, FailureRatio: 0.5}}
			client := NewBotClient(conf)

			resp, err := client.AddLink(123, linkReq)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil && resp.URL != expectedResp.URL {
				t.Fatalf("expected url %s, got %s", expectedResp.URL, resp.URL)
			}
		})
	}
}

func assertAddLinkRequest(t *testing.T, r *http.Request, expected pkg.AddLinkRequest) {
	t.Helper()

	if r.Method != http.MethodPost {
		t.Fatalf("expected POST, got %s", r.Method)
	}
	if r.URL.Path != "/links" {
		t.Fatalf("expected path /links, got %s", r.URL.Path)
	}
	if r.Header.Get(tgHeaderKey) != "123" {
		t.Fatalf("expected header %s=123, got %s", tgHeaderKey, r.Header.Get(tgHeaderKey))
	}
	if r.Header.Get(contentTypeKey) != typeApplicationJSON {
		t.Fatalf("expected header %s=%s, got %s", contentTypeKey, typeApplicationJSON, r.Header.Get(contentTypeKey))
	}

	var reqBody pkg.AddLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		t.Fatal(err)
	}

	if reqBody.Link != expected.Link {
		t.Fatalf("expected link %s, got %s", expected.Link, reqBody.Link)
	}
	if len(reqBody.Tags) != len(expected.Tags) {
		t.Fatalf("expected tags %+v, got %+v", expected.Tags, reqBody.Tags)
	}
	for i := range expected.Tags {
		if reqBody.Tags[i] != expected.Tags[i] {
			t.Fatalf("expected tags %+v, got %+v", expected.Tags, reqBody.Tags)
		}
	}
}

func writeJSONResponse(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()

	if body == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatal(err)
	}
}

func TestClientRemoveLink(t *testing.T) {
	removeReq := bot.RemoveLinkRequest{
		Link: "https://github.com/golang/go",
	}
	expectedResp := bot.LinkResponse{
		URL:  "https://github.com/golang/go",
		Tags: []string{"work"},
	}

	tests := []struct {
		name         string
		status       int
		responseBody any
		expectedErr  error
	}{
		{
			name:         "success",
			status:       http.StatusOK,
			responseBody: expectedResp,
			expectedErr:  nil,
		},
		{
			name:        "bad request",
			status:      http.StatusBadRequest,
			expectedErr: handler.ErrIncorrectRequestParameters,
		},
		{
			name:        "link not exists",
			status:      http.StatusNotFound,
			expectedErr: handler.ErrLinkNotExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assertRemoveLinkRequest(t, r, removeReq)

				w.WriteHeader(tt.status)
				writeJSONResponse(t, w, tt.responseBody)
			}))
			defer server.Close()

			conf := config.BotConfig{ScrapperServerAddress: server.URL, RetryConfig: conf2.RetryConfig{MaxAttempts: 3,
				Delay: 10 * time.Second, RetryableStatuses: []int{500}},
				CircuitBreakerConfig: conf2.CircuitBreakerConfig{Interval: 10 * time.Second, Timeout: 10 * time.Second, MaxRequests: 10, FailureRatio: 0.5}}
			client := NewBotClient(conf)

			resp, err := client.RemoveLink(123, removeReq)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil && resp.URL != expectedResp.URL {
				t.Fatalf("expected url %s, got %s", expectedResp.URL, resp.URL)
			}
		})
	}
}

func assertRemoveLinkRequest(t *testing.T, r *http.Request, expected bot.RemoveLinkRequest) {
	t.Helper()

	if r.Method != http.MethodDelete {
		t.Fatalf("expected DELETE, got %s", r.Method)
	}
	if r.URL.Path != "/links" {
		t.Fatalf("expected path /links, got %s", r.URL.Path)
	}
	if r.Header.Get(tgHeaderKey) != "123" {
		t.Fatalf("expected header %s=123, got %s", tgHeaderKey, r.Header.Get(tgHeaderKey))
	}
	if r.Header.Get(contentTypeKey) != typeApplicationJSON {
		t.Fatalf("expected header %s=%s, got %s", contentTypeKey, typeApplicationJSON, r.Header.Get(contentTypeKey))
	}

	var reqBody bot.RemoveLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		t.Fatal(err)
	}

	if reqBody.Link != expected.Link {
		t.Fatalf("expected link %s, got %s", expected.Link, reqBody.Link)
	}
}

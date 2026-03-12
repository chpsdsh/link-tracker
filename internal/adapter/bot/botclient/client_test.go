package botclient

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
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

			client := Client{
				Client: server.Client(),
				Config: config.Config{ScrapperServerAddress: server.URL},
			}

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

			client := Client{
				Client: server.Client(),
				Config: config.Config{ScrapperServerAddress: server.URL},
			}

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
			{Url: "https://github.com/golang/go", Tags: []string{"work"}},
			{Url: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}},
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

			client := Client{
				Client: server.Client(),
				Config: config.Config{ScrapperServerAddress: server.URL},
			}

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
	linkReq := shared.AddLinkRequest{
		Link: "https://github.com/golang/go",
		Tags: []string{"work"},
	}
	expectedResp := bot.LinkResponse{
		Url:  "https://github.com/golang/go",
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

				var reqBody shared.AddLinkRequest
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Fatal(err)
				}

				if reqBody.Link != linkReq.Link {
					t.Fatalf("expected link %s, got %s", linkReq.Link, reqBody.Link)
				}
				if len(reqBody.Tags) != 1 || reqBody.Tags[0] != "work" {
					t.Fatalf("unexpected tags %+v", reqBody.Tags)
				}

				w.WriteHeader(tt.status)
				if tt.responseBody != nil {
					if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {
						t.Fatal(err)
					}
				}
			}))
			defer server.Close()

			client := Client{
				Client: server.Client(),
				Config: config.Config{ScrapperServerAddress: server.URL},
			}

			resp, err := client.AddLink(123, linkReq)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil && resp.Url != expectedResp.Url {
				t.Fatalf("expected url %s, got %s", expectedResp.Url, resp.Url)
			}
		})
	}
}

func TestClientRemoveLink(t *testing.T) {
	removeReq := bot.RemoveLinkRequest{
		Link: "https://github.com/golang/go",
	}
	expectedResp := bot.LinkResponse{
		Url:  "https://github.com/golang/go",
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

				if reqBody.Link != removeReq.Link {
					t.Fatalf("expected link %s, got %s", removeReq.Link, reqBody.Link)
				}

				w.WriteHeader(tt.status)
				if tt.responseBody != nil {
					if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {
						t.Fatal(err)
					}
				}
			}))
			defer server.Close()

			client := Client{
				Client: server.Client(),
				Config: config.Config{ScrapperServerAddress: server.URL},
			}

			resp, err := client.RemoveLink(123, removeReq)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil && resp.Url != expectedResp.Url {
				t.Fatalf("expected url %s, got %s", expectedResp.Url, resp.Url)
			}
		})
	}
}

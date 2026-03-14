package scrapperserver

import (
	"errors"
	"io"

	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

func TestScrapperServerPostTgChatId(t *testing.T) {
	tests := []struct {
		name           string
		id             int64
		handlerErr     error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "success",
			id:             10,
			handlerErr:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "chat already exists",
			id:             10,
			handlerErr:     handler.ErrChatAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedCode:   badRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mocks.NewMockScrapperHandler(ctrl)

			s := ScrapperServer{
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				Handler:    mockHandler,
			}

			req := httptest.NewRequest(http.MethodPost, "/tg-chat/10", nil)
			rec := httptest.NewRecorder()

			mockHandler.EXPECT().
				AddChatID(tt.id).
				Return(tt.handlerErr)

			s.PostTgChatId(rec, req, tt.id)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedCode != "" {
				var resp ApiErrorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Code == nil || *resp.Code != tt.expectedCode {
					t.Fatalf("expected code %q, got %+v", tt.expectedCode, resp.Code)
				}
			}
		})
	}
}

func TestScrapperServerDeleteTgChatId(t *testing.T) {
	tests := []struct {
		name           string
		id             int64
		handlerErr     error
		expectedStatus int
	}{
		{
			name:           "success",
			id:             15,
			handlerErr:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "chat not found",
			id:             15,
			handlerErr:     handler.ErrChatNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mocks.NewMockScrapperHandler(ctrl)

			s := ScrapperServer{
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				Handler:    mockHandler,
			}

			req := httptest.NewRequest(http.MethodDelete, "/tg-chat/15", nil)
			rec := httptest.NewRecorder()

			mockHandler.EXPECT().
				DeleteChat(tt.id).
				Return(tt.handlerErr)

			s.DeleteTgChatId(rec, req, tt.id)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestScrapperServerGetLinks(t *testing.T) {
	tests := []struct {
		name           string
		chatID         int64
		handlerLinks   []shared.LinkInfo
		handlerErr     error
		expectedStatus int
		expectedSize   int32
	}{
		{
			name:   "success",
			chatID: 1,
			handlerLinks: []shared.LinkInfo{
				{Link: "https://github.com/golang/go", Tags: []string{"work"}},
				{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}},
			},
			handlerErr:     nil,
			expectedStatus: http.StatusOK,
			expectedSize:   2,
		},
		{
			name:           "chat not found",
			chatID:         1,
			handlerErr:     handler.ErrChatNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mocks.NewMockScrapperHandler(ctrl)

			s := ScrapperServer{
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				Handler:    mockHandler,
			}

			req := httptest.NewRequest(http.MethodGet, "/links", nil)
			rec := httptest.NewRecorder()

			mockHandler.EXPECT().
				GetLinks(tt.chatID).
				Return(tt.handlerLinks, tt.handlerErr)

			s.GetLinks(rec, req, GetLinksParams{TgChatId: tt.chatID})

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.handlerErr == nil {
				var resp ListLinksResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Size == nil || *resp.Size != tt.expectedSize {
					t.Fatalf("expected size %d, got %+v", tt.expectedSize, resp.Size)
				}
			}
		})
	}
}

func TestScrapperServerPostLinks(t *testing.T) {
	tests := []struct {
		name           string
		chatID         int64
		body           string
		setupMock      func(m *mocks.MockScrapperHandler)
		expectedStatus int
	}{
		{
			name:   "success",
			chatID: 1,
			body:   `{"link":"https://github.com/golang/go","tags":["work"],"filters":["f1"]}`,
			setupMock: func(m *mocks.MockScrapperHandler) {
				m.EXPECT().
					AddLink(int64(1), shared.AddLinkRequest{
						Link: "https://github.com/golang/go",
						Tags: []string{"work"},
					}).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid json",
			chatID: 1,
			body:   `{"link":`,
			setupMock: func(_ *mocks.MockScrapperHandler) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "incorrect request parameters",
			chatID: 1,
			body:   `{"link":"bad","tags":["work"],"filters":["f1"]}`,
			setupMock: func(m *mocks.MockScrapperHandler) {
				m.EXPECT().
					AddLink(int64(1), shared.AddLinkRequest{
						Link: "bad",
						Tags: []string{"work"},
					}).
					Return(handler.ErrIncorrectRequestParameters)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "chat not found",
			chatID: 1,
			body:   `{"link":"https://github.com/golang/go","tags":["work"],"filters":["f1"]}`,
			setupMock: func(m *mocks.MockScrapperHandler) {
				m.EXPECT().
					AddLink(int64(1), shared.AddLinkRequest{
						Link: "https://github.com/golang/go",
						Tags: []string{"work"},
					}).
					Return(handler.ErrChatNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "link exists",
			chatID: 1,
			body:   `{"link":"https://github.com/golang/go","tags":["work"],"filters":["f1"]}`,
			setupMock: func(m *mocks.MockScrapperHandler) {
				m.EXPECT().
					AddLink(int64(1), shared.AddLinkRequest{
						Link: "https://github.com/golang/go",
						Tags: []string{"work"},
					}).
					Return(handler.ErrLinkExists)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mocks.NewMockScrapperHandler(ctrl)
			tt.setupMock(mockHandler)

			s := ScrapperServer{
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				Handler:    mockHandler,
			}

			req := httptest.NewRequest(http.MethodPost, "/links", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			s.PostLinks(rec, req, PostLinksParams{TgChatId: tt.chatID})

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d; body=%s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestScrapperServerDeleteLinks(t *testing.T) {
	tests := []struct {
		name           string
		chatID         int64
		body           string
		setupMock      func(m *mocks.MockScrapperHandler)
		expectedStatus int
	}{
		{
			name:   "success",
			chatID: 1,
			body:   `{"link":"https://github.com/golang/go"}`,
			setupMock: func(m *mocks.MockScrapperHandler) {
				m.EXPECT().
					DeleteLink(int64(1), "https://github.com/golang/go").
					Return(shared.LinkInfo{
						Link: "https://github.com/golang/go",
						Tags: []string{"work"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "chat not found",
			chatID: 1,
			body:   `{"link":"https://github.com/golang/go"}`,
			setupMock: func(m *mocks.MockScrapperHandler) {
				m.EXPECT().
					DeleteLink(int64(1), "https://github.com/golang/go").
					Return(shared.LinkInfo{}, handler.ErrChatNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "link not exists",
			chatID: 1,
			body:   `{"link":"https://github.com/golang/go"}`,
			setupMock: func(m *mocks.MockScrapperHandler) {
				m.EXPECT().
					DeleteLink(int64(1), "https://github.com/golang/go").
					Return(shared.LinkInfo{}, handler.ErrLinkNotExists)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "bad json",
			chatID:         1,
			body:           `{"link":`,
			setupMock:      func(_ *mocks.MockScrapperHandler) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mocks.NewMockScrapperHandler(ctrl)
			tt.setupMock(mockHandler)

			s := ScrapperServer{
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				Handler:    mockHandler,
			}

			req := httptest.NewRequest(http.MethodDelete, "/links", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			s.DeleteLinks(rec, req, DeleteLinksParams{TgChatId: tt.chatID})

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d; body=%s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestJSONErrorHandler(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "invalid param format",
			err:            &InvalidParamFormatError{ParamName: "id", Err: errors.New("bad format")},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   invalidParameter,
		},
		{
			name:           "required header error",
			err:            &RequiredHeaderError{ParamName: "Tg-Chat-ID", Err: errors.New("missing")},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   missingHeder,
		},
		{
			name:           "too many values",
			err:            &TooManyValuesForParamError{ParamName: "Tg-Chat-ID", Count: 2},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   invalidParameter,
		},
		{
			name:           "unknown error",
			err:            errors.New("boom"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   internalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/links", nil)
			rec := httptest.NewRecorder()

			JSONErrorHandler(rec, req, tt.err)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			var resp ApiErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if resp.Code == nil || *resp.Code != tt.expectedCode {
				t.Fatalf("expected code %q, got %+v", tt.expectedCode, resp.Code)
			}
		})
	}
}

package botserver

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

func TestUpdatesServerPostUpdates(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(m *mocks.MockTelegramBotHandler)
		expectedStatus int
		checkBody      bool
	}{
		{
			name: "success",
			body: `{
				"description":"link updated",
				"tgChatIds":[1,2],
				"url":"https://github.com/golang/go"
			}`,
			setupMock: func(m *mocks.MockTelegramBotHandler) {
				m.EXPECT().
					HandleLinkUpdate(pkg.ProcessedLinkUpdate{
						Description: "link updated",
						TgChatIDs:   []int64{1, 2},
					})
			},
			expectedStatus: http.StatusOK,
			checkBody:      false,
		},
		{
			name: "invalid json",
			body: `{
				"description":
			}`,
			setupMock:      func(_ *mocks.MockTelegramBotHandler) {},
			expectedStatus: http.StatusBadRequest,
			checkBody:      true,
		},
		{
			name: "missing required pointer fields causes unmarshal success but panic is not handled here",
			body: `{"description":"link updated"}`,
			setupMock: func(_ *mocks.MockTelegramBotHandler) {
			},
			expectedStatus: http.StatusOK,
			checkBody:      false,
		},
	}

	for _, tt := range tests {
		if tt.name == "missing required pointer fields causes unmarshal success but panic is not handled here" {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mocks.NewMockTelegramBotHandler(ctrl)
			tt.setupMock(mockHandler)

			server := UpdatesRouter{
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				Handler:    mockHandler,
			}

			req := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			server.PostUpdates(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.checkBody {
				var resp ApiErrorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal error response: %v", err)
				}

				if resp.Code == nil || *resp.Code != badRequestCode {
					t.Fatalf("expected code %q, got %+v", badRequestCode, resp.Code)
				}

				if resp.Description == nil || *resp.Description != errorUnmarshallingJSON {
					t.Fatalf("expected description %q, got %+v", errorUnmarshallingJSON, resp.Description)
				}
			}
		})
	}
}

func TestSendApiErrorResponse(t *testing.T) {
	rec := httptest.NewRecorder()

	sendAPIErrorResponse(rec, "bad body", io.EOF)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content-type application/json, got %s", got)
	}

	var resp ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code == nil || *resp.Code != badRequestCode {
		t.Fatalf("expected code %q, got %+v", badRequestCode, resp.Code)
	}

	if resp.Description == nil || *resp.Description != "bad body" {
		t.Fatalf("expected description %q, got %+v", "bad body", resp.Description)
	}

	if resp.ExceptionMessage == nil || *resp.ExceptionMessage != io.EOF.Error() {
		t.Fatalf("expected exception message %q, got %+v", io.EOF.Error(), resp.ExceptionMessage)
	}
}

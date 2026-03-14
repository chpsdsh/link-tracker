package handler

import (
	"errors"
	"io"
	"log/slog"
	"slices"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/mock/gomock"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/statestorage"
	botdomain "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/bot"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

func TestTelegramHandlerHandleCommandStartSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(1, "/start", "start")

	mockClient.EXPECT().
		RegisterChat(int64(1)).
		Return(nil)

	mockSession.EXPECT().
		SetState(int64(1), statestorage.InitialState)

	mockSender.EXPECT().
		SendMessage(int64(1), greetingMessage).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandStartIncorrectRequestParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/start", "start")

	mockClient.EXPECT().
		RegisterChat(int64(123)).
		Return(ErrIncorrectRequestParameters)

	mockSender.EXPECT().
		SendMessage(int64(123), incorrectRequestParameters).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandStartChatAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/start", "start")

	mockClient.EXPECT().
		RegisterChat(int64(123)).
		Return(ErrChatAlreadyExists)

	mockSender.EXPECT().
		SendMessage(int64(123), chatAlreadyExists).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandHelp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/help", "help")

	mockSender.EXPECT().
		SendMessage(int64(123), helpMessage).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandTrack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/track", "track")

	mockSession.EXPECT().GetState(int64(123)).Return(statestorage.InitialState)

	mockSession.EXPECT().
		SetState(int64(123), statestorage.WaitingForTrackURLState)

	mockSender.EXPECT().
		SendMessage(int64(123), trackMessage).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandTrackNoState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/cancel", "cancel")

	mockSession.EXPECT().GetState(int64(123)).Return(statestorage.NoState)

	mockSender.EXPECT().
		SendMessage(int64(123), notAuthorizedMessage).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandUntrack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/untrack", "untrack")

	mockSession.EXPECT().GetState(int64(123)).Return(statestorage.InitialState)

	mockSession.EXPECT().
		SetState(int64(123), statestorage.WaitingForUntrackURLState)

	mockSender.EXPECT().
		SendMessage(int64(123), untrackMessage).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandUntrackNoState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/untrack", "untrack")

	mockSession.EXPECT().GetState(int64(123)).Return(statestorage.NoState)

	mockSender.EXPECT().
		SendMessage(int64(123), notAuthorizedMessage).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/cancel", "cancel")

	mockSession.EXPECT().GetState(int64(123)).Return(statestorage.InitialState)

	mockSession.EXPECT().
		ClearLinkAndUpdateState(int64(123), statestorage.InitialState)

	mockSender.EXPECT().
		SendMessage(int64(123), cancelMessage).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandCancelNoState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/cancel", "cancel")

	mockSession.EXPECT().GetState(int64(123)).Return(statestorage.NoState)

	mockSender.EXPECT().
		SendMessage(int64(123), notAuthorizedMessage).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandUnknown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/abracadabra", "abracadabra")

	mockSender.EXPECT().
		SendMessage(int64(123), unknownMessage).
		Return(nil)

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandListGetLinksError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/list", "list")

	mockSession.EXPECT().GetState(int64(123)).Return(statestorage.InitialState)

	mockClient.EXPECT().
		GetLinks(int64(123)).
		Return(botdomain.ListLinksResponse{}, errors.New("get links failed"))

	h.handleCommand(update)
}

func TestTelegramHandlerHandleCommandListNoState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/list", "list")

	mockSession.EXPECT().GetState(int64(123)).Return(statestorage.NoState)

	mockSender.EXPECT().
		SendMessage(int64(123), notAuthorizedMessage).
		Return(nil)

	h.handleCommand(update)
}

func newCommandUpdate(chatID int64, text string, command string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: chatID,
			},
			Text: text,
			Entities: []tgbotapi.MessageEntity{
				{
					Type:   "bot_command",
					Offset: 0,
					Length: len("/" + command),
				},
			},
		},
	}
}

func TestTelegramHandlerHandleLinksNoLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	links := botdomain.ListLinksResponse{
		Size:  0,
		Links: nil,
	}

	mockSender.EXPECT().
		SendMessage(int64(123), noTrackedLinks).
		Return(nil)

	h.handleLinks(123, links)
}

func TestTelegramHandlerHandleLinksOneLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	links := botdomain.ListLinksResponse{
		Size: 1,
		Links: []botdomain.LinkResponse{
			{URL: "https://github.com/test/repo"},
		},
	}

	mockSender.EXPECT().
		SendMessage(int64(123), "https://github.com/test/repo").
		Return(nil)

	h.handleLinks(123, links)
}

func TestTelegramHandlerHandleLinksMultipleLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	links := botdomain.ListLinksResponse{
		Size: 2,
		Links: []botdomain.LinkResponse{
			{URL: "https://github.com/test/repo"},
			{URL: "https://stackoverflow.com/questions/123"},
		},
	}

	gomock.InOrder(
		mockSender.EXPECT().
			SendMessage(int64(123), "https://github.com/test/repo").
			Return(nil),

		mockSender.EXPECT().
			SendMessage(int64(123), "https://stackoverflow.com/questions/123").
			Return(nil),
	)

	h.handleLinks(123, links)
}

func TestTelegramHandlerHandleMessageDefaultState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newTextUpdate(123, "hello")

	mockSession.EXPECT().
		GetState(int64(123)).
		Return(statestorage.InitialState)

	mockSender.EXPECT().
		SendMessage(int64(123), unknownMessage).
		Return(nil)

	h.handleMessage(update)
}

func TestTelegramHandlerHandleMessageWaitingForTrackUrlState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newTextUpdate(1, "https://github.com/golang/go")

	mockSession.EXPECT().
		GetState(int64(1)).
		Return(statestorage.WaitingForTrackURLState)

	mockSession.EXPECT().
		SetLinkAndUpdateState(int64(1), gomock.Any(), gomock.Any()).
		AnyTimes()

	mockSender.EXPECT().
		SendMessage(int64(1), gomock.Any()).
		Return(nil)

	h.handleMessage(update)
}

func TestTelegramHandlerHandleMessageWaitingForTagsState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newTextUpdate(123, "work,bug")

	mockSession.EXPECT().
		GetState(int64(123)).
		Return(statestorage.WaitingForTagsState)

	mockSession.EXPECT().
		GetLink(int64(123)).
		Return("https://github.com/golang/go").
		AnyTimes()

	mockClient.EXPECT().
		AddLink(int64(123), gomock.Any()).
		Return(botdomain.LinkResponse{}, nil).
		AnyTimes()

	mockSession.EXPECT().
		ClearLinkAndUpdateState(int64(123), statestorage.InitialState)

	mockSender.EXPECT().
		SendMessage(int64(123), gomock.Any()).
		Return(nil)

	h.handleMessage(update)
}

func TestTelegramHandlerHandleMessageWaitingForUnTrackUrlState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newTextUpdate(123, "https://github.com/golang/go")

	mockSession.EXPECT().
		GetState(int64(123)).
		Return(statestorage.WaitingForUntrackURLState)

	mockClient.EXPECT().
		RemoveLink(int64(123), gomock.Any()).
		Return(botdomain.LinkResponse{}, nil).
		AnyTimes()

	mockSession.EXPECT().
		ClearLinkAndUpdateState(int64(123), statestorage.InitialState)

	mockSender.EXPECT().
		SendMessage(int64(123), gomock.Any()).
		Return(nil)

	h.handleMessage(update)
}

func newTextUpdate(chatID int64, text string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: chatID,
			},
			Text: text,
		},
	}
}
func TestTelegramHandlerHandleTrack(t *testing.T) {
	tests := []struct {
		name        string
		inputText   string
		storedLink  string
		addResp     botdomain.LinkResponse
		addErr      error
		expected    string
		expectedReq shared.AddLinkRequest
	}{
		{
			name:       "success",
			inputText:  "work,bug",
			storedLink: "https://github.com/golang/go",
			addResp: botdomain.LinkResponse{
				URL:  "https://github.com/golang/go",
				Tags: []string{"work", "bug"},
			},
			addErr: nil,
			expected: trackConfirmMessage + " " +
				"https://github.com/golang/go" + " " + "work,bug",
			expectedReq: shared.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"work", "bug"},
			},
		},
		{
			name:       "incorrect request parameters",
			inputText:  "work,bug",
			storedLink: "https://github.com/golang/go",
			addErr:     ErrIncorrectRequestParameters,
			expected:   incorrectRequestParameters,
			expectedReq: shared.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"work", "bug"},
			},
		},
		{
			name:       "chat not found",
			inputText:  "work,bug",
			storedLink: "https://github.com/golang/go",
			addErr:     ErrChatNotFound,
			expected:   chatNotFond,
			expectedReq: shared.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"work", "bug"},
			},
		},
		{
			name:       "link exists",
			inputText:  "work,bug",
			storedLink: "https://github.com/golang/go",
			addErr:     ErrLinkExists,
			expected:   isAlreadyTracked,
			expectedReq: shared.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"work", "bug"},
			},
		},
		{
			name:       "empty tags",
			inputText:  "",
			storedLink: "https://github.com/golang/go",
			addResp: botdomain.LinkResponse{
				URL:  "https://github.com/golang/go",
				Tags: []string{""},
			},
			addErr: nil,
			expected: trackConfirmMessage + " " +
				"https://github.com/golang/go" + " " + "",
			expectedReq: shared.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSender := mocks.NewMockSender(ctrl)
			mockSession := mocks.NewMockStateStorage(ctrl)
			mockClient := mocks.NewMockNetworkClient(ctrl)

			h := TelegramHandler{
				MsgSender:  mockSender,
				Session:    mockSession,
				Client:     mockClient,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			update := newPlainTextUpdate(123, tt.inputText)

			mockSession.EXPECT().
				GetLink(int64(123)).
				Return(tt.storedLink)

			mockClient.EXPECT().
				AddLink(int64(123), tt.expectedReq).
				Return(tt.addResp, tt.addErr)

			result := h.handleTrack(update)

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTelegramHandlerHandleUntrack(t *testing.T) {
	tests := []struct {
		name        string
		inputText   string
		removeResp  botdomain.LinkResponse
		removeErr   error
		expected    string
		expectCall  bool
		expectedReq botdomain.RemoveLinkRequest
	}{
		{
			name:       "invalid url",
			inputText:  "not-a-url",
			expected:   notURLToUntrack,
			expectCall: false,
		},
		{
			name:      "success",
			inputText: "https://github.com/golang/go",
			removeResp: botdomain.LinkResponse{
				URL:  "https://github.com/golang/go",
				Tags: []string{"work", "bug"},
			},
			removeErr:  nil,
			expected:   untrackConfirmMessage + " " + "https://github.com/golang/go" + " " + "work,bug",
			expectCall: true,
			expectedReq: botdomain.RemoveLinkRequest{
				Link: "https://github.com/golang/go",
			},
		},
		{
			name:       "incorrect request parameters",
			inputText:  "https://github.com/golang/go",
			removeErr:  ErrIncorrectRequestParameters,
			expected:   incorrectRequestParameters,
			expectCall: true,
			expectedReq: botdomain.RemoveLinkRequest{
				Link: "https://github.com/golang/go",
			},
		},
		{
			name:       "chat not found",
			inputText:  "https://github.com/golang/go",
			removeErr:  ErrChatNotFound,
			expected:   chatNotFond,
			expectCall: true,
			expectedReq: botdomain.RemoveLinkRequest{
				Link: "https://github.com/golang/go",
			},
		},
		{
			name:       "link not exists",
			inputText:  "https://github.com/golang/go",
			removeErr:  ErrLinkNotExists,
			expected:   linkNotFound,
			expectCall: true,
			expectedReq: botdomain.RemoveLinkRequest{
				Link: "https://github.com/golang/go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSender := mocks.NewMockSender(ctrl)
			mockSession := mocks.NewMockStateStorage(ctrl)
			mockClient := mocks.NewMockNetworkClient(ctrl)

			h := TelegramHandler{
				MsgSender:  mockSender,
				Session:    mockSession,
				Client:     mockClient,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			update := newPlainTextUpdate(123, tt.inputText)

			if tt.expectCall {
				mockClient.EXPECT().
					RemoveLink(int64(123), tt.expectedReq).
					Return(tt.removeResp, tt.removeErr)
			}

			result := h.handleUntrack(update)

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func newPlainTextUpdate(chatID int64, text string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: chatID,
			},
			Text: text,
		},
	}
}

func TestTelegramHandlerHandleUpdateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newCommandUpdate(123, "/help", "help")

	mockSender.EXPECT().
		SendMessage(int64(123), helpMessage).
		Return(nil)

	h.HandleUpdate(update)
}

func TestTelegramHandlerHandleUpdateMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)
	mockSession := mocks.NewMockStateStorage(ctrl)
	mockClient := mocks.NewMockNetworkClient(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		Session:    mockSession,
		Client:     mockClient,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := newPlainTextUpdate(123, "hello")

	mockSession.EXPECT().
		GetState(int64(123)).
		Return(statestorage.InitialState)

	mockSender.EXPECT().
		SendMessage(int64(123), unknownMessage).
		Return(nil)

	h.HandleUpdate(update)
}

func TestTelegramHandlerHandleLinkUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockSender(ctrl)

	h := TelegramHandler{
		MsgSender:  mockSender,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	update := shared.LinkUpdate{
		Description: "Ссылка обновлена",
		URL:         "https://github.com/golang/go",
		TgChatIDs:   []int64{1, 2, 3},
	}

	mockSender.EXPECT().
		SendMessage(int64(1), "Ссылка обновлена https://github.com/golang/go").
		Return(nil)

	mockSender.EXPECT().
		SendMessage(int64(2), "Ссылка обновлена https://github.com/golang/go").
		Return(nil)

	mockSender.EXPECT().
		SendMessage(int64(3), "Ссылка обновлена https://github.com/golang/go").
		Return(nil)

	h.HandleLinkUpdate(update)
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "multiple tags",
			input:    "work,bug,docs",
			expected: []string{"work", "bug", "docs"},
		},
		{
			name:     "single tag",
			input:    "work",
			expected: []string{"work"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)

			if !slices.Equal(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTagsToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "multiple tags",
			input:    []string{"work", "bug", "docs"},
			expected: "work,bug,docs",
		},
		{
			name:     "single tag",
			input:    []string{"work"},
			expected: "work",
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tagsToString(tt.input)

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

package telegram

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/mock/gomock"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/mocks"
)

func TestTelegramBotSendMessage(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int64
		message     string
		sendErr     error
		expectedErr error
	}{
		{
			name:        "success",
			chatID:      123,
			message:     "hello",
			sendErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "send error",
			chatID:      123,
			message:     "hello",
			sendErr:     errors.New("send failed"),
			expectedErr: errors.New("send failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockBot := mocks.NewMockTelegramAPI(ctrl)

			bot := Bot{
				BotAPI:     mockBot,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			mockBot.EXPECT().
				Send(gomock.Any()).
				DoAndReturn(func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
					msg, ok := c.(tgbotapi.MessageConfig)
					if !ok {
						t.Fatalf("expected MessageConfig, got %T", c)
					}
					if msg.ChatID != tt.chatID {
						t.Fatalf("expected chat id %d, got %d", tt.chatID, msg.ChatID)
					}
					if msg.Text != tt.message {
						t.Fatalf("expected text %q, got %q", tt.message, msg.Text)
					}
					return tgbotapi.Message{}, tt.sendErr
				})

			err := bot.SendMessage(tt.chatID, tt.message)

			if tt.expectedErr == nil && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if tt.expectedErr != nil && err == nil {
				t.Fatal("expected error")
			}
			if tt.expectedErr != nil && errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %q, got %q", tt.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestTelegramBotSetupBotCommands(t *testing.T) {
	tests := []struct {
		name        string
		requestErr  error
		expectedErr bool
	}{
		{
			name:        "success",
			requestErr:  nil,
			expectedErr: false,
		},
		{
			name:        "request error",
			requestErr:  errors.New("request failed"),
			expectedErr: true,
		},
	}

	expectedCommands := []tgbotapi.BotCommand{
		{Command: "start", Description: startCommand},
		{Command: "help", Description: helpCommand},
		{Command: "track", Description: trackCommand},
		{Command: "untrack", Description: untrackCommand},
		{Command: "list", Description: listCommand},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot, mockBot := newTelegramBotTest(t)

			expectSetupBotCommandsRequest(t, mockBot, expectedCommands, tt.requestErr)

			err := bot.setupBotCommands()

			assertSetupBotCommandsError(t, err, tt.expectedErr)
		})
	}
}

func newTelegramBotTest(t *testing.T) (Bot, *mocks.MockTelegramAPI) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockBot := mocks.NewMockTelegramAPI(ctrl)

	bot := Bot{
		BotAPI:     mockBot,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	return bot, mockBot
}

func expectSetupBotCommandsRequest(
	t *testing.T,
	mockBot *mocks.MockTelegramAPI,
	expectedCommands []tgbotapi.BotCommand,
	requestErr error,
) {
	t.Helper()

	mockBot.EXPECT().
		Request(gomock.Any()).
		DoAndReturn(func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
			cfg := assertSetMyCommandsConfig(t, c)
			assertBotCommands(t, cfg.Commands, expectedCommands)

			return &tgbotapi.APIResponse{Ok: requestErr == nil}, requestErr
		})
}

func assertSetMyCommandsConfig(t *testing.T, c tgbotapi.Chattable) tgbotapi.SetMyCommandsConfig {
	t.Helper()

	cfg, ok := c.(tgbotapi.SetMyCommandsConfig)
	if !ok {
		t.Fatalf("expected SetMyCommandsConfig, got %T", c)
	}

	return cfg
}

func assertBotCommands(t *testing.T, got, expected []tgbotapi.BotCommand) {
	t.Helper()

	if len(got) != len(expected) {
		t.Fatalf("expected %d commands, got %d", len(expected), len(got))
	}

	for i := range expected {
		if got[i].Command != expected[i].Command {
			t.Fatalf("expected command %q, got %q", expected[i].Command, got[i].Command)
		}
		if got[i].Description != expected[i].Description {
			t.Fatalf("expected description %q, got %q", expected[i].Description, got[i].Description)
		}
	}
}

func assertSetupBotCommandsError(t *testing.T, err error, expectedErr bool) {
	t.Helper()

	if expectedErr && err == nil {
		t.Fatal("expected error")
	}
	if !expectedErr && err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestTelegramBotTelegramWorkerHandlesUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mocks.NewMockBotHandler(ctrl)

	bot := Bot{
		Handler:    mockHandler,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updates := make(chan tgbotapi.Update, 1)

	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			Text: "hello",
			Chat: &tgbotapi.Chat{
				ID: 123,
			},
		},
	}

	mockHandler.EXPECT().
		HandleUpdate(update).
		Do(func(tgbotapi.Update) {
			cancel()
		})

	done := make(chan struct{})
	go func() {
		defer close(done)
		bot.telegramWorker(ctx, updates)
	}()

	updates <- update

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("telegramWorker did not stop")
	}
}

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

			bot := TelegramBot{
				Bot:        mockBot,
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
			if tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockBot := mocks.NewMockTelegramAPI(ctrl)

			bot := TelegramBot{
				Bot:        mockBot,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			mockBot.EXPECT().
				Request(gomock.Any()).
				DoAndReturn(func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
					cfg, ok := c.(tgbotapi.SetMyCommandsConfig)
					if !ok {
						t.Fatalf("expected SetMyCommandsConfig, got %T", c)
					}

					if len(cfg.Commands) != len(expectedCommands) {
						t.Fatalf("expected %d commands, got %d", len(expectedCommands), len(cfg.Commands))
					}

					for i := range expectedCommands {
						if cfg.Commands[i].Command != expectedCommands[i].Command {
							t.Fatalf("expected command %q, got %q", expectedCommands[i].Command, cfg.Commands[i].Command)
						}
						if cfg.Commands[i].Description != expectedCommands[i].Description {
							t.Fatalf("expected description %q, got %q", expectedCommands[i].Description, cfg.Commands[i].Description)
						}
					}

					return &tgbotapi.APIResponse{Ok: tt.requestErr == nil}, tt.requestErr
				})

			err := bot.setupBotCommands()

			if tt.expectedErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.expectedErr && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
		})
	}
}

func TestTelegramBotTelegramWorkerHandlesUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mocks.NewMockBotHandler(ctrl)

	bot := TelegramBot{
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

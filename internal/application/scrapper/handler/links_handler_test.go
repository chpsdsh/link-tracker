package handler

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

func TestLinksHandlerAddChatId(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int64
		chatExists  bool
		expectedErr error
		expectAdd   bool
	}{
		{
			name:        "chat already exists",
			chatID:      1,
			chatExists:  true,
			expectedErr: ErrChatAlreadyExists,
			expectAdd:   false,
		},
		{
			name:        "chat added successfully",
			chatID:      2,
			chatExists:  false,
			expectedErr: nil,
			expectAdd:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)

			h := LinksHandler{
				Repo:       mockRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			mockRepo.EXPECT().
				ChatExists(tt.chatID).
				Return(tt.chatExists)

			if tt.expectAdd {
				mockRepo.EXPECT().
					AddChat(tt.chatID)
			}

			err := h.AddChatId(tt.chatID)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestLinksHandlerDeleteChat(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int64
		chatExists  bool
		expectedErr error
		expectDel   bool
	}{
		{
			name:        "chat not found",
			chatID:      1,
			chatExists:  false,
			expectedErr: ErrChatNotFound,
			expectDel:   false,
		},
		{
			name:        "chat deleted successfully",
			chatID:      2,
			chatExists:  true,
			expectedErr: nil,
			expectDel:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)

			h := LinksHandler{
				Repo:       mockRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			mockRepo.EXPECT().
				ChatExists(tt.chatID).
				Return(tt.chatExists)

			if tt.expectDel {
				mockRepo.EXPECT().
					DeleteChat(tt.chatID)
			}

			err := h.DeleteChat(tt.chatID)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestLinksHandlerGetLinks(t *testing.T) {
	expectedLinks := []shared.LinkInfo{
		{Link: "https://github.com/golang/go", Tags: []string{"work"}},
		{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}},
	}

	tests := []struct {
		name         string
		chatID       int64
		chatExists   bool
		repoLinks    []shared.LinkInfo
		expectedErr  error
		expectedSize int
	}{
		{
			name:         "chat not found",
			chatID:       1,
			chatExists:   false,
			repoLinks:    nil,
			expectedErr:  ErrChatNotFound,
			expectedSize: 0,
		},
		{
			name:         "links returned successfully",
			chatID:       2,
			chatExists:   true,
			repoLinks:    expectedLinks,
			expectedErr:  nil,
			expectedSize: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)

			h := LinksHandler{
				Repo:       mockRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			mockRepo.EXPECT().
				ChatExists(tt.chatID).
				Return(tt.chatExists)

			if tt.chatExists {
				mockRepo.EXPECT().
					GetLinks(tt.chatID).
					Return(tt.repoLinks)
			}

			links, err := h.GetLinks(tt.chatID)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if len(links) != tt.expectedSize {
				t.Fatalf("expected %d links, got %d", tt.expectedSize, len(links))
			}
		})
	}
}

func TestLinksHandlerAddLink(t *testing.T) {
	tests := []struct {
		name          string
		chatID        int64
		chatExists    bool
		request       shared.AddLinkRequest
		existingLinks []shared.LinkInfo
		expectedErr   error
		expectAdd     bool
	}{
		{
			name:       "chat not found",
			chatID:     1,
			chatExists: false,
			request: shared.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"work"},
			},
			expectedErr: ErrChatNotFound,
			expectAdd:   false,
		},
		{
			name:       "incorrect request parameters",
			chatID:     2,
			chatExists: true,
			request: shared.AddLinkRequest{
				Link: "://bad-url",
				Tags: []string{"work"},
			},
			expectedErr: ErrIncorrectRequestParameters,
			expectAdd:   false,
		},
		{
			name:       "link already exists",
			chatID:     3,
			chatExists: true,
			request: shared.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"work"},
			},
			existingLinks: []shared.LinkInfo{
				{Link: "https://github.com/golang/go", Tags: []string{"old"}},
			},
			expectedErr: ErrLinkExists,
			expectAdd:   false,
		},
		{
			name:       "link added successfully",
			chatID:     4,
			chatExists: true,
			request: shared.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"work", "repo"},
			},
			existingLinks: []shared.LinkInfo{
				{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}},
			},
			expectedErr: nil,
			expectAdd:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)

			h := LinksHandler{
				Repo:       mockRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			mockRepo.EXPECT().
				ChatExists(tt.chatID).
				Return(tt.chatExists)

			if tt.chatExists && tt.expectedErr != ErrIncorrectRequestParameters {
				mockRepo.EXPECT().
					GetLinks(tt.chatID).
					Return(tt.existingLinks)
			}

			if tt.expectAdd {
				mockRepo.EXPECT().
					AddLink(tt.chatID, gomock.Any()).
					Do(func(chatID int64, link shared.LinkInfo) {
						if link.Link != tt.request.Link {
							t.Fatalf("expected link %s, got %s", tt.request.Link, link.Link)
						}
						if len(link.Tags) != len(tt.request.Tags) {
							t.Fatalf("expected tags %v, got %v", tt.request.Tags, link.Tags)
						}
						for i := range link.Tags {
							if link.Tags[i] != tt.request.Tags[i] {
								t.Fatalf("expected tags %v, got %v", tt.request.Tags, link.Tags)
							}
						}
						if time.Since(link.LastUpdateTime) > time.Second {
							t.Fatalf("expected recent LastUpdateTime, got %v", link.LastUpdateTime)
						}
					})
			}

			err := h.AddLink(tt.chatID, tt.request)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestLinksHandlerDeleteLink(t *testing.T) {
	deletedLink := shared.LinkInfo{
		Link: "https://github.com/golang/go",
		Tags: []string{"work"},
	}

	tests := []struct {
		name         string
		chatID       int64
		link         string
		chatExists   bool
		deleteResult shared.LinkInfo
		deleteOK     bool
		expectedErr  error
		expectedLink shared.LinkInfo
		expectDelete bool
	}{
		{
			name:         "chat not found",
			chatID:       1,
			link:         "https://github.com/golang/go",
			chatExists:   false,
			expectedErr:  ErrChatNotFound,
			expectedLink: shared.LinkInfo{},
			expectDelete: false,
		},
		{
			name:         "link not exists",
			chatID:       2,
			link:         "https://github.com/golang/go",
			chatExists:   true,
			deleteResult: shared.LinkInfo{},
			deleteOK:     false,
			expectedErr:  ErrLinkNotExists,
			expectedLink: shared.LinkInfo{},
			expectDelete: true,
		},
		{
			name:         "link deleted successfully",
			chatID:       3,
			link:         "https://github.com/golang/go",
			chatExists:   true,
			deleteResult: deletedLink,
			deleteOK:     true,
			expectedErr:  nil,
			expectedLink: deletedLink,
			expectDelete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)

			h := LinksHandler{
				Repo:       mockRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			mockRepo.EXPECT().
				ChatExists(tt.chatID).
				Return(tt.chatExists)

			if tt.expectDelete {
				mockRepo.EXPECT().
					DeleteLink(tt.chatID, tt.link).
					Return(tt.deleteResult, tt.deleteOK)
			}

			linkInfo, err := h.DeleteLink(tt.chatID, tt.link)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if linkInfo.Link != tt.expectedLink.Link {
				t.Fatalf("expected link %s, got %s", tt.expectedLink.Link, linkInfo.Link)
			}
		})
	}
}

package handler

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
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

			err := h.AddChatID(tt.chatID)

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
	expectedLinks := []pkg.LinkInfo{
		{Link: "https://github.com/golang/go", Tags: []string{"work"}},
		{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}},
	}

	tests := []struct {
		name         string
		chatID       int64
		chatExists   bool
		repoLinks    []pkg.LinkInfo
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
		request       pkg.AddLinkRequest
		existingLinks []pkg.LinkInfo
		expectedErr   error
		expectAdd     bool
	}{
		{
			name:       "chat not found",
			chatID:     1,
			chatExists: false,
			request: pkg.AddLinkRequest{
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
			request: pkg.AddLinkRequest{
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
			request: pkg.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"work"},
			},
			existingLinks: []pkg.LinkInfo{
				{Link: "https://github.com/golang/go", Tags: []string{"old"}},
			},
			expectedErr: ErrLinkExists,
			expectAdd:   false,
		},
		{
			name:       "link added successfully",
			chatID:     4,
			chatExists: true,
			request: pkg.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"work", "repo"},
			},
			existingLinks: []pkg.LinkInfo{
				{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}},
			},
			expectedErr: nil,
			expectAdd:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo := newLinksHandlerTest(t)

			expectChatExists(mockRepo, tt.chatID, tt.chatExists)
			expectGetLinksIfNeeded(mockRepo, tt.chatID, tt.chatExists, tt.expectedErr, tt.existingLinks)
			expectAddLinkIfNeeded(t, mockRepo, tt.chatID, tt.request, tt.expectAdd)

			err := h.AddLink(tt.chatID, tt.request)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func newLinksHandlerTest(t *testing.T) (LinksHandler, *mocks.MockRepository) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := mocks.NewMockRepository(ctrl)

	h := LinksHandler{
		Repo:       mockRepo,
		BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	return h, mockRepo
}

func expectChatExists(mockRepo *mocks.MockRepository, chatID int64, chatExists bool) {
	mockRepo.EXPECT().
		ChatExists(chatID).
		Return(chatExists)
}

func expectGetLinksIfNeeded(
	mockRepo *mocks.MockRepository,
	chatID int64,
	chatExists bool,
	expectedErr error,
	existingLinks []pkg.LinkInfo,
) {
	if !chatExists || errors.Is(expectedErr, ErrIncorrectRequestParameters) {
		return
	}

	mockRepo.EXPECT().
		GetLinks(chatID).
		Return(existingLinks)
}

func expectAddLinkIfNeeded(
	t *testing.T,
	mockRepo *mocks.MockRepository,
	chatID int64,
	request pkg.AddLinkRequest,
	expectAdd bool,
) {
	t.Helper()

	if !expectAdd {
		return
	}

	mockRepo.EXPECT().
		AddLink(chatID, gomock.Any()).
		DoAndReturn(func(_ int64, link pkg.LinkInfo) {
			assertAddedLink(t, link, request)
		})
}

func assertAddedLink(t *testing.T, got pkg.LinkInfo, expected pkg.AddLinkRequest) {
	t.Helper()

	if got.Link != expected.Link {
		t.Fatalf("expected link %s, got %s", expected.Link, got.Link)
	}

	if len(got.Tags) != len(expected.Tags) {
		t.Fatalf("expected tags %v, got %v", expected.Tags, got.Tags)
	}

	for i := range got.Tags {
		if got.Tags[i] != expected.Tags[i] {
			t.Fatalf("expected tags %v, got %v", expected.Tags, got.Tags)
		}
	}

	if time.Since(got.LastUpdateTime) > time.Second {
		t.Fatalf("expected recent LastUpdateTime, got %v", got.LastUpdateTime)
	}
}

func TestLinksHandlerDeleteLink(t *testing.T) {
	deletedLink := pkg.LinkInfo{
		Link: "https://github.com/golang/go",
		Tags: []string{"work"},
	}

	tests := []struct {
		name         string
		chatID       int64
		link         string
		chatExists   bool
		deleteResult pkg.LinkInfo
		deleteOK     bool
		expectedErr  error
		expectedLink pkg.LinkInfo
		expectDelete bool
	}{
		{
			name:         "chat not found",
			chatID:       1,
			link:         "https://github.com/golang/go",
			chatExists:   false,
			expectedErr:  ErrChatNotFound,
			expectedLink: pkg.LinkInfo{},
			expectDelete: false,
		},
		{
			name:         "link not exists",
			chatID:       2,
			link:         "https://github.com/golang/go",
			chatExists:   true,
			deleteResult: pkg.LinkInfo{},
			deleteOK:     false,
			expectedErr:  ErrLinkNotExists,
			expectedLink: pkg.LinkInfo{},
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

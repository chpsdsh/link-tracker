package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

func TestLinksService_AddChatID(t *testing.T) {
	tests := []struct {
		name          string
		chatID        int64
		chatExists    bool
		chatExistsErr error
		addChatErr    error
		expectedErr   error
		expectAddCall bool
	}{
		{
			name:        "chat already exists",
			chatID:      1,
			chatExists:  true,
			expectedErr: scrapper.ErrChatAlreadyExists,
		},
		{
			name:          "chat exists check error",
			chatID:        2,
			chatExistsErr: errors.New("db error"),
			expectedErr:   scrapper.ErrInternalError,
		},
		{
			name:          "add chat error",
			chatID:        3,
			chatExists:    false,
			addChatErr:    errors.New("insert error"),
			expectedErr:   scrapper.ErrInternalError,
			expectAddCall: true,
		},
		{
			name:          "success",
			chatID:        4,
			chatExists:    false,
			expectedErr:   nil,
			expectAddCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			chatRepo := mocks.NewMockChatRepository(ctrl)
			tx := mocks.NewMockTransactor(ctrl)

			svc := LinksService{
				ChatsRepo:  chatRepo,
				Transactor: tx,
			}

			tx.EXPECT().
				Transaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

			chatRepo.EXPECT().
				ChatExists(gomock.Any(), tt.chatID).
				Return(tt.chatExists, tt.chatExistsErr)

			if tt.expectAddCall {
				chatRepo.EXPECT().
					AddChat(gomock.Any(), tt.chatID).
					Return(tt.addChatErr)
			}

			err := svc.AddChatID(context.Background(), tt.chatID)

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestLinksService_DeleteChat(t *testing.T) {
	tests := []struct {
		name        string
		chatID      int64
		deleteErr   error
		expectedErr error
	}{
		{
			name:        "success",
			chatID:      1,
			deleteErr:   nil,
			expectedErr: nil,
		},
		{
			name:        "repo error",
			chatID:      2,
			deleteErr:   errors.New("db error"),
			expectedErr: scrapper.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			chatRepo := mocks.NewMockChatRepository(ctrl)

			svc := LinksService{
				ChatsRepo: chatRepo,
			}

			chatRepo.EXPECT().
				DeleteChat(gomock.Any(), tt.chatID).
				Return(tt.deleteErr)

			err := svc.DeleteChat(context.Background(), tt.chatID)

			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.expectedErr)
			}
		})
	}
}

func TestGetLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name          string
		chatID        int64
		linkRepo      func() LinkRepository
		chatRepo      func() ChatRepository
		expectedLinks []pkg.LinkInfo

		expectedErr error
	}{
		{
			name:   "success",
			chatID: 1,
			chatRepo: func() ChatRepository {
				chatRepo := mocks.NewMockChatRepository(ctrl)
				chatRepo.EXPECT().ChatExists(gomock.Any(), gomock.Any()).Return(true, nil)
				return chatRepo
			},
			linkRepo: func() LinkRepository {
				links := []pkg.LinkInfo{{Link: "https://github.com"}, {Link: "https://stackoverflow.com"}}
				linkRepo := mocks.NewMockLinkRepository(ctrl)
				linkRepo.EXPECT().GetUserLinks(gomock.Any(), gomock.Any()).Return(links, nil)
				return linkRepo
			},
			expectedLinks: []pkg.LinkInfo{{Link: "https://github.com"}, {Link: "https://stackoverflow.com"}},
			expectedErr:   nil,
		},
		{
			name:   "chat repo error",
			chatID: 2,
			chatRepo: func() ChatRepository {
				chatRepo := mocks.NewMockChatRepository(ctrl)
				chatRepo.EXPECT().ChatExists(gomock.Any(), gomock.Any()).Return(false, errors.New("db error"))
				return chatRepo
			},
			linkRepo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedErr: scrapper.ErrInternalError,
		},
		{
			name:   "chat not exists",
			chatID: 3,
			chatRepo: func() ChatRepository {
				chatRepo := mocks.NewMockChatRepository(ctrl)
				chatRepo.EXPECT().ChatExists(gomock.Any(), gomock.Any()).Return(false, nil)
				return chatRepo
			},
			linkRepo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedErr: scrapper.ErrChatNotFound,
		},
		{
			name:   "link repository error",
			chatID: 4,
			chatRepo: func() ChatRepository {
				chatRepo := mocks.NewMockChatRepository(ctrl)
				chatRepo.EXPECT().ChatExists(gomock.Any(), gomock.Any()).Return(true, nil)
				return chatRepo
			},
			linkRepo: func() LinkRepository {
				linkRepo := mocks.NewMockLinkRepository(ctrl)
				linkRepo.EXPECT().GetUserLinks(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
				return linkRepo
			},
			expectedErr: scrapper.ErrInternalError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := mocks.NewMockTransactor(ctrl)
			tx.EXPECT().
				TransactionWithReturn(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
					return fn(ctx)
				})

			srv := LinksService{LinkRepo: tt.linkRepo(), ChatsRepo: tt.chatRepo(), Transactor: tx}
			links, err := srv.GetLinks(context.Background(), tt.chatID)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Equal(t, tt.expectedLinks, links)
		})
	}
}

func TestAddLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		chatID      int64
		request     pkg.AddLinkRequest
		linkRepo    func() LinkRepository
		tx          func() Transactor
		expectedErr error
	}{
		{
			name:   "success",
			chatID: 1,
			request: pkg.AddLinkRequest{
				Link: "https://github.com",
				Tags: []string{"tag"},
			},
			linkRepo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)
				repo.EXPECT().
					AddLink(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				return repo
			},
			tx: func() Transactor {
				tx := mocks.NewMockTransactor(ctrl)
				tx.EXPECT().
					Transaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				return tx
			},
			expectedErr: nil,
		},
		{
			name:   "repo error",
			chatID: 2,
			request: pkg.AddLinkRequest{
				Link: "https://github.com",
			},
			linkRepo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)
				repo.EXPECT().
					AddLink(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
				return repo
			},
			tx: func() Transactor {
				tx := mocks.NewMockTransactor(ctrl)
				tx.EXPECT().
					Transaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				return tx
			},
			expectedErr: scrapper.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := LinksService{
				LinkRepo:   tt.linkRepo(),
				Transactor: tt.tx(),
			}

			err := srv.AddLink(context.Background(), tt.chatID, tt.request)

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestDeleteLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name         string
		chatID       int64
		link         string
		chatRepo     func() ChatRepository
		linkRepo     func() LinkRepository
		expectedLink pkg.LinkInfo
		expectedErr  error
	}{
		{
			name:   "success",
			chatID: 1,
			link:   "https://github.com",
			chatRepo: func() ChatRepository {
				repo := mocks.NewMockChatRepository(ctrl)
				repo.EXPECT().
					ChatExists(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return repo
			},
			linkRepo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					LinkExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				repo.EXPECT().
					DeleteLink(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(pkg.LinkInfo{Link: "https://github.com"}, nil)

				return repo
			},
			expectedLink: pkg.LinkInfo{Link: "https://github.com"},
			expectedErr:  nil,
		},
		{
			name:   "chat repo error",
			chatID: 2,
			link:   "url",
			chatRepo: func() ChatRepository {
				repo := mocks.NewMockChatRepository(ctrl)
				repo.EXPECT().
					ChatExists(gomock.Any(), gomock.Any()).
					Return(false, errors.New("db error"))
				return repo
			},
			linkRepo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedErr: scrapper.ErrInternalError,
		},
		{
			name:   "chat not exists",
			chatID: 3,
			link:   "url",
			chatRepo: func() ChatRepository {
				repo := mocks.NewMockChatRepository(ctrl)
				repo.EXPECT().
					ChatExists(gomock.Any(), gomock.Any()).
					Return(false, nil)
				return repo
			},
			linkRepo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedErr: scrapper.ErrChatNotFound,
		},
		{
			name:   "link exists error",
			chatID: 4,
			link:   "url",
			chatRepo: func() ChatRepository {
				repo := mocks.NewMockChatRepository(ctrl)
				repo.EXPECT().
					ChatExists(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return repo
			},
			linkRepo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)
				repo.EXPECT().
					LinkExists(gomock.Any(), gomock.Any()).
					Return(false, errors.New("db error"))
				return repo
			},
			expectedErr: scrapper.ErrInternalError,
		},
		{
			name:   "link not exists",
			chatID: 5,
			link:   "url",
			chatRepo: func() ChatRepository {
				repo := mocks.NewMockChatRepository(ctrl)
				repo.EXPECT().
					ChatExists(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return repo
			},
			linkRepo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)
				repo.EXPECT().
					LinkExists(gomock.Any(), gomock.Any()).
					Return(false, nil)
				return repo
			},
			expectedErr: scrapper.ErrLinkNotExists,
		},
		{
			name:   "delete error",
			chatID: 6,
			link:   "url",
			chatRepo: func() ChatRepository {
				repo := mocks.NewMockChatRepository(ctrl)
				repo.EXPECT().
					ChatExists(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return repo
			},
			linkRepo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					LinkExists(gomock.Any(), gomock.Any()).
					Return(true, nil)

				repo.EXPECT().
					DeleteLink(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(pkg.LinkInfo{}, errors.New("db error"))

				return repo
			},
			expectedErr: scrapper.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := mocks.NewMockTransactor(ctrl)
			tx.EXPECT().
				TransactionWithReturn(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
					return fn(ctx)
				})

			srv := LinksService{
				LinkRepo:   tt.linkRepo(),
				ChatsRepo:  tt.chatRepo(),
				Transactor: tx,
			}

			link, err := srv.DeleteLink(context.Background(), tt.chatID, tt.link)

			require.ErrorIs(t, err, tt.expectedErr)

			if tt.expectedErr == nil {
				assert.Equal(t, tt.expectedLink, link)
			} else {
				assert.Equal(t, pkg.LinkInfo{}, link)
			}
		})
	}
}

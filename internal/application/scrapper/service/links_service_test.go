package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
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

func TestLinksService_GetLinks(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		setupMocks func(
			cacheRepo *mocks.MockCacheRepository,
			transactor *mocks.MockTransactor,
			chatsRepo *mocks.MockChatRepository,
			linkRepo *mocks.MockLinkRepository,
		)
		expectedLinks []pkg.LinkInfo
		expectedErr   error
	}{
		{
			name:   "cache hit",
			chatID: 1,
			setupMocks: func(
				cacheRepo *mocks.MockCacheRepository,
				_ *mocks.MockTransactor,
				_ *mocks.MockChatRepository,
				_ *mocks.MockLinkRepository,
			) {
				expectedLinks := []pkg.LinkInfo{
					{Link: "https://github.com/golang/go"},
				}

				cacheRepo.EXPECT().
					GetUserLinks(gomock.Any(), int64(1), gomock.Any()).
					Return(expectedLinks, nil)
			},
			expectedLinks: []pkg.LinkInfo{
				{Link: "https://github.com/golang/go"},
			},
			expectedErr: nil,
		},
		{
			name:   "cache miss loader success",
			chatID: 1,
			setupMocks: func(
				cacheRepo *mocks.MockCacheRepository,
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				linkRepo *mocks.MockLinkRepository,
			) {
				expectedLinks := []pkg.LinkInfo{
					{Link: "https://github.com/golang/go"},
					{Link: "https://stackoverflow.com/questions/123/test"},
				}

				cacheRepo.EXPECT().
					GetUserLinks(gomock.Any(), int64(1), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						_ int64,
						loader func(context.Context, string) (*[]pkg.LinkInfo, error),
					) ([]pkg.LinkInfo, error) {
						links, err := loader(ctx, "links:list:1")
						if err != nil {
							return nil, err
						}

						return *links, nil
					})

				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						txFunc func(context.Context) (any, error),
					) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(true, nil)

				linkRepo.EXPECT().
					GetUserLinks(gomock.Any(), int64(1)).
					Return(expectedLinks, nil)
			},
			expectedLinks: []pkg.LinkInfo{
				{Link: "https://github.com/golang/go"},
				{Link: "https://stackoverflow.com/questions/123/test"},
			},
			expectedErr: nil,
		},
		{
			name:   "chat not found",
			chatID: 1,
			setupMocks: func(
				cacheRepo *mocks.MockCacheRepository,
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				_ *mocks.MockLinkRepository,
			) {
				cacheRepo.EXPECT().
					GetUserLinks(gomock.Any(), int64(1), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						_ int64,
						loader func(context.Context, string) (*[]pkg.LinkInfo, error),
					) ([]pkg.LinkInfo, error) {
						links, err := loader(ctx, "links:list:1")
						if err != nil {
							return nil, err
						}

						return *links, nil
					})

				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						txFunc func(context.Context) (any, error),
					) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(false, nil)
			},
			expectedLinks: nil,
			expectedErr:   scrapper.ErrInternalError,
		},
		{
			name:   "chat exists returns error",
			chatID: 1,
			setupMocks: func(
				cacheRepo *mocks.MockCacheRepository,
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				_ *mocks.MockLinkRepository,
			) {
				cacheRepo.EXPECT().
					GetUserLinks(gomock.Any(), int64(1), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						_ int64,
						loader func(context.Context, string) (*[]pkg.LinkInfo, error),
					) ([]pkg.LinkInfo, error) {
						links, err := loader(ctx, "links:list:1")
						if err != nil {
							return nil, err
						}

						return *links, nil
					})

				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						txFunc func(context.Context) (any, error),
					) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(false, errors.New("db error"))
			},
			expectedLinks: nil,
			expectedErr:   scrapper.ErrInternalError,
		},
		{
			name:   "link repo returns error",
			chatID: 1,
			setupMocks: func(
				cacheRepo *mocks.MockCacheRepository,
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				linkRepo *mocks.MockLinkRepository,
			) {
				cacheRepo.EXPECT().
					GetUserLinks(gomock.Any(), int64(1), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						_ int64,
						loader func(context.Context, string) (*[]pkg.LinkInfo, error),
					) ([]pkg.LinkInfo, error) {
						links, err := loader(ctx, "links:list:1")
						if err != nil {
							return nil, err
						}

						return *links, nil
					})

				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						txFunc func(context.Context) (any, error),
					) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(true, nil)

				linkRepo.EXPECT().
					GetUserLinks(gomock.Any(), int64(1)).
					Return(nil, errors.New("db error"))
			},
			expectedLinks: nil,
			expectedErr:   scrapper.ErrInternalError,
		},
		{
			name:   "transaction returns wrong type",
			chatID: 1,
			setupMocks: func(
				cacheRepo *mocks.MockCacheRepository,
				transactor *mocks.MockTransactor,
				_ *mocks.MockChatRepository,
				_ *mocks.MockLinkRepository,
			) {
				cacheRepo.EXPECT().
					GetUserLinks(gomock.Any(), int64(1), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						_ int64,
						loader func(context.Context, string) (*[]pkg.LinkInfo, error),
					) ([]pkg.LinkInfo, error) {
						links, err := loader(ctx, "links:list:1")
						if err != nil {
							return nil, err
						}

						return *links, nil
					})

				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					Return("wrong type", nil)
			},
			expectedLinks: nil,
			expectedErr:   scrapper.ErrInternalError,
		},
		{
			name:   "cache returns error",
			chatID: 1,
			setupMocks: func(
				cacheRepo *mocks.MockCacheRepository,
				_ *mocks.MockTransactor,
				_ *mocks.MockChatRepository,
				_ *mocks.MockLinkRepository,
			) {
				cacheRepo.EXPECT().
					GetUserLinks(gomock.Any(), int64(1), gomock.Any()).
					Return(nil, errors.New("cache error"))
			},
			expectedLinks: nil,
			expectedErr:   scrapper.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cacheRepo := mocks.NewMockCacheRepository(ctrl)
			transactor := mocks.NewMockTransactor(ctrl)
			chatsRepo := mocks.NewMockChatRepository(ctrl)
			linkRepo := mocks.NewMockLinkRepository(ctrl)

			tt.setupMocks(cacheRepo, transactor, chatsRepo, linkRepo)

			service := LinksService{
				CacheRepo:  cacheRepo,
				Transactor: transactor,
				ChatsRepo:  chatsRepo,
				LinkRepo:   linkRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			links, err := service.GetLinks(context.Background(), tt.chatID)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, links)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedLinks, links)
		})
	}
}

func TestLinksService_AddLink(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		request    pkg.AddLinkRequest
		setupMocks func(
			transactor *mocks.MockTransactor,
			linkRepo *mocks.MockLinkRepository,
			cacheRepo *mocks.MockCacheRepository,
		)
		expectedErr error
	}{
		{
			name:   "success",
			chatID: 1,
			request: pkg.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"go", "github"},
			},
			setupMocks: func(
				transactor *mocks.MockTransactor,
				linkRepo *mocks.MockLinkRepository,
				cacheRepo *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					Transaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
						return txFunc(ctx)
					})

				linkRepo.EXPECT().
					AddLink(gomock.Any(), int64(1), pkg.LinkInfo{
						Link: "https://github.com/golang/go",
						Tags: []string{"go", "github"},
					}).
					Return(nil)

				cacheRepo.EXPECT().
					DeleteUserLinks(gomock.Any(), int64(1)).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "add link error",
			chatID: 1,
			request: pkg.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"go"},
			},
			setupMocks: func(
				transactor *mocks.MockTransactor,
				linkRepo *mocks.MockLinkRepository,
				_ *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					Transaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
						return txFunc(ctx)
					})

				linkRepo.EXPECT().
					AddLink(gomock.Any(), int64(1), pkg.LinkInfo{
						Link: "https://github.com/golang/go",
						Tags: []string{"go"},
					}).
					Return(errors.New("db error"))
			},
			expectedErr: scrapper.ErrInternalError,
		},
		{
			name:   "cache delete error does not fail request",
			chatID: 1,
			request: pkg.AddLinkRequest{
				Link: "https://github.com/golang/go",
				Tags: []string{"go"},
			},
			setupMocks: func(
				transactor *mocks.MockTransactor,
				linkRepo *mocks.MockLinkRepository,
				cacheRepo *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					Transaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
						return txFunc(ctx)
					})

				linkRepo.EXPECT().
					AddLink(gomock.Any(), int64(1), pkg.LinkInfo{
						Link: "https://github.com/golang/go",
						Tags: []string{"go"},
					}).
					Return(nil)

				cacheRepo.EXPECT().
					DeleteUserLinks(gomock.Any(), int64(1)).
					Return(errors.New("cache error"))
			},
			expectedErr: nil,
		},
		{
			name:   "transaction error",
			chatID: 1,
			request: pkg.AddLinkRequest{
				Link: "https://github.com/golang/go",
			},
			setupMocks: func(
				transactor *mocks.MockTransactor,
				_ *mocks.MockLinkRepository,
				_ *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					Transaction(gomock.Any(), gomock.Any()).
					Return(errors.New("transaction error"))
			},
			expectedErr: errors.New("transaction error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			transactor := mocks.NewMockTransactor(ctrl)
			linkRepo := mocks.NewMockLinkRepository(ctrl)
			cacheRepo := mocks.NewMockCacheRepository(ctrl)

			tt.setupMocks(transactor, linkRepo, cacheRepo)

			service := LinksService{
				Transactor: transactor,
				LinkRepo:   linkRepo,
				CacheRepo:  cacheRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			err := service.AddLink(context.Background(), tt.chatID, tt.request)

			if tt.expectedErr != nil {
				require.Error(t, err)

				if errors.Is(tt.expectedErr, scrapper.ErrInternalError) {
					require.ErrorIs(t, err, scrapper.ErrInternalError)
				}

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLinksService_DeleteLink(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		link       string
		setupMocks func(
			transactor *mocks.MockTransactor,
			chatsRepo *mocks.MockChatRepository,
			linkRepo *mocks.MockLinkRepository,
			cacheRepo *mocks.MockCacheRepository,
		)
		expectedLink pkg.LinkInfo
		expectedErr  error
	}{
		{
			name:   "success",
			chatID: 1,
			link:   "https://github.com/golang/go",
			setupMocks: func(
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				linkRepo *mocks.MockLinkRepository,
				cacheRepo *mocks.MockCacheRepository,
			) {
				expectedLink := pkg.LinkInfo{
					Link: "https://github.com/golang/go",
					Tags: []string{"go"},
				}

				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, txFunc func(context.Context) (any, error)) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(true, nil)

				linkRepo.EXPECT().
					LinkExists(gomock.Any(), "https://github.com/golang/go").
					Return(true, nil)

				linkRepo.EXPECT().
					DeleteLink(gomock.Any(), int64(1), "https://github.com/golang/go").
					Return(expectedLink, nil)

				cacheRepo.EXPECT().
					DeleteUserLinks(gomock.Any(), int64(1)).
					Return(nil)
			},
			expectedLink: pkg.LinkInfo{
				Link: "https://github.com/golang/go",
				Tags: []string{"go"},
			},
			expectedErr: nil,
		},
		{
			name:   "chat exists error",
			chatID: 1,
			link:   "https://github.com/golang/go",
			setupMocks: func(
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				_ *mocks.MockLinkRepository,
				_ *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, txFunc func(context.Context) (any, error)) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(false, errors.New("db error"))
			},
			expectedLink: pkg.LinkInfo{},
			expectedErr:  scrapper.ErrInternalError,
		},
		{
			name:   "chat not found",
			chatID: 1,
			link:   "https://github.com/golang/go",
			setupMocks: func(
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				_ *mocks.MockLinkRepository,
				_ *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, txFunc func(context.Context) (any, error)) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(false, nil)
			},
			expectedLink: pkg.LinkInfo{},
			expectedErr:  scrapper.ErrInternalError,
		},
		{
			name:   "link exists error",
			chatID: 1,
			link:   "https://github.com/golang/go",
			setupMocks: func(
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				linkRepo *mocks.MockLinkRepository,
				_ *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, txFunc func(context.Context) (any, error)) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(true, nil)

				linkRepo.EXPECT().
					LinkExists(gomock.Any(), "https://github.com/golang/go").
					Return(false, errors.New("db error"))
			},
			expectedLink: pkg.LinkInfo{},
			expectedErr:  scrapper.ErrInternalError,
		},
		{
			name:   "link not exists",
			chatID: 1,
			link:   "https://github.com/golang/go",
			setupMocks: func(
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				linkRepo *mocks.MockLinkRepository,
				_ *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, txFunc func(context.Context) (any, error)) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(true, nil)

				linkRepo.EXPECT().
					LinkExists(gomock.Any(), "https://github.com/golang/go").
					Return(false, nil)
			},
			expectedLink: pkg.LinkInfo{},
			expectedErr:  scrapper.ErrInternalError,
		},
		{
			name:   "delete link error",
			chatID: 1,
			link:   "https://github.com/golang/go",
			setupMocks: func(
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				linkRepo *mocks.MockLinkRepository,
				_ *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, txFunc func(context.Context) (any, error)) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(true, nil)

				linkRepo.EXPECT().
					LinkExists(gomock.Any(), "https://github.com/golang/go").
					Return(true, nil)

				linkRepo.EXPECT().
					DeleteLink(gomock.Any(), int64(1), "https://github.com/golang/go").
					Return(pkg.LinkInfo{}, errors.New("db error"))
			},
			expectedLink: pkg.LinkInfo{},
			expectedErr:  scrapper.ErrInternalError,
		},
		{
			name:   "cache delete error does not fail request",
			chatID: 1,
			link:   "https://github.com/golang/go",
			setupMocks: func(
				transactor *mocks.MockTransactor,
				chatsRepo *mocks.MockChatRepository,
				linkRepo *mocks.MockLinkRepository,
				cacheRepo *mocks.MockCacheRepository,
			) {
				expectedLink := pkg.LinkInfo{
					Link: "https://github.com/golang/go",
				}

				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, txFunc func(context.Context) (any, error)) (any, error) {
						return txFunc(ctx)
					})

				chatsRepo.EXPECT().
					ChatExists(gomock.Any(), int64(1)).
					Return(true, nil)

				linkRepo.EXPECT().
					LinkExists(gomock.Any(), "https://github.com/golang/go").
					Return(true, nil)

				linkRepo.EXPECT().
					DeleteLink(gomock.Any(), int64(1), "https://github.com/golang/go").
					Return(expectedLink, nil)

				cacheRepo.EXPECT().
					DeleteUserLinks(gomock.Any(), int64(1)).
					Return(errors.New("cache error"))
			},
			expectedLink: pkg.LinkInfo{
				Link: "https://github.com/golang/go",
			},
			expectedErr: nil,
		},
		{
			name:   "transaction returns wrong type",
			chatID: 1,
			link:   "https://github.com/golang/go",
			setupMocks: func(
				transactor *mocks.MockTransactor,
				_ *mocks.MockChatRepository,
				_ *mocks.MockLinkRepository,
				_ *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					Return("wrong type", nil)
			},
			expectedLink: pkg.LinkInfo{},
			expectedErr:  scrapper.ErrInternalError,
		},
		{
			name:   "transaction error",
			chatID: 1,
			link:   "https://github.com/golang/go",
			setupMocks: func(
				transactor *mocks.MockTransactor,
				_ *mocks.MockChatRepository,
				_ *mocks.MockLinkRepository,
				_ *mocks.MockCacheRepository,
			) {
				transactor.EXPECT().
					TransactionWithReturn(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("transaction error"))
			},
			expectedLink: pkg.LinkInfo{},
			expectedErr:  scrapper.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			transactor := mocks.NewMockTransactor(ctrl)
			chatsRepo := mocks.NewMockChatRepository(ctrl)
			linkRepo := mocks.NewMockLinkRepository(ctrl)
			cacheRepo := mocks.NewMockCacheRepository(ctrl)

			tt.setupMocks(transactor, chatsRepo, linkRepo, cacheRepo)

			service := LinksService{
				Transactor: transactor,
				ChatsRepo:  chatsRepo,
				LinkRepo:   linkRepo,
				CacheRepo:  cacheRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			linkInfo, err := service.DeleteLink(context.Background(), tt.chatID, tt.link)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Equal(t, tt.expectedLink, linkInfo)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedLink, linkInfo)
		})
	}
}

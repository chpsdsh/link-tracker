package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/mocks"
	utils "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service/requesterutils"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

func TestHandleIssueUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		gitLink        scrapper.GithubLink
		initialResult  scrapper.LinkProcessingResult
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult scrapper.LinkProcessingResult
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubIssueRequest("https://api.github.com/repos/golang/go/issues").
					Return([]scrapper.GithubIssue{
						{
							Title:     "Issue title",
							Body:      "Issue body",
							CreatedAt: oldTime,
							UpdatedAt: newTime,
							User: struct {
								Login string `json:"login"`
							}{
								Login: "octocat",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return([]int64{1, 2}, nil)

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
				Events: []pkg.LinkUpdate{
					{
						Description: utils.FormatIssue(scrapper.GithubIssue{
							Title:     "Issue title",
							Body:      "Issue body",
							CreatedAt: oldTime,
							UpdatedAt: newTime,
							User: struct {
								Login string `json:"login"`
							}{
								Login: "octocat",
							},
						}),
						URL:       "https://github.com/golang/go",
						TgChatIDs: []int64{1, 2},
					},
				},
			},
		},
		{
			name: "github issue request error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubIssueRequest("https://api.github.com/repos/golang/go/issues").
					Return(nil, errors.New("network error"))

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
		{
			name: "get chat ids error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubIssueRequest("https://api.github.com/repos/golang/go/issues").
					Return([]scrapper.GithubIssue{
						{
							Title:     "Issue title",
							Body:      "Issue body",
							CreatedAt: oldTime,
							UpdatedAt: newTime,
							User: struct {
								Login string `json:"login"`
							}{
								Login: "octocat",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return(nil, errors.New("db error"))

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
			},
		},
		{
			name: "no new issues",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: newTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubIssueRequest("https://api.github.com/repos/golang/go/issues").
					Return([]scrapper.GithubIssue{
						{
							Title:     "Old issue",
							Body:      "Old body",
							CreatedAt: oldTime,
							UpdatedAt: oldTime,
							User: struct {
								Login string `json:"login"`
							}{
								Login: "octocat",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				Repo:       tt.repo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			result := r.handleIssueUpdates(tt.gitLink, tt.linkInfo, tt.initialResult)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestHandlePullRequestsUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		gitLink        scrapper.GithubLink
		initialResult  scrapper.LinkProcessingResult
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult scrapper.LinkProcessingResult
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubPullRequestRequest("https://api.github.com/repos/golang/go/pulls").
					Return([]scrapper.GithubPullRequest{
						{
							Title:     "PR title",
							Body:      "PR body",
							CreatedAt: oldTime,
							UpdatedAt: newTime,
							User: struct {
								Login string `json:"login"`
							}{
								Login: "octocat",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return([]int64{1, 2}, nil)

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
				Events: []pkg.LinkUpdate{
					{
						Description: utils.FormatPullRequest(scrapper.GithubPullRequest{
							Title:     "PR title",
							Body:      "PR body",
							CreatedAt: oldTime,
							UpdatedAt: newTime,
							User: struct {
								Login string `json:"login"`
							}{
								Login: "octocat",
							},
						}),
						URL:       "https://github.com/golang/go",
						TgChatIDs: []int64{1, 2},
					},
				},
			},
		},
		{
			name: "github pull request request error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubPullRequestRequest("https://api.github.com/repos/golang/go/pulls").
					Return(nil, errors.New("network error"))

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
		{
			name: "get chat ids error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubPullRequestRequest("https://api.github.com/repos/golang/go/pulls").
					Return([]scrapper.GithubPullRequest{
						{
							Title:     "PR title",
							Body:      "PR body",
							CreatedAt: oldTime,
							UpdatedAt: newTime,
							User: struct {
								Login string `json:"login"`
							}{
								Login: "octocat",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return(nil, errors.New("db error"))

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
			},
		},
		{
			name: "no new pull requests",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: newTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubPullRequestRequest("https://api.github.com/repos/golang/go/pulls").
					Return([]scrapper.GithubPullRequest{
						{
							Title:     "Old PR",
							Body:      "Old body",
							CreatedAt: oldTime,
							UpdatedAt: oldTime,
							User: struct {
								Login string `json:"login"`
							}{
								Login: "octocat",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				Repo:       tt.repo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			result := r.handlePullRequestsUpdates(tt.gitLink, tt.linkInfo, tt.initialResult)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestHandleRepositoryUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		gitLink        scrapper.GithubLink
		initialResult  scrapper.LinkProcessingResult
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult scrapper.LinkProcessingResult
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubRequest("https://api.github.com/repos/golang/go").
					Return(scrapper.GitHubRepositoryResponse{
						UpdatedAt: newTime,
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return([]int64{1, 2}, nil)

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
				Events: []pkg.LinkUpdate{
					{
						Description: "Repository updated:",
						URL:         "https://github.com/golang/go",
						TgChatIDs:   []int64{1, 2},
					},
				},
			},
		},
		{
			name: "github repository request error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubRequest("https://api.github.com/repos/golang/go").
					Return(scrapper.GitHubRepositoryResponse{}, errors.New("network error"))

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
		{
			name: "get chat ids error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubRequest("https://api.github.com/repos/golang/go").
					Return(scrapper.GitHubRepositoryResponse{
						UpdatedAt: newTime,
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return(nil, errors.New("db error"))

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
			},
		},
		{
			name: "no repository update",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: newTime,
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubRequest("https://api.github.com/repos/golang/go").
					Return(scrapper.GitHubRepositoryResponse{
						UpdatedAt: oldTime,
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				Repo:       tt.repo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			result := r.handleRepositoryUpdates(tt.gitLink, tt.linkInfo, tt.initialResult)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestSendFailureUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		linkInfo    pkg.LinkInfo
		description string
		repo        func() LinkRepository
		outboxRepo  func() OutboxRepository
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link: "https://github.com/golang/go",
			},
			description: "Error parsing github link",
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return([]int64{1, 2}, nil)

				return repo
			},
			outboxRepo: func() OutboxRepository {
				outboxRepo := mocks.NewMockOutboxRepository(ctrl)

				outboxRepo.EXPECT().
					SaveUpdate(gomock.Any(), pkg.LinkUpdate{
						Description: "Error parsing github link",
						TgChatIDs:   []int64{1, 2},
						URL:         "https://github.com/golang/go",
					}).
					Return(nil)

				return outboxRepo
			},
		},
		{
			name: "get chat ids error",
			linkInfo: pkg.LinkInfo{
				Link: "https://github.com/golang/go",
			},
			description: "Error parsing github link",
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return(nil, errors.New("db error"))

				return repo
			},
			outboxRepo: func() OutboxRepository {
				return mocks.NewMockOutboxRepository(ctrl)
			},
		},
		{
			name: "save update error",
			linkInfo: pkg.LinkInfo{
				Link: "https://github.com/golang/go",
			},
			description: "Error parsing github link",
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return([]int64{1, 2}, nil)

				return repo
			},
			outboxRepo: func() OutboxRepository {
				outboxRepo := mocks.NewMockOutboxRepository(ctrl)

				outboxRepo.EXPECT().
					SaveUpdate(gomock.Any(), pkg.LinkUpdate{
						Description: "Error parsing github link",
						TgChatIDs:   []int64{1, 2},
						URL:         "https://github.com/golang/go",
					}).
					Return(errors.New("outbox error"))

				return outboxRepo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			r := LinksRequester{
				Repo:       tt.repo(),
				OutboxRepo: tt.outboxRepo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			r.sendFailureUpdate(tt.linkInfo, tt.description)
		})
	}
}

func TestUpdateTimeAndSendToOutbox(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		link             pkg.LinkInfo
		processingResult scrapper.LinkProcessingResult
		repo             func() LinkRepository
		outboxRepo       func() OutboxRepository
		transactor       func() Transactor
	}{
		{
			name: "success",
			link: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			processingResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
				Events: []pkg.LinkUpdate{
					{
						URL:         "https://github.com/golang/go",
						Description: "issue update",
						TgChatIDs:   []int64{1, 2},
					},
					{
						URL:         "https://github.com/golang/go",
						Description: "pr update",
						TgChatIDs:   []int64{1, 2},
					},
				},
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					UpdateLinksTime(gomock.Any(), newTime, "https://github.com/golang/go").
					Return(nil)

				return repo
			},
			outboxRepo: func() OutboxRepository {
				outboxRepo := mocks.NewMockOutboxRepository(ctrl)

				outboxRepo.EXPECT().
					SaveUpdate(gomock.Any(), pkg.LinkUpdate{
						URL:         "https://github.com/golang/go",
						Description: "issue update",
						TgChatIDs:   []int64{1, 2},
					}).
					Return(nil)

				outboxRepo.EXPECT().
					SaveUpdate(gomock.Any(), pkg.LinkUpdate{
						URL:         "https://github.com/golang/go",
						Description: "pr update",
						TgChatIDs:   []int64{1, 2},
					}).
					Return(nil)

				return outboxRepo
			},
			transactor: func() Transactor {
				tx := mocks.NewMockTransactor(ctrl)

				tx.EXPECT().
					Transaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				return tx
			},
		},
		{
			name: "no update time change",
			link: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			processingResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
				Events: []pkg.LinkUpdate{
					{
						URL:         "https://github.com/golang/go",
						Description: "old update",
						TgChatIDs:   []int64{1},
					},
				},
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			outboxRepo: func() OutboxRepository {
				return mocks.NewMockOutboxRepository(ctrl)
			},
			transactor: func() Transactor {
				return mocks.NewMockTransactor(ctrl)
			},
		},
		{
			name: "update link time error",
			link: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			processingResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
				Events: []pkg.LinkUpdate{
					{
						URL:         "https://github.com/golang/go",
						Description: "issue update",
						TgChatIDs:   []int64{1},
					},
				},
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					UpdateLinksTime(gomock.Any(), newTime, "https://github.com/golang/go").
					Return(errors.New("db error"))

				return repo
			},
			outboxRepo: func() OutboxRepository {
				return mocks.NewMockOutboxRepository(ctrl)
			},
			transactor: func() Transactor {
				tx := mocks.NewMockTransactor(ctrl)

				tx.EXPECT().
					Transaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				return tx
			},
		},
		{
			name: "save outbox error",
			link: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			processingResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
				Events: []pkg.LinkUpdate{
					{
						URL:         "https://github.com/golang/go",
						Description: "issue update",
						TgChatIDs:   []int64{1},
					},
				},
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					UpdateLinksTime(gomock.Any(), newTime, "https://github.com/golang/go").
					Return(nil)

				return repo
			},
			outboxRepo: func() OutboxRepository {
				outboxRepo := mocks.NewMockOutboxRepository(ctrl)

				outboxRepo.EXPECT().
					SaveUpdate(gomock.Any(), pkg.LinkUpdate{
						URL:         "https://github.com/golang/go",
						Description: "issue update",
						TgChatIDs:   []int64{1},
					}).
					Return(errors.New("outbox error"))

				return outboxRepo
			},
			transactor: func() Transactor {
				tx := mocks.NewMockTransactor(ctrl)

				tx.EXPECT().
					Transaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				return tx
			},
		},
		{
			name: "transaction error",
			link: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			processingResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
				Events: []pkg.LinkUpdate{
					{
						URL:         "https://github.com/golang/go",
						Description: "issue update",
						TgChatIDs:   []int64{1},
					},
				},
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			outboxRepo: func() OutboxRepository {
				return mocks.NewMockOutboxRepository(ctrl)
			},
			transactor: func() Transactor {
				tx := mocks.NewMockTransactor(ctrl)

				tx.EXPECT().
					Transaction(gomock.Any(), gomock.Any()).
					Return(errors.New("transaction error"))

				return tx
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			r := LinksRequester{
				Repo:       tt.repo(),
				OutboxRepo: tt.outboxRepo(),
				Transactor: tt.transactor(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			r.updateTimeAndSendToOutbox(tt.link, tt.processingResult)
		})
	}
}

func TestLinksIteration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string

		links []pkg.LinkInfo
		err   error

		numWorkers int

		expectedBatches int
		expectedResult  bool
	}{
		{
			name: "success split into batches",
			links: []pkg.LinkInfo{
				{Link: "1"},
				{Link: "2"},
				{Link: "3"},
				{Link: "4"},
			},
			numWorkers:      2,
			expectedBatches: 2,
			expectedResult:  true,
		},
		{
			name:            "empty links",
			links:           []pkg.LinkInfo{},
			numWorkers:      2,
			expectedBatches: 0,
			expectedResult:  false,
		},
		{
			name:            "repo error",
			err:             errors.New("db error"),
			numWorkers:      2,
			expectedBatches: 0,
			expectedResult:  false,
		},
		{
			name: "links less then workers",
			links: []pkg.LinkInfo{
				{Link: "1"},
				{Link: "2"},
			},
			numWorkers:      5,
			expectedBatches: 1,
			expectedResult:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := mocks.NewMockLinkRepository(ctrl)

			mockRepo.EXPECT().
				GetAllLinks(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.links, tt.err)

			linksChan := make(chan []pkg.LinkInfo, 10)

			r := LinksRequester{
				Repo:       mockRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				BatchSize:  10,
				LinksPool: LinksPool{
					NumWorkers: tt.numWorkers,
					LinksChan:  linksChan,
				},
			}

			result := r.linksIteration(0)

			assert.Equal(t, tt.expectedResult, result)

			collected := [][]pkg.LinkInfo{}

			for len(linksChan) > 0 {
				batch := <-linksChan
				collected = append(collected, batch)
			}

			assert.Len(t, collected, tt.expectedBatches)
		})
	}
}

func TestHandleStackOverflowAnswers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Unix(1000, 0).UTC()
	newTime := time.Unix(2000, 0).UTC()

	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		soLink         scrapper.StackOverflowLink
		initialResult  scrapper.LinkProcessingResult
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult scrapper.LinkProcessingResult
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: oldTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowAnswersRequest("https://api.stackexchange.com/2.3/questions/123/answers?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowAnswersResponse{
						Items: []scrapper.StackOverflowAnswer{
							{
								LastActivityDate: 2000,
								CreationDate:     1500,
								Body:             "answer body",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://stackoverflow.com/questions/123/test").
					Return([]int64{1, 2}, nil)

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
				Events: []pkg.LinkUpdate{
					{
						Description: "Ответ\n\nАвтор: \nСоздан: 1970-01-01T00:25:00Z\n\nanswer body",
						URL:         "https://stackoverflow.com/questions/123/test",
						TgChatIDs:   []int64{1, 2},
					},
				},
			},
		},
		{
			name: "request error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: oldTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowAnswersRequest("https://api.stackexchange.com/2.3/questions/123/answers?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowAnswersResponse{}, errors.New("network error"))

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
		{
			name: "get chat ids error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: oldTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowAnswersRequest("https://api.stackexchange.com/2.3/questions/123/answers?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowAnswersResponse{
						Items: []scrapper.StackOverflowAnswer{
							{
								LastActivityDate: 2000,
								CreationDate:     1500,
								Body:             "answer body",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://stackoverflow.com/questions/123/test").
					Return(nil, errors.New("db error"))

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
			},
		},
		{
			name: "no new answers",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: newTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowAnswersRequest("https://api.stackexchange.com/2.3/questions/123/answers?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowAnswersResponse{
						Items: []scrapper.StackOverflowAnswer{
							{
								LastActivityDate: 1000,
								CreationDate:     900,
								Body:             "old answer body",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				Repo:       tt.repo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			result := r.handleStackOverflowAnswers(tt.soLink, tt.linkInfo, tt.initialResult)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
func TestHandleStackOverflowComments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Unix(1000, 0).UTC()
	newTime := time.Unix(2000, 0).UTC()

	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		soLink         scrapper.StackOverflowLink
		initialResult  scrapper.LinkProcessingResult
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult scrapper.LinkProcessingResult
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: oldTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowCommentsRequest("https://api.stackexchange.com/2.3/questions/123/comments?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowCommentsResponse{
						Items: []scrapper.StackOverflowComment{
							{
								CreationDate: 2000,
								Body:         "comment body",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://stackoverflow.com/questions/123/test").
					Return([]int64{1, 2}, nil)

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
				Events: []pkg.LinkUpdate{
					{
						Description: "Комментарий\n\nАвтор: \nСоздан: 1970-01-01T00:33:20Z\n\ncomment body",
						URL:         "https://stackoverflow.com/questions/123/test",
						TgChatIDs:   []int64{1, 2},
					},
				},
			},
		},
		{
			name: "request error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: oldTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowCommentsRequest("https://api.stackexchange.com/2.3/questions/123/comments?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowCommentsResponse{}, errors.New("network error"))

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
		{
			name: "get chat ids error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: oldTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowCommentsRequest("https://api.stackexchange.com/2.3/questions/123/comments?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowCommentsResponse{
						Items: []scrapper.StackOverflowComment{
							{
								CreationDate: 2000,
								Body:         "comment body",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://stackoverflow.com/questions/123/test").
					Return(nil, errors.New("db error"))

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
			},
		},
		{
			name: "no new comments",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: newTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowCommentsRequest("https://api.stackexchange.com/2.3/questions/123/comments?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowCommentsResponse{
						Items: []scrapper.StackOverflowComment{
							{
								CreationDate: 1000,
								Body:         "old comment body",
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				Repo:       tt.repo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			result := r.handleStackOverflowComments(tt.soLink, tt.linkInfo, tt.initialResult)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestHandleStackOverflowQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Unix(1000, 0).UTC()
	newTime := time.Unix(2000, 0).UTC()

	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		soLink         scrapper.StackOverflowLink
		initialResult  scrapper.LinkProcessingResult
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult scrapper.LinkProcessingResult
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: oldTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowQuestionRequest("https://api.stackexchange.com/2.3/questions/123?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowQuestionResponse{
						Items: []scrapper.StackOverflowQuestion{
							{
								LastActivityDate: 2000,
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://stackoverflow.com/questions/123/test").
					Return([]int64{1, 2}, nil)

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
				Events: []pkg.LinkUpdate{
					{
						Description: "Question updated:",
						URL:         "https://stackoverflow.com/questions/123/test",
						TgChatIDs:   []int64{1, 2},
					},
				},
			},
		},
		{
			name: "request error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: oldTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowQuestionRequest("https://api.stackexchange.com/2.3/questions/123?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowQuestionResponse{}, errors.New("network error"))

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
		{
			name: "get chat ids error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: oldTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowQuestionRequest("https://api.stackexchange.com/2.3/questions/123?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowQuestionResponse{
						Items: []scrapper.StackOverflowQuestion{
							{
								LastActivityDate: 2000,
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://stackoverflow.com/questions/123/test").
					Return(nil, errors.New("db error"))

				return repo
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: newTime,
			},
		},
		{
			name: "no question update",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/questions/123/test",
				LastUpdateTime: newTime,
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			initialResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowQuestionRequest("https://api.stackexchange.com/2.3/questions/123?site=stackoverflow&filter=withbody").
					Return(scrapper.StackOverflowQuestionResponse{
						Items: []scrapper.StackOverflowQuestion{
							{
								LastActivityDate: 1000,
							},
						},
					}, nil)

				return client
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: scrapper.LinkProcessingResult{
				UpdateTime: oldTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				Repo:       tt.repo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			result := r.handleStackOverflowQuestion(tt.soLink, tt.linkInfo, tt.initialResult)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

package service

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service/utils"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

func TestLinksRequester_SendUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		linkInfo pkg.LinkInfo

		linkRepo func() LinkRepository
		client   func() NetworkClient
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			linkRepo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return([]int64{1, 2}, nil)

				return repo
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					SendLinkUpdate(pkg.LinkUpdate{
						Description: "Ссылка обновлена",
						TgChatIDs:   []int64{1, 2},
						URL:         "https://github.com/golang/go",
					}).
					Return(nil)

				return client
			},
		},
		{
			name: "update error",

			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			linkRepo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return(nil, errors.New("db error"))

				return repo
			},
			client: func() NetworkClient {
				return mocks.NewMockNetworkClient(ctrl)
			},
		},
		{
			name: "get chat ids error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			linkRepo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return(nil, errors.New("db error"))

				return repo
			},
			client: func() NetworkClient {
				return mocks.NewMockNetworkClient(ctrl)
			},
		},
		{
			name: "send error",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			linkRepo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").
					Return([]int64{1}, nil)

				return repo
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					SendLinkUpdate(pkg.LinkUpdate{Description: "Ссылка обновлена", TgChatIDs: []int64{1}, URL: "https://github.com/golang/go"}).
					Return(errors.New("network error"))

				return client
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			r := LinksRequester{
				Repo:       tt.linkRepo(),
				Client:     tt.client(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			r.sendUpdate(tt.linkInfo, "Ссылка обновлена")
		})
	}
}

func TestHandleIssueUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		gitLink        scrapper.GithubLink
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult time.Time
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: time.Time{}.UTC(),
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().DoGithubIssueRequest("https://api.github.com/repos/golang/go/issues").
					Return([]scrapper.GithubIssue{{
						Title:     "Issue",
						Body:      "New commit",
						CreatedAt: time.Time{}.UTC(),
						UpdatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC).UTC(),
					}}, nil)

				client.EXPECT().
					SendLinkUpdate(pkg.LinkUpdate{Description: utils.FormatIssue(scrapper.GithubIssue{
						Title:     "Issue",
						Body:      "New commit",
						CreatedAt: time.Time{}.UTC(),
						UpdatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC).UTC(),
					}),
						TgChatIDs: []int64{1, 2},
						URL:       "https://github.com/golang/go"},
					).Return(nil)
				return client
			},

			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)
				repo.EXPECT().GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").Return([]int64{1, 2}, nil)
				return repo
			},
			expectedResult: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "github request failure",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: time.Time{}.UTC(),
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().DoGithubIssueRequest("https://api.github.com/repos/golang/go/issues").
					Return(nil, errors.New("network error"))

				return client
			},
			expectedResult: time.Time{}.UTC(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				Repo:       tt.repo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}
			resultTime := r.handleIssueUpdates(tt.gitLink, tt.linkInfo)

			assert.Equal(t, tt.expectedResult, resultTime)
		})
	}
}

func TestHandlePullRequestsUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		gitLink        scrapper.GithubLink
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult time.Time
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: time.Time{}.UTC(),
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().DoGithubPullRequestRequest("https://api.github.com/repos/golang/go/pulls").
					Return([]scrapper.GithubPullRequest{{
						Title:     "Issue",
						Body:      "New commit",
						CreatedAt: time.Time{}.UTC(),
						UpdatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC).UTC(),
					}}, nil)

				client.EXPECT().
					SendLinkUpdate(pkg.LinkUpdate{Description: utils.FormatPullRequest(scrapper.GithubPullRequest{
						Title:     "Issue",
						Body:      "New commit",
						CreatedAt: time.Time{}.UTC(),
						UpdatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC).UTC(),
					}),
						TgChatIDs: []int64{1, 2},
						URL:       "https://github.com/golang/go"},
					).Return(nil)
				return client
			},

			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)
				repo.EXPECT().GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").Return([]int64{1, 2}, nil)
				return repo
			},
			expectedResult: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "github request failure",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: time.Time{}.UTC(),
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().DoGithubPullRequestRequest("https://api.github.com/repos/golang/go/pulls").
					Return(nil, errors.New("network error"))

				return client
			},
			expectedResult: time.Time{}.UTC(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				Repo:       tt.repo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}
			resultTime := r.handlePullRequestsUpdates(tt.gitLink, tt.linkInfo)

			assert.Equal(t, tt.expectedResult, resultTime)
		})
	}
}

func TestHandleRepositoryUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		gitLink        scrapper.GithubLink
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult time.Time
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: time.Time{}.UTC(),
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().DoGithubRequest("https://api.github.com/repos/golang/go").
					Return(scrapper.GitHubRepositoryResponse{UpdatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)}, nil)

				client.EXPECT().
					SendLinkUpdate(pkg.LinkUpdate{Description: "Repository updated:",
						TgChatIDs: []int64{1, 2},
						URL:       "https://github.com/golang/go"},
					).Return(nil)
				return client
			},

			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)
				repo.EXPECT().GetChatIDsByLink(gomock.Any(), "https://github.com/golang/go").Return([]int64{1, 2}, nil)
				return repo
			},
			expectedResult: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "github request failure",
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: time.Time{}.UTC(),
			},
			gitLink: scrapper.GithubLink{Owner: "golang", Repo: "go"},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().DoGithubRequest("https://api.github.com/repos/golang/go").
					Return(scrapper.GitHubRepositoryResponse{UpdatedAt: time.Time{}.UTC()}, errors.New("network error"))

				return client
			},
			expectedResult: time.Time{}.UTC(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				Repo:       tt.repo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}
			resultTime := r.handleRepositoryUpdates(tt.gitLink, tt.linkInfo)

			assert.Equal(t, tt.expectedResult, resultTime)
		})
	}
}

func TestHandleStackOverflowAnswers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		soLink         scrapper.StackOverflowLink
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult time.Time
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/2.3/questions/123",
				LastUpdateTime: time.Unix(1000, 0).UTC(),
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			client: func() NetworkClient {
				c := mocks.NewMockNetworkClient(ctrl)

				c.EXPECT().
					DoStackOverflowAnswersRequest(gomock.Any()).
					Return(scrapper.StackOverflowAnswersResponse{
						Items: []scrapper.StackOverflowAnswer{
							{LastActivityDate: 2000},
						},
					}, nil)

				c.EXPECT().
					SendLinkUpdate(pkg.LinkUpdate{Description: utils.FormatStackOverflowAnswer(scrapper.StackOverflowAnswer{
						LastActivityDate: 2000,
					}),
						TgChatIDs: []int64{1, 2},
						URL:       "https://stackoverflow.com/2.3/questions/123"},
					).Return(nil)
				return c
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)
				repo.EXPECT().GetChatIDsByLink(gomock.Any(), "https://stackoverflow.com/2.3/questions/123").Return([]int64{1, 2}, nil)
				return repo
			},
			expectedResult: time.Unix(2000, 0).UTC(),
		},
		{
			name: "no newer answers",
			linkInfo: pkg.LinkInfo{
				LastUpdateTime: time.Unix(5000, 0).UTC(),
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			client: func() NetworkClient {
				c := mocks.NewMockNetworkClient(ctrl)
				c.EXPECT().
					DoStackOverflowAnswersRequest(gomock.Any()).
					Return(scrapper.StackOverflowAnswersResponse{
						Items: []scrapper.StackOverflowAnswer{
							{LastActivityDate: 1000},
						},
					}, nil)

				return c
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: time.Unix(5000, 0).UTC(),
		},
		{
			name: "request error",
			linkInfo: pkg.LinkInfo{
				LastUpdateTime: time.Unix(5000, 0).UTC(),
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			client: func() NetworkClient {
				c := mocks.NewMockNetworkClient(ctrl)

				c.EXPECT().
					DoStackOverflowAnswersRequest(gomock.Any()).
					Return(scrapper.StackOverflowAnswersResponse{}, errors.New("error"))

				return c
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: time.Unix(5000, 0).UTC(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				Repo:       tt.repo(),
			}

			result := r.handleStackOverflowAnswers(tt.soLink, tt.linkInfo)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestHandleStackOverflowComments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		soLink         scrapper.StackOverflowLink
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult time.Time
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/2.3/questions/123",
				LastUpdateTime: time.Unix(1000, 0).UTC(),
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			client: func() NetworkClient {
				c := mocks.NewMockNetworkClient(ctrl)

				c.EXPECT().
					DoStackOverflowCommentsRequest(gomock.Any()).
					Return(scrapper.StackOverflowCommentsResponse{
						Items: []scrapper.StackOverflowComment{
							{CreationDate: 2000},
						},
					}, nil)

				c.EXPECT().
					SendLinkUpdate(pkg.LinkUpdate{Description: utils.FormatStackOverflowComment(scrapper.StackOverflowComment{
						CreationDate: 2000,
					}),
						TgChatIDs: []int64{1, 2},
						URL:       "https://stackoverflow.com/2.3/questions/123"},
					).Return(nil)
				return c
			},
			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)
				repo.EXPECT().GetChatIDsByLink(gomock.Any(), "https://stackoverflow.com/2.3/questions/123").Return([]int64{1, 2}, nil)
				return repo
			},
			expectedResult: time.Unix(2000, 0).UTC(),
		},
		{
			name: "no newer answers",
			linkInfo: pkg.LinkInfo{
				LastUpdateTime: time.Unix(5000, 0).UTC(),
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			client: func() NetworkClient {
				c := mocks.NewMockNetworkClient(ctrl)
				c.EXPECT().
					DoStackOverflowCommentsRequest(gomock.Any()).
					Return(scrapper.StackOverflowCommentsResponse{
						Items: []scrapper.StackOverflowComment{
							{CreationDate: 1000},
						},
					}, nil)

				return c
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: time.Unix(5000, 0).UTC(),
		},
		{
			name: "request error",
			linkInfo: pkg.LinkInfo{
				LastUpdateTime: time.Unix(5000, 0).UTC(),
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			client: func() NetworkClient {
				c := mocks.NewMockNetworkClient(ctrl)

				c.EXPECT().
					DoStackOverflowCommentsRequest(gomock.Any()).
					Return(scrapper.StackOverflowCommentsResponse{}, errors.New("error"))

				return c
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: time.Unix(5000, 0).UTC(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				Repo:       tt.repo(),
			}

			result := r.handleStackOverflowComments(tt.soLink, tt.linkInfo)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestHandleStackOverflowQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		linkInfo       pkg.LinkInfo
		soLink         scrapper.StackOverflowLink
		client         func() NetworkClient
		repo           func() LinkRepository
		expectedResult time.Time
	}{
		{
			name: "success",
			linkInfo: pkg.LinkInfo{
				Link:           "https://stackoverflow.com/2.3/questions/123",
				LastUpdateTime: time.Unix(1000, 0).UTC(),
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			client: func() NetworkClient {
				c := mocks.NewMockNetworkClient(ctrl)

				c.EXPECT().
					DoStackOverflowQuestionRequest(gomock.Any()).
					Return(scrapper.StackOverflowQuestionResponse{
						Items: []struct {
							LastActivityDate int64 `json:"last_activity_date"`
						}{
							{LastActivityDate: 4000},
						},
					}, nil)

				c.EXPECT().
					SendLinkUpdate(pkg.LinkUpdate{Description: "Question updated:",
						TgChatIDs: []int64{1, 2},
						URL:       "https://stackoverflow.com/2.3/questions/123"},
					).Return(nil)

				return c
			},

			repo: func() LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)
				repo.EXPECT().GetChatIDsByLink(gomock.Any(), "https://stackoverflow.com/2.3/questions/123").Return([]int64{1, 2}, nil)
				return repo
			},
			expectedResult: time.Unix(4000, 0).UTC(),
		},
		{
			name: "request error",
			linkInfo: pkg.LinkInfo{
				LastUpdateTime: time.Unix(6000, 0).UTC(),
			},
			soLink: scrapper.StackOverflowLink{ID: "123"},
			client: func() NetworkClient {
				c := mocks.NewMockNetworkClient(ctrl)

				c.EXPECT().
					DoStackOverflowQuestionRequest(gomock.Any()).
					Return(scrapper.StackOverflowQuestionResponse{}, errors.New("network error"))

				return c
			},
			repo: func() LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			expectedResult: time.Unix(6000, 0).UTC(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := LinksRequester{
				Client:     tt.client(),
				Repo:       tt.repo(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			result := r.handleStackOverflowQuestion(tt.soLink, tt.linkInfo)
			assert.Equal(t, tt.expectedResult, result)
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

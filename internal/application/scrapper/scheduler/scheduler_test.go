package scheduler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

func TestParseGithubLink(t *testing.T) {
	tests := []struct {
		name        string
		link        string
		expected    scrapper.GithubLink
		expectedErr error
	}{
		{
			name: "repo link",
			link: "https://github.com/golang/go",
			expected: scrapper.GithubLink{
				Type:  scrapper.GithubRepo,
				Owner: "golang",
				Repo:  "go",
			},
		},
		{
			name: "issue link",
			link: "https://github.com/golang/go/issues/123",
			expected: scrapper.GithubLink{
				Type:  scrapper.GithubIssue,
				Owner: "golang",
				Repo:  "go",
				ID:    "123",
			},
		},
		{
			name: "pull link",
			link: "https://github.com/golang/go/pull/77",
			expected: scrapper.GithubLink{
				Type:  scrapper.GithubPull,
				Owner: "golang",
				Repo:  "go",
				ID:    "77",
			},
		},
		{
			name:        "not github host",
			link:        "https://gitlab.com/golang/go",
			expectedErr: ErrNotGitHubURL,
		},
		{
			name:        "invalid github path",
			link:        "https://github.com/golang",
			expectedErr: ErrInvalidGitHubURL,
		},
		{
			name:        "unsupported github url",
			link:        "https://github.com/golang/go/wiki/home",
			expectedErr: ErrUnsupportedGithubURL,
		},
		{
			name:        "not url",
			link:        "://bad-url",
			expectedErr: ErrNotURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseGithubLink(tt.link)

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestParseStackOverflowLink(t *testing.T) {
	tests := []struct {
		name        string
		link        string
		expected    scrapper.StackOverflowLink
		expectedErr error
	}{
		{
			name: "valid question link",
			link: "https://stackoverflow.com/questions/11227809/test",
			expected: scrapper.StackOverflowLink{
				Type: scrapper.StackOverflowQuestion,
				ID:   "11227809",
			},
		},
		{
			name: "valid question link without title",
			link: "https://stackoverflow.com/questions/11227809",
			expected: scrapper.StackOverflowLink{
				Type: scrapper.StackOverflowQuestion,
				ID:   "11227809",
			},
		},
		{
			name:        "not stackoverflow host",
			link:        "https://github.com/questions/11227809",
			expectedErr: ErrNotStackOverflow,
		},
		{
			name:        "invalid path",
			link:        "https://stackoverflow.com/answers/11227809",
			expectedErr: ErrInvalidStackOverflowURL,
		},
		{
			name:        "too short path",
			link:        "https://stackoverflow.com/questions",
			expectedErr: ErrInvalidStackOverflowURL,
		},
		{
			name:        "not url",
			link:        "://bad-url",
			expectedErr: ErrNotURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseStackOverflowLink(tt.link)

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestLinksRequester_SendUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(5 * time.Minute)

	tests := []struct {
		name       string
		updateTime time.Time
		linkInfo   pkg.LinkInfo

		linkRepo func() service.LinkRepository
		client   func() NetworkClient
	}{
		{
			name:       "success",
			updateTime: newTime,
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					UpdateLinksTime(gomock.Any(), newTime, "https://github.com/golang/go").
					Return(nil)

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
			name:       "no update when time equal",
			updateTime: oldTime,
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			linkRepo: func() service.LinkRepository {
				return mocks.NewMockLinkRepository(ctrl)
			},
			client: func() NetworkClient {
				return mocks.NewMockNetworkClient(ctrl)
			},
		},
		{
			name:       "update error",
			updateTime: newTime,
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					UpdateLinksTime(gomock.Any(), newTime, gomock.Any()).
					Return(errors.New("db error"))

				return repo
			},
			client: func() NetworkClient {
				return mocks.NewMockNetworkClient(ctrl)
			},
		},
		{
			name:       "get chat ids error",
			updateTime: newTime,
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					UpdateLinksTime(gomock.Any(), newTime, gomock.Any()).
					Return(nil)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db error"))

				return repo
			},
			client: func() NetworkClient {
				return mocks.NewMockNetworkClient(ctrl)
			},
		},
		{
			name:       "send error",
			updateTime: newTime,
			linkInfo: pkg.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					UpdateLinksTime(gomock.Any(), newTime, gomock.Any()).
					Return(nil)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), gomock.Any()).
					Return([]int64{1}, nil)

				return repo
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					SendLinkUpdate(gomock.Any()).
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

			r.sendUpdate(context.Background(), tt.updateTime, tt.linkInfo)
		})
	}
}

func TestLinksRequester_HandleGithubLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(10 * time.Minute)

	tests := []struct {
		name     string
		linkRepo func() service.LinkRepository
		client   func() NetworkClient
	}{
		{
			name: "ignore non github links",
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), gitHubHost, linksRequestLimit, 0).
					Return([]pkg.LinkInfo{
						{Link: "https://stackoverflow.com/questions/11227809", LastUpdateTime: oldTime},
					}, nil)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), gitHubHost, linksRequestLimit, linksRequestLimit).
					Return(nil, nil)

				return repo
			},
			client: func() NetworkClient {
				return mocks.NewMockNetworkClient(ctrl)
			},
		},
		{
			name: "github link updated successfully",
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				link := "https://github.com/golang/go"

				repo.EXPECT().
					GetAllLinks(gomock.Any(), gitHubHost, linksRequestLimit, 0).
					Return([]pkg.LinkInfo{
						{Link: link, LastUpdateTime: oldTime},
					}, nil)

				repo.EXPECT().
					UpdateLinksTime(gomock.Any(), newTime, link).
					Return(nil)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), link).
					Return([]int64{1, 2}, nil)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), gitHubHost, linksRequestLimit, linksRequestLimit).
					Return(nil, nil)

				return repo
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				apiURL := "https://api.github.com/repos/golang/go"

				client.EXPECT().
					DoGithubRequest(apiURL).
					Return(scrapper.GitHubUpdate{
						UpdatedAt: newTime.Format(time.RFC3339),
					}, nil)

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
			name: "github client error",
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), gitHubHost, linksRequestLimit, 0).
					Return([]pkg.LinkInfo{
						{Link: "https://github.com/golang/go", LastUpdateTime: oldTime},
					}, nil)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), gitHubHost, linksRequestLimit, linksRequestLimit).
					Return(nil, nil)

				return repo
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubRequest(gomock.Any()).
					Return(scrapper.GitHubUpdate{}, errors.New("github error"))

				return client
			},
		},
		{
			name: "invalid update time",
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), gitHubHost, linksRequestLimit, 0).
					Return([]pkg.LinkInfo{
						{Link: "https://github.com/golang/go", LastUpdateTime: oldTime},
					}, nil)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), gitHubHost, linksRequestLimit, linksRequestLimit).
					Return(nil, nil)

				return repo
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoGithubRequest(gomock.Any()).
					Return(scrapper.GitHubUpdate{UpdatedAt: "bad-time"}, nil)

				return client
			},
		},
		{
			name: "repo error",
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), gitHubHost, linksRequestLimit, 0).
					Return(nil, errors.New("db error"))

				return repo
			},
			client: func() NetworkClient {
				return mocks.NewMockNetworkClient(ctrl)
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

			r.HandleGithubLinks()
		})
	}
}

func TestLinksRequester_HandleStackOverflowLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(15 * time.Minute)

	tests := []struct {
		name string

		linkRepo func() service.LinkRepository
		client   func() NetworkClient
	}{
		{
			name: "ignore non stackoverflow links",
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), stackOverflowHost, linksRequestLimit, 0).
					Return([]pkg.LinkInfo{
						{Link: "https://github.com/golang/go", LastUpdateTime: oldTime},
					}, nil)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), stackOverflowHost, linksRequestLimit, linksRequestLimit).
					Return(nil, nil)

				return repo
			},
			client: func() NetworkClient {
				return mocks.NewMockNetworkClient(ctrl)
			},
		},
		{
			name: "stackoverflow link updated successfully",
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				link := "https://stackoverflow.com/questions/11227809/test"

				repo.EXPECT().
					GetAllLinks(gomock.Any(), stackOverflowHost, linksRequestLimit, 0).
					Return([]pkg.LinkInfo{
						{Link: link, LastUpdateTime: oldTime},
					}, nil)

				repo.EXPECT().
					UpdateLinksTime(gomock.Any(), newTime, link).
					Return(nil)

				repo.EXPECT().
					GetChatIDsByLink(gomock.Any(), link).
					Return([]int64{100}, nil)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), stackOverflowHost, linksRequestLimit, linksRequestLimit).
					Return(nil, nil)

				return repo
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				apiURL := "https://api.stackexchange.com/2.3/questions/11227809?site=stackoverflow"

				client.EXPECT().
					DoStackOverflowRequest(apiURL).
					Return(scrapper.StackOverflowUpdate{
						Items: []struct {
							LastActivityDate int64 `json:"last_activity_date"`
						}{
							{LastActivityDate: newTime.Unix()},
						},
					}, nil)

				client.EXPECT().
					SendLinkUpdate(pkg.LinkUpdate{
						Description: "Ссылка обновлена",
						TgChatIDs:   []int64{100},
						URL:         "https://stackoverflow.com/questions/11227809/test",
					}).
					Return(nil)

				return client
			},
		},
		{
			name: "stackoverflow client error",
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), stackOverflowHost, linksRequestLimit, 0).
					Return([]pkg.LinkInfo{
						{Link: "https://stackoverflow.com/questions/11227809/test", LastUpdateTime: oldTime},
					}, nil)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), stackOverflowHost, linksRequestLimit, linksRequestLimit).
					Return(nil, nil)

				return repo
			},
			client: func() NetworkClient {
				client := mocks.NewMockNetworkClient(ctrl)

				client.EXPECT().
					DoStackOverflowRequest(gomock.Any()).
					Return(scrapper.StackOverflowUpdate{}, errors.New("error"))

				return client
			},
		},
		{
			name: "repo error",
			linkRepo: func() service.LinkRepository {
				repo := mocks.NewMockLinkRepository(ctrl)

				repo.EXPECT().
					GetAllLinks(gomock.Any(), stackOverflowHost, linksRequestLimit, 0).
					Return(nil, errors.New("db error"))

				return repo
			},
			client: func() NetworkClient {
				return mocks.NewMockNetworkClient(ctrl)
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

			r.HandleStackOverflowLinks()
		})
	}
}

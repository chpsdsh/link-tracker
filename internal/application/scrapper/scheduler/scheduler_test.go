package scheduler

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
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
			expectedErr: NotGitHubUrlError,
		},
		{
			name:        "invalid github path",
			link:        "https://github.com/golang",
			expectedErr: InvalidGitHubUrlError,
		},
		{
			name:        "unsupported github url",
			link:        "https://github.com/golang/go/wiki/home",
			expectedErr: UnsupportedGithubUrlError,
		},
		{
			name:        "not url",
			link:        "://bad-url",
			expectedErr: NotUrlError,
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
			expectedErr: NotStackOverflowError,
		},
		{
			name:        "invalid path",
			link:        "https://stackoverflow.com/answers/11227809",
			expectedErr: InvalidStackOverflowUrlError,
		},
		{
			name:        "too short path",
			link:        "https://stackoverflow.com/questions",
			expectedErr: InvalidStackOverflowUrlError,
		},
		{
			name:        "not url",
			link:        "://bad-url",
			expectedErr: NotUrlError,
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

func TestLinksRequesterSendUpdate(t *testing.T) {
	oldTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(5 * time.Minute)

	tests := []struct {
		name            string
		updateTime      time.Time
		linkInfo        shared.LinkInfo
		expectUpdate    bool
		chatIDs         []int64
		sendUpdateError error
	}{
		{
			name:       "link updated and notification sent",
			updateTime: newTime,
			linkInfo: shared.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			expectUpdate: true,
			chatIDs:      []int64{1, 2},
		},
		{
			name:       "link not updated when time equal",
			updateTime: oldTime,
			linkInfo: shared.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			expectUpdate: false,
		},
		{
			name:       "link not updated when time older",
			updateTime: oldTime.Add(-time.Minute),
			linkInfo: shared.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			expectUpdate: false,
		},
		{
			name:       "send update error",
			updateTime: newTime,
			linkInfo: shared.LinkInfo{
				Link:           "https://github.com/golang/go",
				LastUpdateTime: oldTime,
			},
			expectUpdate:    true,
			chatIDs:         []int64{10},
			sendUpdateError: errors.New("network error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockNetworkClient(ctrl)
			mockRepo := mocks.NewMockRepository(ctrl)

			r := LinksRequester{
				Client:     mockClient,
				Repo:       mockRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			if tt.expectUpdate {
				mockRepo.EXPECT().
					UpdateLinksTime(tt.updateTime, tt.linkInfo)

				mockRepo.EXPECT().
					GetChatIdsByLink(tt.linkInfo.Link).
					Return(tt.chatIDs)

				mockClient.EXPECT().
					SendLinkUpdate(shared.LinkUpdate{
						Description: "Ссылка обновлена",
						TgChatIds:   tt.chatIDs,
						Url:         tt.linkInfo.Link,
					}).
					Return(tt.sendUpdateError)
			}

			r.sendUpdate(tt.updateTime, tt.linkInfo)
		})
	}
}

func TestLinksRequesterHandleGithubLinks(t *testing.T) {
	oldTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(10 * time.Minute)

	tests := []struct {
		name  string
		links []shared.LinkInfo
		setup func(mockClient *mocks.MockNetworkClient, mockRepo *mocks.MockRepository)
	}{
		{
			name: "ignore non github links",
			links: []shared.LinkInfo{
				{Link: "https://stackoverflow.com/questions/11227809", LastUpdateTime: oldTime},
			},
			setup: func(mockClient *mocks.MockNetworkClient, mockRepo *mocks.MockRepository) {},
		},
		{
			name: "github link updated successfully",
			links: []shared.LinkInfo{
				{Link: "https://github.com/golang/go", LastUpdateTime: oldTime},
			},
			setup: func(mockClient *mocks.MockNetworkClient, mockRepo *mocks.MockRepository) {
				linkInfo := shared.LinkInfo{
					Link:           "https://github.com/golang/go",
					LastUpdateTime: oldTime,
				}

				mockClient.EXPECT().
					DoGithubRequest("https://api.github.com/repos/golang/go").
					Return(scrapper.GitHubUpdate{UpdatedAt: newTime.Format(time.RFC3339)}, nil)

				mockRepo.EXPECT().
					UpdateLinksTime(newTime, linkInfo)

				mockRepo.EXPECT().
					GetChatIdsByLink("https://github.com/golang/go").
					Return([]int64{1, 2})

				mockClient.EXPECT().
					SendLinkUpdate(shared.LinkUpdate{
						Description: "Ссылка обновлена",
						TgChatIds:   []int64{1, 2},
						Url:         "https://github.com/golang/go",
					}).
					Return(nil)
			},
		},
		{
			name: "github client error",
			links: []shared.LinkInfo{
				{Link: "https://github.com/golang/go", LastUpdateTime: oldTime},
			},
			setup: func(mockClient *mocks.MockNetworkClient, mockRepo *mocks.MockRepository) {
				mockClient.EXPECT().
					DoGithubRequest("https://api.github.com/repos/golang/go").
					Return(scrapper.GitHubUpdate{}, errors.New("github error"))
			},
		},
		{
			name: "github invalid update time",
			links: []shared.LinkInfo{
				{Link: "https://github.com/golang/go", LastUpdateTime: oldTime},
			},
			setup: func(mockClient *mocks.MockNetworkClient, mockRepo *mocks.MockRepository) {
				mockClient.EXPECT().
					DoGithubRequest("https://api.github.com/repos/golang/go").
					Return(scrapper.GitHubUpdate{UpdatedAt: "bad-time"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockNetworkClient(ctrl)
			mockRepo := mocks.NewMockRepository(ctrl)

			r := LinksRequester{
				Client:     mockClient,
				Repo:       mockRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			mockRepo.EXPECT().
				GetAllLinks().
				Return(tt.links)

			tt.setup(mockClient, mockRepo)

			r.HandleGithubLinks()
		})
	}
}

func TestLinksRequesterHandleStackOverflowLinks(t *testing.T) {
	oldTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(15 * time.Minute)

	tests := []struct {
		name  string
		links []shared.LinkInfo
		setup func(mockClient *mocks.MockNetworkClient, mockRepo *mocks.MockRepository)
	}{
		{
			name: "ignore non stackoverflow links",
			links: []shared.LinkInfo{
				{Link: "https://github.com/golang/go", LastUpdateTime: oldTime},
			},
			setup: func(mockClient *mocks.MockNetworkClient, mockRepo *mocks.MockRepository) {},
		},
		{
			name: "stackoverflow link updated successfully",
			links: []shared.LinkInfo{
				{Link: "https://stackoverflow.com/questions/11227809/test", LastUpdateTime: oldTime},
			},
			setup: func(mockClient *mocks.MockNetworkClient, mockRepo *mocks.MockRepository) {
				linkInfo := shared.LinkInfo{
					Link:           "https://stackoverflow.com/questions/11227809/test",
					LastUpdateTime: oldTime,
				}

				mockClient.EXPECT().
					DoStackOverflowRequest("https://api.stackexchange.com/2.3/questions/11227809?site=stackoverflow").
					Return(scrapper.StackOverflowUpdate{
						Items: []struct {
							LastActivityDate int64 `json:"last_activity_date"`
						}{
							{
								LastActivityDate: newTime.Unix(),
							},
						},
					}, nil)

				mockRepo.EXPECT().
					UpdateLinksTime(newTime, linkInfo)

				mockRepo.EXPECT().
					GetChatIdsByLink("https://stackoverflow.com/questions/11227809/test").
					Return([]int64{100})

				mockClient.EXPECT().
					SendLinkUpdate(shared.LinkUpdate{
						Description: "Ссылка обновлена",
						TgChatIds:   []int64{100},
						Url:         "https://stackoverflow.com/questions/11227809/test",
					}).
					Return(nil)
			},
		},
		{
			name: "stackoverflow client error",
			links: []shared.LinkInfo{
				{Link: "https://stackoverflow.com/questions/11227809/test", LastUpdateTime: oldTime},
			},
			setup: func(mockClient *mocks.MockNetworkClient, mockRepo *mocks.MockRepository) {
				mockClient.EXPECT().
					DoStackOverflowRequest("https://api.stackexchange.com/2.3/questions/11227809?site=stackoverflow").
					Return(scrapper.StackOverflowUpdate{}, errors.New("stackoverflow error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockNetworkClient(ctrl)
			mockRepo := mocks.NewMockRepository(ctrl)

			r := LinksRequester{
				Client:     mockClient,
				Repo:       mockRepo,
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			mockRepo.EXPECT().
				GetAllLinks().
				Return(tt.links)

			tt.setup(mockClient, mockRepo)

			r.HandleStackOverflowLinks()
		})
	}
}

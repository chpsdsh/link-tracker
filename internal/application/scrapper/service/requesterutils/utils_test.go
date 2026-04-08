package requesterutils

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
				Owner: "golang",
				Repo:  "go",
			},
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
				ID: "11227809",
			},
		},
		{
			name: "valid question link without title",
			link: "https://stackoverflow.com/questions/11227809",
			expected: scrapper.StackOverflowLink{
				ID: "11227809",
			},
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

func TestFormatPullRequest(t *testing.T) {
	tests := []struct {
		name string
		pr   scrapper.GithubPullRequest

		expectedContains []string
	}{
		{
			name: "success without truncation",
			pr: scrapper.GithubPullRequest{
				Title:     "Fix bug",
				Body:      "Short description",
				CreatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC),
				User: struct {
					Login string `json:"login"`
				}{Login: "alice"},
			},
			expectedContains: []string{
				"Pull Request",
				"Fix bug",
				"alice",
				"Short description",
			},
		},
		{
			name: "body is truncated",
			pr: scrapper.GithubPullRequest{
				Title:     "Big PR",
				Body:      strings.Repeat("a", 210),
				CreatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC),
				User: struct {
					Login string `json:"login"`
				}{Login: "bob"},
			},
			expectedContains: []string{
				"...",
			},
		},
		{
			name: "empty body",
			pr: scrapper.GithubPullRequest{
				Title:     "Empty body",
				Body:      "",
				CreatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC),
				User: struct {
					Login string `json:"login"`
				}{Login: "charlie"},
			},
			expectedContains: []string{
				"Empty body",
				"charlie",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPullRequest(tt.pr)

			for _, substr := range tt.expectedContains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestFormatIssue(t *testing.T) {
	tests := []struct {
		name string
		iss  scrapper.GithubIssue

		expectedContains []string
	}{
		{
			name: "success without truncation",
			iss: scrapper.GithubIssue{
				Title:     "Crash bug",
				Body:      "Some issue description",
				CreatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC),
				User: struct {
					Login string `json:"login"`
				}{Login: "dave"},
			},
			expectedContains: []string{
				"Issue",
				"Crash bug",
				"dave",
				"Some issue description",
			},
		},
		{
			name: "body is truncated",
			iss: scrapper.GithubIssue{
				Title:     "Big issue",
				Body:      strings.Repeat("b", 210),
				CreatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC),
				User: struct {
					Login string `json:"login"`
				}{Login: "eve"},
			},
			expectedContains: []string{
				"...",
			},
		},
		{
			name: "empty body",
			iss: scrapper.GithubIssue{
				Title:     "No body",
				Body:      "",
				CreatedAt: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC),
				User: struct {
					Login string `json:"login"`
				}{Login: "frank"},
			},
			expectedContains: []string{
				"No body",
				"frank",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatIssue(tt.iss)

			for _, substr := range tt.expectedContains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestFormatStackOverflowAnswer(t *testing.T) {
	tests := []struct {
		name             string
		input            scrapper.StackOverflowAnswer
		expectedContains []string
	}{
		{
			name: "success with html",
			input: scrapper.StackOverflowAnswer{
				CreationDate: 1710000000,
				Body:         "<p>Hello <b>world</b></p>",
				Owner: struct {
					DisplayName string `json:"display_name"`
				}{
					DisplayName: "user1",
				},
			},
			expectedContains: []string{
				"Ответ",
				"user1",
				"Hello world",
			},
		},
		{
			name: "body truncated",
			input: scrapper.StackOverflowAnswer{
				CreationDate: 1710000000,
				Body:         strings.Repeat("a", descriptionMaxLength+10),
				Owner: struct {
					DisplayName string `json:"display_name"`
				}{
					DisplayName: "user2",
				},
			},
			expectedContains: []string{
				"...",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatStackOverflowAnswer(tt.input)

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestFormatStackOverflowComment(t *testing.T) {
	tests := []struct {
		name             string
		input            scrapper.StackOverflowComment
		expectedContains []string
	}{
		{
			name: "success with html",
			input: scrapper.StackOverflowComment{
				CreationDate: 1710000000,
				Body:         "<p>Comment <i>text</i></p>",
				Owner: struct {
					DisplayName string `json:"display_name"`
				}{
					DisplayName: "comment_user",
				},
			},
			expectedContains: []string{
				"Комментарий",
				"comment_user",
				"Comment text",
			},
		},
		{
			name: "body truncated",
			input: scrapper.StackOverflowComment{
				CreationDate: 1710000000,
				Body:         strings.Repeat("b", descriptionMaxLength+5),
				Owner: struct {
					DisplayName string `json:"display_name"`
				}{
					DisplayName: "user2",
				},
			},
			expectedContains: []string{
				"...",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatStackOverflowComment(tt.input)

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

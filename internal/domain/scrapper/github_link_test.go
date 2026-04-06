package scrapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubLink_ConvertToURL(t *testing.T) {
	tests := []struct {
		name     string
		link     GithubLink
		option   GithubLinkOption
		expected string
	}{
		{
			name:     "repository url",
			link:     GithubLink{Owner: "golang", Repo: "go"},
			option:   GithubLinkOptionRepository,
			expected: "https://api.github.com/repos/golang/go",
		},
		{
			name:     "issues url",
			link:     GithubLink{Owner: "golang", Repo: "go"},
			option:   GithubLinkOptionIssue,
			expected: "https://api.github.com/repos/golang/go/issues",
		},
		{
			name:     "pull requests url",
			link:     GithubLink{Owner: "golang", Repo: "go"},
			option:   GithubLinkPullRequest,
			expected: "https://api.github.com/repos/golang/go/pulls",
		},
		{
			name:     "unknown option",
			link:     GithubLink{Owner: "golang", Repo: "go"},
			option:   GithubLinkOption(999),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.link.ConvertToURL(tt.option)
			assert.Equal(t, tt.expected, result)
		})
	}
}

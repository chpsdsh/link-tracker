package scrapper

import "fmt"

type GithubLinkOption int

const (
	GithubLinkOptionRepository GithubLinkOption = iota
	GithubLinkOptionIssue
	GithubLinkPullRequest
)

type GithubLink struct {
	Owner string
	Repo  string
}

func (g GithubLink) ConvertToURL(option GithubLinkOption) string {
	var url string
	switch option {
	case GithubLinkOptionRepository:
		url = fmt.Sprintf(
			"https://api.github.com/repos/%s/%s",
			g.Owner,
			g.Repo,
		)
	case GithubLinkOptionIssue:
		url = fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/issues",
			g.Owner,
			g.Repo,
		)
	case GithubLinkPullRequest:
		url = fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/pulls",
			g.Owner,
			g.Repo,
		)
	}
	return url
}

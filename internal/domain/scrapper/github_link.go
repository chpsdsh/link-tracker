package scrapper

import "fmt"

type GithubLinkType int

const (
	GithubRepo GithubLinkType = iota
	GithubIssue
	GithubPull
)

type GithubLink struct {
	Type  GithubLinkType
	Owner string
	Repo  string
	ID    string
}

func (g GithubLink) ConvertToUrl() string {
	var url string
	switch g.Type {
	case GithubRepo:
		url = fmt.Sprintf(
			"https://api.github.com/repos/%s/%s",
			g.Owner,
			g.Repo,
		)
	case GithubIssue:
		url = fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/issues/%s",
			g.Owner,
			g.Repo,
			g.ID,
		)
	case GithubPull:
		url = fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/pulls/%s",
			g.Owner,
			g.Repo,
			g.ID,
		)
	}
	return url
}

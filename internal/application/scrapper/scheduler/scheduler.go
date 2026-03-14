//go:generate mockgen -source scheduler.go -destination=../mocks/scheduler_mocks.go -package=mocks
package scheduler

import (
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

var (
	ErrNotGitHubURL               = errors.New("not github")
	ErrInvalidGitHubURL           = errors.New("invalid github url")
	ErrIncorrectRequestParameters = errors.New("incorrect request parameters")
	ErrNotStackOverflow           = errors.New("not StackOverflow")
	ErrInvalidStackOverflowURL    = errors.New("invalid StackOverflow url")
	ErrUnsupportedGithubURL       = errors.New("unsupported github url")
	ErrNotURL                     = errors.New("not url")
)

const (
	linkTrackInterval         = time.Second * 10
	gitHubHost                = "github.com"
	gitHubIssues              = "issues"
	gitHubPullRequests        = "pull"
	stackOverflowHost         = "stackoverflow.com"
	stackOverflowQuestions    = "questions"
	minGithubURLParts         = 2
	gitHubIssueURLParts       = 4
	gitHubPullRequestURLParts = 4
	minStackOverflowURLParts  = 2
)

type NetworkClient interface {
	DoGithubRequest(url string) (scrapper.GitHubUpdate, error)
	SendLinkUpdate(update pkg.LinkUpdate) error
	DoStackOverflowRequest(url string) (scrapper.StackOverflowUpdate, error)
}

type LinksRequester struct {
	Client     NetworkClient
	Scheduler  gocron.Scheduler
	Repo       handler.Repository
	BaseLogger *slog.Logger
}

func (r LinksRequester) StartLinkRequester() {
	_, err := r.Scheduler.NewJob(
		gocron.DurationJob(
			linkTrackInterval,
		),
		gocron.NewTask(
			r.HandleStackOverflowLinks,
		),
	)
	if err != nil {
		return
	}

	_, err = r.Scheduler.NewJob(
		gocron.DurationJob(
			linkTrackInterval,
		),
		gocron.NewTask(
			r.HandleGithubLinks,
		),
	)
	if err != nil {
		return
	}

	r.Scheduler.Start()
}

func (r LinksRequester) HandleGithubLinks() {
	links := r.Repo.GetAllLinks()
	for _, l := range links {
		link, err := ParseGithubLink(l.Link)
		if err != nil {
			continue
		}

		gitUpdate, err := r.Client.DoGithubRequest(link.ConvertToURL())
		if err != nil {
			r.BaseLogger.Error("error during github quarry", slog.String("error", err.Error()))
			continue
		}

		updateTime, err := time.Parse(time.RFC3339, gitUpdate.UpdatedAt)
		if err != nil {
			r.BaseLogger.Error("error during update", slog.String("error", err.Error()))
			continue
		}

		r.sendUpdate(updateTime, l)
	}
}

func (r LinksRequester) HandleStackOverflowLinks() {
	links := r.Repo.GetAllLinks()
	for _, l := range links {
		link, err := ParseStackOverflowLink(l.Link)
		if err != nil {
			continue
		}
		stackUpdate, err := r.Client.DoStackOverflowRequest(link.ConvertToURL())
		if err != nil {
			r.BaseLogger.Error("error during stack overflow", slog.String("error", err.Error()))
			continue
		}

		updateTime := time.Unix(stackUpdate.Items[0].LastActivityDate, 0).UTC()

		r.sendUpdate(updateTime, l)
	}
}

func (r LinksRequester) sendUpdate(updateTime time.Time, linkInfo pkg.LinkInfo) {
	if updateTime.After(linkInfo.LastUpdateTime) {
		r.Repo.UpdateLinksTime(updateTime, linkInfo)

		chatIDs := r.Repo.GetChatIDsByLink(linkInfo.Link)
		update := pkg.LinkUpdate{Description: "Ссылка обновлена", TgChatIDs: chatIDs, URL: linkInfo.Link}

		err := r.Client.SendLinkUpdate(update)
		if err != nil {
			r.BaseLogger.Error("error sending link update", slog.String("error", err.Error()))
			return
		}
		r.BaseLogger.Info("link is sent to chats", slog.String("link", linkInfo.Link), slog.Any("chats", chatIDs))
	}
}

func ParseGithubLink(link string) (scrapper.GithubLink, error) {
	u, err := url.Parse(link)
	if err != nil {
		return scrapper.GithubLink{}, errors.Join(err, ErrNotURL)
	}

	if u.Host != gitHubHost {
		return scrapper.GithubLink{}, ErrNotGitHubURL
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")

	if len(parts) < minGithubURLParts {
		return scrapper.GithubLink{}, ErrInvalidGitHubURL
	}

	owner := parts[0]
	repo := parts[1]

	switch {
	case len(parts) == minGithubURLParts:
		return scrapper.GithubLink{
			Type:  scrapper.GithubRepo,
			Owner: owner,
			Repo:  repo,
		}, nil
	case len(parts) == gitHubIssueURLParts && parts[2] == gitHubIssues:
		return scrapper.GithubLink{
			Type:  scrapper.GithubIssue,
			Owner: owner,
			Repo:  repo,
			ID:    parts[3],
		}, nil
	case len(parts) == gitHubPullRequestURLParts && parts[2] == gitHubPullRequests:
		return scrapper.GithubLink{
			Type:  scrapper.GithubPull,
			Owner: owner,
			Repo:  repo,
			ID:    parts[3],
		}, nil
	}

	return scrapper.GithubLink{}, ErrUnsupportedGithubURL
}

func ParseStackOverflowLink(link string) (scrapper.StackOverflowLink, error) {
	u, err := url.Parse(link)
	if err != nil {
		return scrapper.StackOverflowLink{}, errors.Join(err, ErrNotURL)
	}

	if u.Host != stackOverflowHost {
		return scrapper.StackOverflowLink{}, ErrNotStackOverflow
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")

	if len(parts) < minStackOverflowURLParts {
		return scrapper.StackOverflowLink{}, ErrInvalidStackOverflowURL
	}

	if parts[0] != stackOverflowQuestions {
		return scrapper.StackOverflowLink{}, ErrInvalidStackOverflowURL
	}

	id := parts[1]

	return scrapper.StackOverflowLink{
		Type: scrapper.StackOverflowQuestion,
		ID:   id,
	}, nil
}

//go:generate mockgen -source scheduler.go -destination=../mocks/scheduler_mocks.go -package=mocks
package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service"
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
	linksHandleDuration       = time.Second * 10
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
	linksRequestLimit         = 5
)

type NetworkClient interface {
	DoGithubRequest(url string) (scrapper.GitHubUpdate, error)
	SendLinkUpdate(update pkg.LinkUpdate) error
	DoStackOverflowRequest(url string) (scrapper.StackOverflowUpdate, error)
}

type LinksRequester struct {
	Client     NetworkClient
	Scheduler  gocron.Scheduler
	Repo       service.LinkRepository
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
	offset := 0
	for r.githubIteration(offset) {
		offset += linksRequestLimit
	}

}

func (r LinksRequester) githubIteration(offset int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), linksHandleDuration)
	defer cancel()

	links, err := r.Repo.GetAllLinks(ctx, gitHubHost, linksRequestLimit, offset)

	if err != nil {
		r.BaseLogger.Error("error getting github links", slog.String("error", err.Error()))
		return false
	}

	if len(links) == 0 {
		return false
	}

	for _, l := range links {
		link, parseErr := ParseGithubLink(l.Link)
		if parseErr != nil {
			continue
		}

		gitUpdate, sendErr := r.Client.DoGithubRequest(link.ConvertToURL())
		if sendErr != nil {
			r.BaseLogger.Error("error during github query", slog.String("error", sendErr.Error()))
			continue
		}

		updateTime, parseErr := time.Parse(time.RFC3339, gitUpdate.UpdatedAt)
		if parseErr != nil {
			r.BaseLogger.Error("error during update", slog.String("error", parseErr.Error()))
			continue
		}

		r.sendUpdate(ctx, updateTime, l)
	}
	return true
}

func (r LinksRequester) HandleStackOverflowLinks() {
	offset := 0
	for r.stackOverflowIteration(offset) {
		offset += linksRequestLimit
	}
}

func (r LinksRequester) stackOverflowIteration(offset int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), linksHandleDuration)
	defer cancel()

	links, err := r.Repo.GetAllLinks(ctx, stackOverflowHost, linksRequestLimit, offset)

	if err != nil {
		r.BaseLogger.Error("error getting stack overflow links", slog.String("error", err.Error()))
		return false
	}

	if len(links) == 0 {
		return false
	}

	for _, l := range links {
		link, parseErr := ParseStackOverflowLink(l.Link)
		if parseErr != nil {
			continue
		}
		stackUpdate, sendErr := r.Client.DoStackOverflowRequest(link.ConvertToURL())
		if sendErr != nil {
			r.BaseLogger.Error("error during stack overflow", slog.String("error", sendErr.Error()))
			continue
		}

		updateTime := time.Unix(stackUpdate.Items[0].LastActivityDate, 0).UTC()

		r.sendUpdate(ctx, updateTime, l)
	}
	return true
}

func (r LinksRequester) sendUpdate(ctx context.Context, updateTime time.Time, linkInfo pkg.LinkInfo) {
	if updateTime.After(linkInfo.LastUpdateTime) {
		if err := r.Repo.UpdateLinksTime(ctx, updateTime, linkInfo.Link); err != nil {
			r.BaseLogger.Error("error updating links", slog.String("error", err.Error()))
			return
		}

		chatIDs, err := r.Repo.GetChatIDsByLink(ctx, linkInfo.Link)
		if err != nil {
			r.BaseLogger.Error("error getting chatIDs", slog.String("error", err.Error()))
			return
		}

		update := pkg.LinkUpdate{Description: "Ссылка обновлена", TgChatIDs: chatIDs, URL: linkInfo.Link}

		if err = r.Client.SendLinkUpdate(update); err != nil {
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

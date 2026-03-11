package scheduler

import (
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

var (
	notGitHubUrlError               = errors.New("not github")
	invalidGitHubUrlError           = errors.New("invalid github url")
	IncorrectRequestParametersError = errors.New("incorrect request parameters")
)

const (
	linkTrackInterval  = time.Second * 10
	gitHubHost         = "github.com"
	gitHubIssues       = "issues"
	gitHubPullRequests = "pull"
)

type NetworkClient interface {
	DoGithubRequest(url string) (scrapper.GitHubUpdate, error)
	SendLinkUpdate(update shared.LinkUpdate) error
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

		gitUpdate, err := r.Client.DoGithubRequest(link.ConvertToUrl())
		if err != nil {
			r.BaseLogger.Error("error during github quarry", slog.String("error", err.Error()))
			continue
		}

		updateTime, err := time.Parse(time.RFC3339, gitUpdate.UpdatedAt)
		if err != nil {
			r.BaseLogger.Error("error during update", slog.String("error", err.Error()))
			continue
		}

		diff := l.LastUpdateTime.Sub(updateTime)
		r.BaseLogger.Info("times", slog.Any("updateTime", updateTime), slog.Any("last", l.LastUpdateTime), slog.Any("diff", diff))
		if updateTime.After(l.LastUpdateTime) {
			r.Repo.UpdateLinksTime(updateTime, l)

			chatIds := r.Repo.GetChatIdsByLink(l.Link)
			update := shared.LinkUpdate{Description: "link updated", TgChatIds: chatIds, Url: l.Link}

			err := r.Client.SendLinkUpdate(update)
			if err != nil {
				r.BaseLogger.Error("error sending link update", slog.String("error", err.Error()))
				continue
			}
			r.BaseLogger.Info("link is sent to chats", slog.String("link", l.Link), slog.Any("chats", chatIds))
		}
	}
}

func (r LinksRequester) HandleStackOverflowLinks() {
	links := r.Repo.GetAllLinks()
	for _, link := range links {
		if strings.Contains(link.Link, "stackoverflow") {

		}
	}
}

func ParseGithubLink(link string) (scrapper.GithubLink, error) {
	u, err := url.Parse(link)
	if err != nil {
		return scrapper.GithubLink{}, err
	}

	if u.Host != gitHubHost {
		return scrapper.GithubLink{}, notGitHubUrlError
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")

	if len(parts) < 2 {
		return scrapper.GithubLink{}, invalidGitHubUrlError
	}

	owner := parts[0]
	repo := parts[1]

	switch {
	case len(parts) == 2:
		return scrapper.GithubLink{
			Type:  scrapper.GithubRepo,
			Owner: owner,
			Repo:  repo,
		}, nil
	case len(parts) == 4 && parts[2] == gitHubIssues:
		return scrapper.GithubLink{
			Type:  scrapper.GithubIssue,
			Owner: owner,
			Repo:  repo,
			ID:    parts[3],
		}, nil
	case len(parts) == 4 && parts[2] == gitHubPullRequests:
		return scrapper.GithubLink{
			Type:  scrapper.GithubPull,
			Owner: owner,
			Repo:  repo,
			ID:    parts[3],
		}, nil
	}

	return scrapper.GithubLink{}, errors.New("unsupported github url")
}

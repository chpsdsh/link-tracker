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
	NotGitHubUrlError               = errors.New("not github")
	InvalidGitHubUrlError           = errors.New("invalid github url")
	IncorrectRequestParametersError = errors.New("incorrect request parameters")
	NotStackOverflowError           = errors.New("not StackOverflow")
	InvalidStackOverflowUrlError    = errors.New("invalid StackOverflow url")
)

const (
	linkTrackInterval      = time.Second * 10
	gitHubHost             = "github.com"
	gitHubIssues           = "issues"
	gitHubPullRequests     = "pull"
	stackOverflowHost      = "stackoverflow.com"
	stackOverflowQuestions = "questions"
)

type NetworkClient interface {
	DoGithubRequest(url string) (scrapper.GitHubUpdate, error)
	SendLinkUpdate(update shared.LinkUpdate) error
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

		r.sendUpdate(updateTime, l)
	}
}

func (r LinksRequester) sendUpdate(updateTime time.Time, linkInfo shared.LinkInfo) {
	if updateTime.After(linkInfo.LastUpdateTime) {
		r.Repo.UpdateLinksTime(updateTime, linkInfo)

		chatIds := r.Repo.GetChatIdsByLink(linkInfo.Link)
		update := shared.LinkUpdate{Description: "link updated", TgChatIds: chatIds, Url: linkInfo.Link}

		err := r.Client.SendLinkUpdate(update)
		if err != nil {
			r.BaseLogger.Error("error sending link update", slog.String("error", err.Error()))
			return
		}
		r.BaseLogger.Info("link is sent to chats", slog.String("link", linkInfo.Link), slog.Any("chats", chatIds))
	}
}

func (r LinksRequester) HandleStackOverflowLinks() {
	links := r.Repo.GetAllLinks()
	for _, l := range links {
		link, err := ParseStackOverflowLink(l.Link)
		if err != nil {
			continue
		}
		stackUpdate, err := r.Client.DoStackOverflowRequest(link.ConvertToUrl())
		if err != nil {
			r.BaseLogger.Error("error during stack overflow", slog.String("error", err.Error()))
			continue
		}

		updateTime := time.Unix(stackUpdate.LastActivityDate, 0)

		r.sendUpdate(updateTime, l)
	}
}

func ParseGithubLink(link string) (scrapper.GithubLink, error) {
	u, err := url.Parse(link)
	if err != nil {
		return scrapper.GithubLink{}, err
	}

	if u.Host != gitHubHost {
		return scrapper.GithubLink{}, NotGitHubUrlError
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")

	if len(parts) < 2 {
		return scrapper.GithubLink{}, InvalidGitHubUrlError
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

func ParseStackOverflowLink(link string) (scrapper.StackOverflowLink, error) {
	u, err := url.Parse(link)
	if err != nil {
		return scrapper.StackOverflowLink{}, err
	}

	if u.Host != stackOverflowHost {
		return scrapper.StackOverflowLink{}, NotStackOverflowError
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")

	if len(parts) < 2 {
		return scrapper.StackOverflowLink{}, InvalidStackOverflowUrlError
	}

	if parts[0] != stackOverflowQuestions {
		return scrapper.StackOverflowLink{}, InvalidStackOverflowUrlError
	}

	id := parts[1]

	return scrapper.StackOverflowLink{
		Type: scrapper.StackOverflowQuestion,
		ID:   id,
	}, nil
}

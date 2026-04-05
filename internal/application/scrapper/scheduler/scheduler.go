package scheduler

import (
	"errors"
	"time"

	"github.com/go-co-op/gocron/v2"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service"
)

var (
	ErrNotGitHubURL            = errors.New("not github")
	ErrInvalidGitHubURL        = errors.New("invalid github url")
	ErrNotStackOverflow        = errors.New("not StackOverflow")
	ErrInvalidStackOverflowURL = errors.New("invalid StackOverflow url")
	ErrUnsupportedGithubURL    = errors.New("unsupported github url")
	ErrNotURL                  = errors.New("not url")
)

const (
	linkTrackInterval = time.Second * 10
	gitHubHost        = "github.com"
	stackOverflowHost = "stackoverflow.com"
	linksRequestLimit = 5
)

type ScrapperScheduler struct {
	Scheduler      gocron.Scheduler
	LinksRequester service.LinksRequester
}

func (r ScrapperScheduler) StartScrapperScheduler() {
	_, err := r.Scheduler.NewJob(
		gocron.DurationJob(
			linkTrackInterval,
		),
		gocron.NewTask(
			r.LinksRequester.HandleLinks,
		),
	)
	if err != nil {
		return
	}

	r.Scheduler.Start()
}

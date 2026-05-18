package scheduler

import (
	"time"

	"github.com/go-co-op/gocron/v2"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service"
)

const (
	linkTrackInterval         = time.Second * 10
	notificationCheckInterval = time.Second * 5
)

type ScrapperScheduler struct {
	Scheduler      gocron.Scheduler
	LinksRequester service.LinksRequester
	UpdatesSender  service.UpdatesSender
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
	_, err = r.Scheduler.NewJob(
		gocron.DurationJob(
			notificationCheckInterval,
		),
		gocron.NewTask(
			r.UpdatesSender.SendUpdates,
		),
	)
	if err != nil {
		return
	}

	r.Scheduler.Start()
}

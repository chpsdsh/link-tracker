package scheduler

import (
	"fmt"
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

func (s ScrapperScheduler) StartScrapperScheduler() error {
	_, err := s.Scheduler.NewJob(
		gocron.DurationJob(
			linkTrackInterval,
		),
		gocron.NewTask(
			s.LinksRequester.HandleLinks,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create job for handling links: %w", err)
	}

	_, err = s.Scheduler.NewJob(
		gocron.DurationJob(
			notificationCheckInterval,
		),
		gocron.NewTask(
			s.UpdatesSender.SendUpdates,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create job for sending links: %w", err)
	}

	s.Scheduler.Start()
	return nil
}

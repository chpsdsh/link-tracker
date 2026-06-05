package scrappermetrics

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type LinksCounterScheduler struct {
	Scheduler gocron.Scheduler
	Updater   LinksOnTrackUpdater
}

func NewLinksCounterScheduler(updater LinksOnTrackUpdater) (LinksCounterScheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return LinksCounterScheduler{}, fmt.Errorf("error creating gocron scheduler: %w", err)
	}
	return LinksCounterScheduler{Scheduler: scheduler, Updater: updater}, nil
}

func (s LinksCounterScheduler) Start(interval time.Duration) {
	_, err := s.Scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(s.Updater.UpdateLinksCount),
	)
	if err != nil {
		return
	}
	s.Scheduler.Start()
}

func (s LinksCounterScheduler) Stop() error {
	if err := s.Scheduler.Shutdown(); err != nil {
		return fmt.Errorf("could not stop scheduler: %w", err)
	}
	return nil
}

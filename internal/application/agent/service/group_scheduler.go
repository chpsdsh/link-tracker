package service

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/config"
)

type GroupScheduler struct {
	Scheduler      gocron.Scheduler
	UpdatesGrouper UpdatesGrouper
}

func NewGroupScheduler(updatesGrouper UpdatesGrouper) (GroupScheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return GroupScheduler{}, fmt.Errorf("creating gocron scheduler: %w", err)
	}
	return GroupScheduler{Scheduler: scheduler, UpdatesGrouper: updatesGrouper}, nil
}

func (s GroupScheduler) StartAgentScheduler(cfg config.AIAgentConfig) {
	_, err := s.Scheduler.NewJob(
		gocron.DurationJob(
			cfg.GroupWindow,
		),
		gocron.NewTask(
			s.UpdatesGrouper.Flush,
		),
	)
	if err != nil {
		return
	}

	s.Scheduler.Start()
}

func (s GroupScheduler) ShutdownScheduler() error {
	if err := s.Scheduler.Shutdown(); err != nil {
		return fmt.Errorf("shutting down scheduler: %w", err)
	}
	return nil
}

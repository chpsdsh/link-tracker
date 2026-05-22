package service

import (
	"fmt"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

const stubPriority = "stub"

type Sender interface {
	SendLinkUpdate(update pkg.ProcessedLinkUpdate, eventID string) error
	Close()
}

type UpdatesSummarizer interface {
	Summarize(update pkg.LinkUpdate) (string, error)
}

type AgentService struct {
	Sender     Sender
	Summarizer UpdatesSummarizer
}

func (a AgentService) Summarize(update pkg.LinkUpdate) error {
	summarizedString, err := a.Summarizer.Summarize(update)
	if err != nil {
		return fmt.Errorf("could not summarize link updates: %w", err)
	}
	processedUpdate := pkg.ProcessedLinkUpdate{
		Description: summarizedString,
		ID:          update.ID,
		TgChatIDs:   update.TgChatIDs,
		Priority:    stubPriority,
	}

	slog.Info("sending link update: ", processedUpdate)
	if err = a.Sender.SendLinkUpdate(processedUpdate, summarizedString); err != nil {
		return fmt.Errorf("could not send link update: %w", err)
	}
	return nil
}

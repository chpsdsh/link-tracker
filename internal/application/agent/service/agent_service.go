package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/agent"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

var ErrUnknownPriority = errors.New("unknown priority")

const (
	highPriority   = "HIGH"
	mediumPriority = "MEDIUM"
	lowPriority    = "LOW"
)

type Sender interface {
	SendLinkUpdate(update pkg.ProcessedLinkUpdate, eventID string) error
	Close()
}

type Prioritizer interface {
	FindPriority(description string) agent.Priority
}

type UpdatesSummarizer interface {
	Summarize(update pkg.LinkUpdate) (string, error)
}

type UpdatesGrouper interface {
	Add(update agent.UpdateToGroup)
	Flush()
}

type AgentService struct {
	Sender      Sender
	Summarizer  UpdatesSummarizer
	Prioritizer Prioritizer
	BaseLogger  *slog.Logger
	Grouper     UpdatesGrouper
}

func (a AgentService) ProcessLinkUpdate(update pkg.LinkUpdate) error {
	summarizedString, err := a.Summarizer.Summarize(update)
	if err != nil {
		a.BaseLogger.Error("could not summarize link updates",
			slog.String("error", err.Error()),
			slog.Any("update", update),
		)
		return fmt.Errorf("could not summarize link updates: %w", err)
	}

	priority := a.Prioritizer.FindPriority(summarizedString)

	a.Grouper.Add(agent.UpdateToGroup{Description: summarizedString, Priority: priority, URL: update.URL, TgChatIDs: update.TgChatIDs})

	return nil
}

func (a AgentService) FormAndSendUpdate(update agent.UpdateToSend) error {
	prioritySting, err := priorityToString(update.GroupedUpdates.Priority)
	if err != nil {
		return fmt.Errorf("could not get priority: %w", err)
	}

	description := formatUpdate(update)

	eventID := uuid.New()

	if err = a.Sender.SendLinkUpdate(
		pkg.ProcessedLinkUpdate{
			Description: description,
			TgChatIDs:   []int64{update.TgChatID},
			Priority:    prioritySting,
		}, eventID.String()); err != nil {
		return fmt.Errorf("could not send link update: %w", err)
	}
	return nil
}

func formatUpdate(update agent.UpdateToSend) string {
	var builder strings.Builder

	for i, u := range update.GroupedUpdates.Updates {
		builder.WriteString(strconv.Itoa(i + 1))
		builder.WriteString(") ")
		builder.WriteString(u.Description)
		builder.WriteString("\n")
		builder.WriteString("Ссылка: ")
		builder.WriteString(u.URL)
		builder.WriteString("\n")
	}

	return builder.String()
}

func priorityToString(priority agent.Priority) (string, error) {
	switch priority {
	case agent.PriorityHigh:
		return highPriority, nil
	case agent.PriorityMedium:
		return mediumPriority, nil
	case agent.PriorityLow:
		return lowPriority, nil
	default:
		return "", ErrUnknownPriority
	}
}

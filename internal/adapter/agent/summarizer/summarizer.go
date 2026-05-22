package summarizer

import (
	"errors"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

var ErrNoDestination = errors.New("no destination to summraize")

type LinksSummarizer struct {
	Config config.AIAgentConfig
}

func NewLinksSummarizer(cfg config.AIAgentConfig) LinksSummarizer {
	return LinksSummarizer{Config: cfg}
}

func (s LinksSummarizer) Summarize(update pkg.LinkUpdate) (string, error) { //TODO: переписать на AI API
	if len(update.Description) == 0 {
		return "", ErrNoDestination
	}
	if len(update.Description) > s.Config.Threshold {
		return update.Description[:s.Config.Threshold] + "...", nil
	}
	return update.Description, nil
}

package summarizer

import (
	"errors"
	"fmt"
	"strings"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

var (
	ErrTooShortUpdate       = errors.New("update is too shrot")
	ErrContainsBannedWords  = errors.New("update contains banned words")
	ErrContainsBannedAuthor = errors.New("update contains banned author")
)

type LinksSummarizer struct {
	Config config.AIAgentConfig
}

func NewLinksSummarizer(cfg config.AIAgentConfig) LinksSummarizer {
	return LinksSummarizer{Config: cfg}
}

func (s LinksSummarizer) Summarize(update pkg.LinkUpdate) (string, error) { //TODO: переписать на AI API
	if len(update.Description) < s.Config.MinLength {
		return "", ErrTooShortUpdate
	}

	if err := s.checkBannedWords(update); err != nil {
		return "", fmt.Errorf("banned words: %w", err)
	}

	if err := s.checkBannedAuthors(update); err != nil {
		return "", fmt.Errorf("banned authors: %w", err)
	}

	if len(update.Description) > s.Config.Threshold {
		return update.Description[:s.Config.Threshold] + "...", nil
	}

	return update.Description, nil
}

func (s LinksSummarizer) checkBannedAuthors(update pkg.LinkUpdate) error {
	for _, author := range s.Config.ExcludedAuthors {
		if strings.Contains(update.Description, "Автор: "+author) {
			return fmt.Errorf("%w: %s", ErrContainsBannedWords, author)
		}
	}
	return nil
}

func (s LinksSummarizer) checkBannedWords(update pkg.LinkUpdate) error {
	for _, word := range s.Config.StopWords {
		if strings.Contains(update.Description, word) {
			return fmt.Errorf("%w: %s", ErrContainsBannedWords, word)
		}
	}
	return nil
}

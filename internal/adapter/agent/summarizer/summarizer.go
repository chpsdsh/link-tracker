package summarizer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

var (
	ErrTooShortUpdate       = errors.New("update is too shrot")
	ErrContainsBannedWords  = errors.New("update contains banned words")
	ErrContainsBannedAuthor = errors.New("update contains banned author")
)

const (
	apiRequestTimeout = time.Second * 15
	agentPrompt       = "Summarize the following update in 2–3 sentences"
	gptString         = "gpt://%s/"
)

type LinksSummarizer struct {
	Config config.AIAgentConfig
	Client openai.Client
}

func NewLinksSummarizer(cfg config.AIAgentConfig) LinksSummarizer {
	client := openai.NewClient(
		option.WithAPIKey(cfg.YandexAgentConfig.APIKey),
		option.WithBaseURL(cfg.YandexAgentConfig.BaseURL),
	)
	return LinksSummarizer{Config: cfg, Client: client}
}

func (s LinksSummarizer) Summarize(update pkg.LinkUpdate) (string, error) {
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
		summarizedResult, err := s.makeSummarization(update)
		if err != nil {
			return "", fmt.Errorf("summarizing update: %w", err)
		}
		return summarizedResult, nil
	}

	return update.Description, nil
}

func (s LinksSummarizer) checkBannedAuthors(update pkg.LinkUpdate) error {
	for _, author := range s.Config.ExcludedAuthors {
		if strings.Contains(update.Description, "Автор: "+author) {
			return fmt.Errorf("%w: %s", ErrContainsBannedAuthor, author)
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

func (s LinksSummarizer) makeSummarization(update pkg.LinkUpdate) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiRequestTimeout)
	defer cancel()

	resp, err := s.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: fmt.Sprintf(gptString+s.Config.YandexAgentConfig.Model, s.Config.YandexAgentConfig.FolderID),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(agentPrompt + " " + update.Description + " " + update.URL),
		},
	})
	if err != nil {
		return "", fmt.Errorf("summarizing link update: %w", err)
	}
	return resp.Choices[0].Message.Content, nil
}

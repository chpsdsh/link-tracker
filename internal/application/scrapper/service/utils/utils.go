package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/k3a/html2text"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

var (
	ErrInvalidGitHubURL        = errors.New("invalid github url")
	ErrInvalidStackOverflowURL = errors.New("invalid StackOverflow url")
	ErrUnsupportedGithubURL    = errors.New("unsupported github url")
)

const (
	descriptionMaxLength     = 200
	stackOverflowQuestions   = "questions"
	minGithubURLParts        = 3
	minStackOverflowURLParts = 3
)

func FormatPullRequest(pr scrapper.GithubPullRequest) string {
	body := pr.Body
	if len(body) > descriptionMaxLength {
		body = body[:descriptionMaxLength] + "..."
	}

	return fmt.Sprintf(
		"Pull Request\n\nНазвание: %s\nАвтор: %s\nСоздан: %s\n\n%s",
		pr.Title,
		pr.User.Login,
		pr.CreatedAt.Format(time.RFC3339),
		body,
	)
}

func FormatIssue(issue scrapper.GithubIssue) string {
	body := issue.Body
	if len(body) > descriptionMaxLength {
		body = body[:descriptionMaxLength] + "..."
	}

	return fmt.Sprintf(
		"Issue\n\nНазвание: %s\nАвтор: %s\nСоздан: %s\n\n%s",
		issue.Title,
		issue.User.Login,
		issue.CreatedAt.Format(time.RFC3339),
		body,
	)
}

func FormatStackOverflowAnswer(a scrapper.StackOverflowAnswer) string {
	body := a.Body
	if len(body) > descriptionMaxLength {
		body = body[:descriptionMaxLength] + "..."
	}

	return fmt.Sprintf(
		"Ответ\n\nАвтор: %s\nСоздан: %s\n\n%s",
		a.Owner.DisplayName,
		time.Unix(a.CreationDate, 0).UTC().Format(time.RFC3339),
		html2text.HTML2Text(body),
	)
}

func FormatStackOverflowComment(c scrapper.StackOverflowComment) string {
	body := c.Body
	if len(body) > descriptionMaxLength {
		body = body[:descriptionMaxLength] + "..."
	}

	return fmt.Sprintf(
		"Комментарий\n\nАвтор: %s\nСоздан: %s\n\n%s",
		c.Owner.DisplayName,
		time.Unix(c.CreationDate, 0).UTC().Format(time.RFC3339),
		html2text.HTML2Text(body),
	)
}

func ParseGithubLink(link string) (scrapper.GithubLink, error) {
	parts := strings.Split(strings.Trim(link[8:], "/"), "/")

	if len(parts) < minGithubURLParts {
		slog.Info("invalid link", slog.Any("parts", parts))
		return scrapper.GithubLink{}, ErrInvalidGitHubURL
	}

	owner := parts[1]
	repo := parts[2]

	if len(parts) == minGithubURLParts {
		return scrapper.GithubLink{
			Owner: owner,
			Repo:  repo,
		}, nil
	}

	return scrapper.GithubLink{}, ErrUnsupportedGithubURL
}

func ParseStackOverflowLink(link string) (scrapper.StackOverflowLink, error) {

	parts := strings.Split(strings.Trim(link[8:], "/"), "/")

	if len(parts) < minStackOverflowURLParts {
		return scrapper.StackOverflowLink{}, ErrInvalidStackOverflowURL
	}

	if parts[1] != stackOverflowQuestions {
		return scrapper.StackOverflowLink{}, ErrInvalidStackOverflowURL
	}

	id := parts[2]

	return scrapper.StackOverflowLink{
		ID: id,
	}, nil
}

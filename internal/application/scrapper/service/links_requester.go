//go:generate mockgen -source links_requester.go -destination=../mocks/links_requester_mocks.go -package=mocks
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

var (
	ErrInvalidGitHubURL        = errors.New("invalid github url")
	ErrInvalidStackOverflowURL = errors.New("invalid StackOverflow url")
	ErrUnsupportedGithubURL    = errors.New("unsupported github url")
)

const (
	linksHandleDuration       = time.Second * 10
	repositoryRequestDuration = time.Second * 5
	stackOverflowQuestions    = "questions"
	minGithubURLParts         = 3
	minStackOverflowURLParts  = 2
	descriptionMaxLength      = 200
)

type NetworkClient interface {
	DoGithubRequest(url string) (scrapper.GitHubUpdate, error)
	DoGithubIssueRequest(url string) ([]scrapper.GithubIssue, error)
	DoGithubPullRequestRequest(url string) ([]scrapper.GithubPullRequest, error)
	SendLinkUpdate(update pkg.LinkUpdate) error
	DoStackOverflowQuestionRequest(url string) (scrapper.StackOverflowUpdate, error)
	DoStackOverflowAnswersRequest(url string) (scrapper.StackOverflowAnswersResponse, error)
	DoStackOverflowCommentsRequest(url string) (scrapper.StackOverflowCommentsResponse, error)
}

type LinksRequester struct {
	Client     NetworkClient
	Repo       LinkRepository
	LinksPool  LinksPool
	BaseLogger *slog.Logger
	BatchSize  int
}

func NewLinkRequester(client NetworkClient, repo LinkRepository, numWorkers, batchSize int, logger *slog.Logger) LinksRequester {
	workerPool := NewLinksPool(numWorkers)
	return LinksRequester{Client: client, Repo: repo, LinksPool: workerPool, BatchSize: batchSize, BaseLogger: logger}
}

func (r LinksRequester) HandleLinks() {
	offset := 0
	for r.linksIteration(offset) {
		offset += r.BatchSize
	}
}

func (r LinksRequester) Start(ctx context.Context, wg *sync.WaitGroup) {
	for range r.LinksPool.NumWorkers {
		wg.Go(func() {
			r.worker(ctx)
		})
	}
}

func (r LinksRequester) linksIteration(offset int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), linksHandleDuration)
	defer cancel()

	links, err := r.Repo.GetAllLinks(ctx, r.BatchSize, offset)
	if err != nil {
		r.BaseLogger.Error("error getting github links", slog.String("error", err.Error()))
		return false
	}

	if len(links) == 0 {
		return false
	}

	startOffset := 0
	var batchSize int
	if len(links) > r.LinksPool.NumWorkers {
		batchSize = len(links) / r.LinksPool.NumWorkers
	} else {
		batchSize = len(links)
	}
	endOffset := batchSize
	for i := range r.LinksPool.NumWorkers {
		if endOffset > len(links) {
			endOffset = len(links)
		}
		slog.Info("processing link", slog.Int("startoffset", startOffset+i), slog.Int("endoffset", endOffset))
		r.LinksPool.LinksChan <- links[startOffset:endOffset]

		startOffset = endOffset + 1
		if startOffset > len(links) {
			break
		}

		if i+1 == r.LinksPool.NumWorkers {
			endOffset = len(links)
		} else {
			endOffset += batchSize + 1
		}
	}
	return true
}

func (r LinksRequester) handleGithubLink(link pkg.LinkInfo) {
	gitLink, err := parseGithubLink(link.Link)
	if err != nil {
		r.BaseLogger.Error("error parsing github link", slog.String("error", err.Error()))
		r.sendUpdate(link, "Error parsing github link")
		return
	}

	timeToUpdate := link.LastUpdateTime

	issueLastUpdateTime := r.handleIssueUpdates(gitLink, link)
	if issueLastUpdateTime.After(timeToUpdate) {
		timeToUpdate = issueLastUpdateTime
	}

	pullRequestUpdateTime := r.handlePullRequestsUpdates(gitLink, link)
	if pullRequestUpdateTime.After(timeToUpdate) {
		timeToUpdate = pullRequestUpdateTime
	}

	repoUpdateTime := r.handleRepositoryUpdates(gitLink, link)
	if repoUpdateTime.After(timeToUpdate) {
		timeToUpdate = repoUpdateTime
	}

	ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestDuration)
	defer cancel()

	if timeToUpdate.After(link.LastUpdateTime) {
		if err = r.Repo.UpdateLinksTime(ctx, timeToUpdate, link.Link); err != nil {
			r.BaseLogger.Error("error updating github link", slog.String("error", err.Error()))
		}
	}
}

func (r LinksRequester) handleIssueUpdates(gitLink scrapper.GithubLink, link pkg.LinkInfo) time.Time {
	issueUpdate, err := r.Client.DoGithubIssueRequest(gitLink.ConvertToURL(scrapper.GithubLinkOptionIssue))
	if err != nil {
		r.BaseLogger.Error("error requesting issues:", slog.String("error", err.Error()))
		return link.LastUpdateTime
	}
	newUpdateTime := link.LastUpdateTime

	for _, item := range issueUpdate {
		if item.UpdatedAt.After(link.LastUpdateTime) {
			r.sendUpdate(link, formatIssue(item))
			if item.UpdatedAt.After(newUpdateTime) {
				newUpdateTime = item.UpdatedAt
			}
		}
	}

	return newUpdateTime
}

func (r LinksRequester) handlePullRequestsUpdates(gitLink scrapper.GithubLink, link pkg.LinkInfo) time.Time {
	pullRequestUpdate, err := r.Client.DoGithubPullRequestRequest(gitLink.ConvertToURL(scrapper.GithubLinkPullRequest))
	if err != nil {
		r.BaseLogger.Error("error requesting issues:", slog.String("error", err.Error()))
		return link.LastUpdateTime
	}

	newUpdateTime := link.LastUpdateTime

	for _, item := range pullRequestUpdate {
		slog.Info("updated_at", slog.Any("updated_at", item.UpdatedAt))
		if item.UpdatedAt.After(link.LastUpdateTime) {
			r.sendUpdate(link, formatPullRequest(item))
			if item.UpdatedAt.After(newUpdateTime) {
				newUpdateTime = item.UpdatedAt
			}
		}
	}

	return newUpdateTime
}

func (r LinksRequester) handleRepositoryUpdates(gitLink scrapper.GithubLink, link pkg.LinkInfo) time.Time {
	gitUpdate, err := r.Client.DoGithubRequest(gitLink.ConvertToURL(scrapper.GithubLinkOptionRepository))
	if err != nil {
		r.BaseLogger.Error("error during github query", slog.String("error", err.Error()))
		return link.LastUpdateTime
	}
	if gitUpdate.UpdatedAt.After(link.LastUpdateTime) {
		r.sendUpdate(link, "Repository updated:"+link.Link)
	}
	return gitUpdate.UpdatedAt
}

func formatPullRequest(pr scrapper.GithubPullRequest) string {
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

func formatIssue(issue scrapper.GithubIssue) string {
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

func (r LinksRequester) worker(ctx context.Context) {
	for {
		select {
		case links := <-r.LinksPool.LinksChan:
			for _, link := range links {
				switch scrapper.GetLinkType(link.Link) {
				case scrapper.GithubLinkType:
					r.handleGithubLink(link)
				case scrapper.StackOverflowLinkType:
					r.handleStackOverflowLink(link)
				case scrapper.UnknownLinkType:
					r.sendUpdate(link, "only github and stackOverflow links are supported")
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func (r LinksRequester) handleStackOverflowLink(link pkg.LinkInfo) {
	stackOverflowLink, err := parseStackOverflowLink(link.Link)
	if err != nil {
		r.sendUpdate(link, "Error parsing stackOverflow link")
		return
	}

	timeToUpdate := link.LastUpdateTime

	answerTimeToUpdate := r.handleStackOverflowAnswers(stackOverflowLink, link)
	if answerTimeToUpdate.After(timeToUpdate) {
		timeToUpdate = answerTimeToUpdate
	}

	commentsTimeToUpdate := r.handleStackOverflowComments(stackOverflowLink, link)
	if commentsTimeToUpdate.After(timeToUpdate) {
		timeToUpdate = commentsTimeToUpdate
	}

	questionTimeToUpdate := r.handleStackOverflowQuestion(stackOverflowLink, link)
	if questionTimeToUpdate.After(timeToUpdate) {
		timeToUpdate = questionTimeToUpdate
	}

	ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestDuration)
	defer cancel()

	if timeToUpdate.After(link.LastUpdateTime) {
		if err = r.Repo.UpdateLinksTime(ctx, timeToUpdate, link.Link); err != nil {
			r.BaseLogger.Error("error updating github link", slog.String("error", err.Error()))
		}
	}
}

func (r LinksRequester) handleStackOverflowAnswers(stackOverflowLink scrapper.StackOverflowLink, link pkg.LinkInfo) time.Time {
	stackUpdate, err := r.Client.DoStackOverflowAnswersRequest(stackOverflowLink.ConvertToURL(scrapper.StackOverflowLinkAnswer))
	timeToUpdate := link.LastUpdateTime
	if err != nil {
		r.BaseLogger.Error("error querying stack overflow", slog.String("error", err.Error()))
		return timeToUpdate
	}

	for _, item := range stackUpdate.Items {
		updateTime := time.Unix(item.LastActivityDate, 0).UTC()
		if updateTime.After(timeToUpdate) {
			timeToUpdate = updateTime
		}
	}
	return timeToUpdate
}

func (r LinksRequester) handleStackOverflowComments(stackOverflowLink scrapper.StackOverflowLink, link pkg.LinkInfo) time.Time {
	stackUpdate, err := r.Client.DoStackOverflowCommentsRequest(stackOverflowLink.ConvertToURL(scrapper.StackOverflowLinkComment))
	timeToUpdate := link.LastUpdateTime
	if err != nil {
		r.BaseLogger.Error("error querying stack overflow", slog.String("error", err.Error()))
		return timeToUpdate
	}

	for _, item := range stackUpdate.Items {
		updateTime := time.Unix(item.LastActivityDate, 0).UTC()
		if updateTime.After(timeToUpdate) {
			timeToUpdate = updateTime
		}
	}
	return timeToUpdate
}

func (r LinksRequester) handleStackOverflowQuestion(stackOverflowLink scrapper.StackOverflowLink, link pkg.LinkInfo) time.Time {
	stackUpdate, err := r.Client.DoStackOverflowQuestionRequest(stackOverflowLink.ConvertToURL(scrapper.StackOverflowLinkQuestion))
	timeToUpdate := link.LastUpdateTime
	if err != nil {
		r.BaseLogger.Error("error querying stack overflow", slog.String("error", err.Error()))
		return timeToUpdate
	}

	for _, item := range stackUpdate.Items {
		updateTime := time.Unix(item.LastActivityDate, 0).UTC()
		if updateTime.After(timeToUpdate) {
			timeToUpdate = updateTime
		}
	}
	return timeToUpdate
}

func (r LinksRequester) sendUpdate(linkInfo pkg.LinkInfo, description string) {
	ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestDuration)
	defer cancel()

	chatIDs, err := r.Repo.GetChatIDsByLink(ctx, linkInfo.Link)
	if err != nil {
		r.BaseLogger.Error("error getting chatIDs", slog.String("error", err.Error()))
		return
	}

	update := pkg.LinkUpdate{Description: description, TgChatIDs: chatIDs, URL: linkInfo.Link}

	if err = r.Client.SendLinkUpdate(update); err != nil {
		r.BaseLogger.Error("error sending link update", slog.String("error", err.Error()))
		return
	}
	r.BaseLogger.Info("link is sent to chats", slog.String("link", linkInfo.Link), slog.Any("chats", chatIDs))

}

func parseGithubLink(link string) (scrapper.GithubLink, error) {
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

func parseStackOverflowLink(link string) (scrapper.StackOverflowLink, error) {

	parts := strings.Split(strings.Trim(link[8:], "/"), "/")

	if len(parts) < minStackOverflowURLParts {
		return scrapper.StackOverflowLink{}, ErrInvalidStackOverflowURL
	}

	if parts[1] != stackOverflowQuestions {
		return scrapper.StackOverflowLink{}, ErrInvalidStackOverflowURL
	}

	id := parts[1]

	return scrapper.StackOverflowLink{
		ID: id,
	}, nil
}

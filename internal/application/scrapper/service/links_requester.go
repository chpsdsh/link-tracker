//go:generate mockgen -source links_requester.go -destination=../mocks/links_requester_mocks.go -package=mocks
package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	utils "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/service/requesterutils"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

const (
	linksHandleDuration       = time.Second * 10
	repositoryRequestDuration = time.Second * 5
)

type NetworkClient interface {
	DoGithubRequest(url string) (scrapper.GitHubRepositoryResponse, error)
	DoGithubIssueRequest(url string) ([]scrapper.GithubIssue, error)
	DoGithubPullRequestRequest(url string) ([]scrapper.GithubPullRequest, error)
	DoStackOverflowQuestionRequest(url string) (scrapper.StackOverflowQuestionResponse, error)
	DoStackOverflowAnswersRequest(url string) (scrapper.StackOverflowAnswersResponse, error)
	DoStackOverflowCommentsRequest(url string) (scrapper.StackOverflowCommentsResponse, error)
}

type OutboxRepository interface {
	SaveUpdate(ctx context.Context, update pkg.LinkUpdate) error
	GetUpdates(ctx context.Context) ([]scrapper.OutboxEvent, error)
	UpdateSendTime(ctx context.Context, id int64) error
}

type LinksRequester struct {
	Client             NetworkClient
	NotificationSender Sender
	Repo               LinkRepository
	LinksPool          LinksPool
	BaseLogger         *slog.Logger
	OutboxRepo         OutboxRepository
	Transactor         Transactor
	BatchSize          int
}

func NewLinkRequester(
	client NetworkClient,
	sender Sender,
	repo LinkRepository,
	outboxRepo OutboxRepository,
	transactor Transactor,
	numWorkers, batchSize int,
	logger *slog.Logger) LinksRequester {
	workerPool := NewLinksPool(numWorkers)
	return LinksRequester{Client: client,
		NotificationSender: sender,
		Repo:               repo,
		LinksPool:          workerPool,
		BatchSize:          batchSize,
		Transactor:         transactor,
		OutboxRepo:         outboxRepo,
		BaseLogger:         logger}
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
					r.sendFailureUpdate(link, "Only github and stackOverflow links are supported")
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func (r LinksRequester) handleGithubLink(link pkg.LinkInfo) {
	gitLink, err := utils.ParseGithubLink(link.Link)
	if err != nil {
		r.BaseLogger.Error("error parsing github link", slog.String("error", err.Error()))
		r.sendFailureUpdate(link, "Error parsing github link")
		return
	}

	var processingResult scrapper.LinkProcessingResult

	processingResult = r.handleIssueUpdates(gitLink, link, processingResult)

	processingResult = r.handlePullRequestsUpdates(gitLink, link, processingResult)

	processingResult = r.handleRepositoryUpdates(gitLink, link, processingResult)

	ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestDuration)
	defer cancel()

	if processingResult.UpdateTime.After(link.LastUpdateTime) {
		if err = r.Transactor.Transaction(ctx, func(ctx context.Context) error {
			if err = r.Repo.UpdateLinksTime(ctx, processingResult.UpdateTime, link.Link); err != nil {
				r.BaseLogger.Error("error updating github link", slog.String("error", err.Error()))
				return fmt.Errorf("error updating github link: %w", err)
			}
			for _, event := range processingResult.Events {
				if err = r.OutboxRepo.SaveUpdate(ctx, event); err != nil {
					r.BaseLogger.Error("error saving github link", slog.String("error", err.Error()))
					return fmt.Errorf("error saving github link to outbox table: %w", err)
				}
			}
			return nil
		}); err != nil {
			r.BaseLogger.Error("error saving link updates to outbox table", slog.String("error", err.Error()))
		}
	}
}

func (r LinksRequester) handleIssueUpdates(gitLink scrapper.GithubLink, link pkg.LinkInfo, processingResult scrapper.LinkProcessingResult) scrapper.LinkProcessingResult {
	issueUpdate, err := r.Client.DoGithubIssueRequest(gitLink.ConvertToURL(scrapper.GithubLinkOptionIssue))
	if err != nil {
		r.BaseLogger.Error("error requesting issues:", slog.String("error", err.Error()))
		return processingResult
	}

	for _, item := range issueUpdate {
		if item.UpdatedAt.After(link.LastUpdateTime) {
			if item.UpdatedAt.After(processingResult.UpdateTime) {
				processingResult.UpdateTime = item.UpdatedAt
			}
			ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestDuration)

			chatIDs, chatIDsErr := r.Repo.GetChatIDsByLink(ctx, link.Link)
			if chatIDsErr != nil {
				r.BaseLogger.Error("error getting chatIDs", slog.String("error", chatIDsErr.Error()))
				cancel()
				return processingResult
			}
			processingResult.Events = append(processingResult.Events, pkg.LinkUpdate{Description: utils.FormatIssue(item), URL: link.Link, TgChatIDs: chatIDs})
			cancel()
		}
	}

	return processingResult
}

func (r LinksRequester) handlePullRequestsUpdates(gitLink scrapper.GithubLink, link pkg.LinkInfo, processingResult scrapper.LinkProcessingResult) scrapper.LinkProcessingResult {
	pullRequestUpdate, err := r.Client.DoGithubPullRequestRequest(gitLink.ConvertToURL(scrapper.GithubLinkPullRequest))
	if err != nil {
		r.BaseLogger.Error("error requesting issues:", slog.String("error", err.Error()))
		return processingResult
	}

	for _, item := range pullRequestUpdate {
		if item.UpdatedAt.After(link.LastUpdateTime) {
			if item.UpdatedAt.After(processingResult.UpdateTime) {
				processingResult.UpdateTime = item.UpdatedAt
			}
			ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestDuration)

			chatIDs, chatIDsErr := r.Repo.GetChatIDsByLink(ctx, link.Link)
			if chatIDsErr != nil {
				r.BaseLogger.Error("error getting chatIDs", slog.String("error", chatIDsErr.Error()))
				cancel()
				return processingResult
			}
			processingResult.Events = append(processingResult.Events, pkg.LinkUpdate{Description: utils.FormatPullRequest(item), URL: link.Link, TgChatIDs: chatIDs})
			cancel()
		}
	}

	return processingResult
}

func (r LinksRequester) handleRepositoryUpdates(gitLink scrapper.GithubLink, link pkg.LinkInfo, processingResult scrapper.LinkProcessingResult) scrapper.LinkProcessingResult {
	repoUpdate, err := r.Client.DoGithubRequest(gitLink.ConvertToURL(scrapper.GithubLinkOptionRepository))
	if err != nil {
		r.BaseLogger.Error("error during github query", slog.String("error", err.Error()))
		return processingResult
	}

	if repoUpdate.UpdatedAt.After(link.LastUpdateTime) {
		if repoUpdate.UpdatedAt.After(processingResult.UpdateTime) {
			processingResult.UpdateTime = repoUpdate.UpdatedAt
		}

		ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestDuration)
		defer cancel()

		chatIDs, chatIDsErr := r.Repo.GetChatIDsByLink(ctx, link.Link)
		if chatIDsErr != nil {
			r.BaseLogger.Error("error getting chatIDs", slog.String("error", chatIDsErr.Error()))
			return processingResult
		}

		processingResult.Events = append(processingResult.Events, pkg.LinkUpdate{Description: "Repository updated:", URL: link.Link, TgChatIDs: chatIDs})
	}
	return processingResult
}

func (r LinksRequester) handleStackOverflowLink(link pkg.LinkInfo) {

	stackOverflowLink, err := utils.ParseStackOverflowLink(link.Link)
	if err != nil {
		r.sendFailureUpdate(link, "Error parsing stackOverflow link")
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
			r.sendFailureUpdate(link, utils.FormatStackOverflowAnswer(item))
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
		updateTime := time.Unix(item.CreationDate, 0).UTC()
		if updateTime.After(timeToUpdate) {
			r.sendFailureUpdate(link, utils.FormatStackOverflowComment(item))
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
			r.sendFailureUpdate(link, "Question updated:")
			timeToUpdate = updateTime
		}
	}
	return timeToUpdate
}

func (r LinksRequester) sendFailureUpdate(linkInfo pkg.LinkInfo, description string) {
	ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestDuration)
	defer cancel()

	chatIDs, err := r.Repo.GetChatIDsByLink(ctx, linkInfo.Link)
	if err != nil {
		r.BaseLogger.Error("error getting chatIDs", slog.String("error", err.Error()))
		return
	}

	update := pkg.LinkUpdate{Description: description, TgChatIDs: chatIDs, URL: linkInfo.Link}

	if err = r.OutboxRepo.SaveUpdate(ctx, update); err != nil {
		r.BaseLogger.Error("error sending link update", slog.String("error", err.Error()))
		return
	}
	r.BaseLogger.Info("link is sent to chats", slog.String("link", linkInfo.Link), slog.Any("chats", chatIDs))
}

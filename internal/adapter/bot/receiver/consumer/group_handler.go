package consumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/botmetrics"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/dlqproducer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/repository"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

var (
	ErrValidatingLinkUpdate = errors.New("error validating LinkUpdate")
	ErrRetriesLimitExceeded = errors.New("all retries failed")
	ErrNoEventIDInHeader    = errors.New("no event id found in header")
)

const (
	maxRetries               = 5
	kafkaHeaderKey           = "event_id"
	repositoryRequestTimeout = 5 * time.Second
)

type InboxRepository interface {
	Save(ctx context.Context, eventID, consumerName string) error
	UpdateProcessedTime(ctx context.Context, eventID string) error
}

type GroupHandler struct {
	logger            *slog.Logger
	dlqProducer       dlqproducer.DlqProducer
	TgHandler         handler.TelegramBotHandler
	InboxRepo         InboxRepository
	consumerGroupName string
}

func NewGroupHandler(dlqProducer dlqproducer.DlqProducer, tgHandler handler.TelegramBotHandler, inboxRepo InboxRepository, logger *slog.Logger, consumerGroupName string) GroupHandler {
	return GroupHandler{dlqProducer: dlqProducer, TgHandler: tgHandler, logger: logger, InboxRepo: inboxRepo, consumerGroupName: consumerGroupName}
}

func (GroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (GroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h GroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		botmetrics.CommandRequestTotal.WithLabelValues("consumer-message").Inc()
		startTime := time.Now()
		linkUpdate := pkg.ProcessedLinkUpdate{}
		if err := json.Unmarshal(msg.Value, &linkUpdate); err != nil {
			h.logger.Error("unmarshall:", slog.String("err", err.Error()))
			if dlqErr := h.dlqProducer.SendToDLQ(msg, err); dlqErr != nil {
				h.logger.Error("dlq send failed", slog.String("err", dlqErr.Error()))
				continue
			}
			sess.MarkMessage(msg, "")
			sess.Commit()
			botmetrics.ObserveCommandDuration(startTime, kafkaScope, "consumer-message")
			continue
		}

		eventID, err := h.deduplicateMessage(sess, msg)
		if err != nil {
			if errors.Is(err, repository.ErrNotificationAlreadySent) {
				sess.MarkMessage(msg, "")
				sess.Commit()
				botmetrics.ObserveCommandDuration(startTime, kafkaScope, "consumer-message")
				continue
			}
			h.logger.Error("deduplicate message:", slog.String("err", err.Error()))
		}

		if linkUpdate.Description == "" || linkUpdate.Priority == "" || len(linkUpdate.TgChatIDs) == 0 {
			h.logger.Error("validation:", slog.String("err", ErrValidatingLinkUpdate.Error()))
			if dlqErr := h.dlqProducer.SendToDLQ(msg, ErrValidatingLinkUpdate); dlqErr != nil {
				h.logger.Error("dlq send failed", slog.String("err", dlqErr.Error()))
				continue
			}
			sess.MarkMessage(msg, "")
			sess.Commit()
			botmetrics.ObserveCommandDuration(startTime, kafkaScope, "consumer-message")
			continue
		}

		if err = h.processWithRetry(linkUpdate); err != nil {
			h.logger.Error("consuming message failed:", slog.String("err", err.Error()))
			if dlqErr := h.dlqProducer.SendToDLQ(msg, err); dlqErr != nil {
				h.logger.Error("dlq send failed", slog.String("err", dlqErr.Error()))
				continue
			}
			sess.MarkMessage(msg, "")
			sess.Commit()
			botmetrics.ObserveCommandDuration(startTime, kafkaScope, "consumer-message")
			continue
		}
		sess.MarkMessage(msg, "")
		sess.Commit()
		botmetrics.ObserveCommandDuration(startTime, kafkaScope, "consumer-message")

		ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestTimeout)

		if err = h.InboxRepo.UpdateProcessedTime(ctx, eventID); err != nil {
			h.logger.Error("update processed time in inbox table failed:", slog.String("err", err.Error()))
		}
		cancel()
	}
	return nil
}

func (h GroupHandler) deduplicateMessage(sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) (string, error) {
	eventID, ok := getHeader(msg, kafkaHeaderKey)
	if !ok || eventID == "" {
		if err := h.dlqProducer.SendToDLQ(msg, ErrNoEventIDInHeader); err != nil {
			h.logger.Error("dlq send failed", slog.String("err", err.Error()))
			return eventID, fmt.Errorf("error sending to dlq %w", err)
		}
		sess.MarkMessage(msg, "")
		sess.Commit()
		return eventID, ErrNoEventIDInHeader
	}
	ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestTimeout)
	defer cancel()
	if err := h.InboxRepo.Save(ctx, eventID, h.consumerGroupName); err != nil {
		if errors.Is(err, repository.ErrNotificationAlreadySent) {
			sess.MarkMessage(msg, "")
			sess.Commit()
			return eventID, fmt.Errorf("notification already sent %w", err)
		}

		if err = h.dlqProducer.SendToDLQ(msg, ErrNoEventIDInHeader); err != nil {
			h.logger.Error("dlq send failed", slog.String("err", err.Error()))
			return eventID, fmt.Errorf("error sending to dlq %w", err)
		}
		sess.MarkMessage(msg, "")
		sess.Commit()
	}
	return eventID, nil
}

func (h GroupHandler) processWithRetry(linkUpdate pkg.ProcessedLinkUpdate) error {
	var err error
	for range maxRetries {
		err = h.TgHandler.HandleLinkUpdate(linkUpdate)
		if err == nil {
			return nil
		}
	}
	return errors.Join(err, ErrRetriesLimitExceeded)
}

func getHeader(msg *sarama.ConsumerMessage, key string) (string, bool) {
	for _, h := range msg.Headers {
		if string(h.Key) == key {
			return string(h.Value), true
		}
	}
	return "", false

}

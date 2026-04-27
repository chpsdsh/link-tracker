package consumer

import (
	"errors"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/bot/producer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/handler"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

var (
	ErrValidatingLinkUpdate = errors.New("error validating LinkUpdate")
	ErrRetriesLimitExceeded = errors.New("all retries failed")
)

const (
	maxRetries = 5
)

type GroupHandler struct {
	logger      *slog.Logger
	dlqProducer producer.DlqProducer
	TgHandler   handler.TelegramBotHandler
}

func NewGroupHandler(dlqProducer producer.DlqProducer, tgHandler handler.TelegramBotHandler, logger *slog.Logger) GroupHandler {
	return GroupHandler{dlqProducer: dlqProducer, TgHandler: tgHandler, logger: logger}
}

func (GroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (GroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h GroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		linkUpdate := pkg.LinkUpdate{}
		if err := json.Unmarshal(msg.Value, &linkUpdate); err != nil {
			h.logger.Error("unmarshall:", slog.String("err", err.Error()))
			if dlqErr := h.dlqProducer.SendToDLQ(msg, err); dlqErr != nil {
				h.logger.Error("dlq send failed", slog.String("err", dlqErr.Error()))
				continue
			}
			sess.MarkMessage(msg, "")
			sess.Commit()
			continue
		}

		if linkUpdate.Description == "" || linkUpdate.URL == "" || len(linkUpdate.TgChatIDs) == 0 {
			h.logger.Error("validation:", slog.String("err", ErrValidatingLinkUpdate.Error()))
			if dlqErr := h.dlqProducer.SendToDLQ(msg, ErrValidatingLinkUpdate); dlqErr != nil {
				h.logger.Error("dlq send failed", slog.String("err", dlqErr.Error()))
				continue
			}
			sess.MarkMessage(msg, "")
			sess.Commit()
			continue
		}

		if err := h.processWithRetry(linkUpdate); err != nil {
			h.logger.Error("consuming message failed:", slog.String("err", err.Error()))
			if dlqErr := h.dlqProducer.SendToDLQ(msg, err); dlqErr != nil {
				h.logger.Error("dlq send failed", slog.String("err", dlqErr.Error()))
				continue
			}
			sess.MarkMessage(msg, "")
			sess.Commit()
			continue
		}
		sess.MarkMessage(msg, "")
		sess.Commit()
	}
	return nil
}

func (h GroupHandler) processWithRetry(linkUpdate pkg.LinkUpdate) error {
	var err error
	for range maxRetries {
		err = h.TgHandler.HandleLinkUpdate(linkUpdate)
		if err == nil {
			return nil
		}
	}
	return errors.Join(err, ErrRetriesLimitExceeded)
}

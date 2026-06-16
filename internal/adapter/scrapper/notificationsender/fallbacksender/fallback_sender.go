//go:generate mockgen -source fallback_sender.go -destination=../mocks/fallback_sender.go -package=mocks

package fallbacksender

import (
	"fmt"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

type UpdateSender interface {
	SendLinkUpdate(update pkg.LinkUpdate, key string) error
	Close()
}

type FallbackSender struct {
	kafkaSender UpdateSender
	httpSender  UpdateSender
}

func NewFallbackSender(kafkaSender UpdateSender, httpSender UpdateSender) *FallbackSender {
	return &FallbackSender{kafkaSender: kafkaSender, httpSender: httpSender}
}

func (c FallbackSender) SendLinkUpdate(update pkg.LinkUpdate, eventID string) error {
	if err := c.httpSender.SendLinkUpdate(update, ""); err != nil {
		if kafkaErr := c.kafkaSender.SendLinkUpdate(update, eventID); kafkaErr != nil {
			return fmt.Errorf("error sending update using http err: %w kafka fallback err: %w ", kafkaErr, err)
		}
	}
	return nil
}

func (c FallbackSender) Close() {
	c.kafkaSender.Close()
	c.httpSender.Close()
}

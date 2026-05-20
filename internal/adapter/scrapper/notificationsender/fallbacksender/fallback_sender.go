package fallbacksender

import (
	"fmt"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/notificationsender/httpsender"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/notificationsender/producer"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

type FallbackSender struct {
	kafkaSender *producer.KafkaProducer
	httpSender  *httpsender.UpdatesClient
}

func NewFallbackSender(kafkaSender *producer.KafkaProducer, httpSender *httpsender.UpdatesClient) *FallbackSender {
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

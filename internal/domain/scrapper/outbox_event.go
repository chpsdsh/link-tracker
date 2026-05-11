package scrapper

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"

type OutboxEvent struct {
	ID      int64
	EventID string
	Payload pkg.LinkUpdate
}

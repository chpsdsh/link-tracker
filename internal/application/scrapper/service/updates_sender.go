package service

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

type Sender interface {
	SendLinkUpdate(update pkg.LinkUpdate, eventID string) error
	Close()
}

type UpdatesSender struct {
	OutboxRepo         OutboxRepository
	Transactor         Transactor
	NotificationSender Sender
	BaseLogger         *slog.Logger
}

func (u UpdatesSender) SendUpdates() {
	ctx, cancel := context.WithTimeout(context.Background(), repositoryRequestDuration)
	defer cancel()

	if err := u.Transactor.Transaction(ctx, func(ctx context.Context) error {
		linksUpdates, err := u.OutboxRepo.GetUpdates(ctx)
		if err != nil {
			return fmt.Errorf("error getting updates form outbox: %w", err)
		}

		for _, update := range linksUpdates {
			if err = u.NotificationSender.SendLinkUpdate(update.Payload, update.EventID); err != nil {
				return fmt.Errorf("error sending link update: %w", err)
			}

			if err = u.OutboxRepo.UpdateSendTime(ctx, update.ID); err != nil {
				return fmt.Errorf("error updating outbox table: %w", err)
			}
		}

		return nil
	}); err != nil {
		u.BaseLogger.Error("error sending link update", slog.String("error", err.Error()))
	}
}

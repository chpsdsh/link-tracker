package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/database"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/metrics"
)

const databaseScope = "database"

var (
	ErrNotificationAlreadySent = errors.New("notification already sent")
)

type InboxRepository struct {
	db *pgxpool.Pool
}

func NewInboxRepository(db *pgxpool.Pool) *InboxRepository {
	return &InboxRepository{db: db}
}

func (r *InboxRepository) Save(ctx context.Context, eventID, consumerName string) error {
	q := database.GetQuerier(ctx, r.db)
	startTime := time.Now()
	_, err := q.Exec(ctx, `
		INSERT INTO inbox(event_id, consumer_name) 
		VALUES ($1, $2)
		ON CONFLICT (event_id, consumer_name) DO NOTHING`, eventID, consumerName)
	metrics.ObserveRequestDuration(startTime, databaseScope, "inbox")

	if err != nil {
		if database.IsUniqueViolation(err) {
			return ErrNotificationAlreadySent
		}
	}
	return nil
}

func (r *InboxRepository) UpdateProcessedTime(ctx context.Context, eventID string) error {
	q := database.GetQuerier(ctx, r.db)
	startTime := time.Now()
	_, err := q.Exec(ctx, `
	UPDATE inbox SET processed_at = now()
	WHERE event_id = $1`, eventID)
	metrics.ObserveRequestDuration(startTime, databaseScope, "inbox")

	if err != nil {
		return fmt.Errorf("update processed time: %w", err)
	}
	return nil
}

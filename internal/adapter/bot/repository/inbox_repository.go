package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/database"
)

var (
	ErrNotificationAlreadySent = errors.New("notification already sent")
)

type InboxRepository struct {
	db *pgxpool.Pool
}

func NewInboxRepository(db *pgxpool.Pool) *InboxRepository {
	return &InboxRepository{db: db}
}

func (r *InboxRepository) Save(ctx context.Context, eventID string) error {
	q := database.GetQuerier(ctx, r.db)
	tag, err := q.Exec(ctx, `
		INSERT INTO inbox(event_id) 
		VALUES ($1)
		ON CONFLICT (event_id) DO NOTHING`, eventID)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return ErrNotificationAlreadySent
		}
	}
	tag.RowsAffected()
	return nil
}

func (r *InboxRepository) UpdateProcessedTime(ctx context.Context, eventID string) error {
	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, `
	UPDATE inbox SET processed_at = now()
	WHERE event_id = $1`, eventID)
	if err != nil {
		return fmt.Errorf("update processed time: %w", err)
	}
	return nil
}

package sqlrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/database"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/metrics"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

const databaseScope = "database"

type ChatRepository struct {
	db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) ChatExists(ctx context.Context, chatID int64) (bool, error) {
	q := database.GetQuerier(ctx, r.db)

	var exists bool
	startTime := time.Now()
	err := q.QueryRow(ctx, `select exists(
	select 1 from chats where chat_id = $1
    )
	`, chatID).Scan(&exists)
	metrics.ObserveRequestDuration(startTime, databaseScope, "chats")

	if err != nil {
		return false, fmt.Errorf("error check chat existence: %w", err)
	}
	return exists, nil
}

func (r *ChatRepository) AddChat(ctx context.Context, chatID int64) error {
	q := database.GetQuerier(ctx, r.db)

	startTime := time.Now()
	_, err := q.Exec(ctx, `insert into chats (chat_id) values ($1)`, chatID)
	metrics.ObserveRequestDuration(startTime, databaseScope, "chats")

	switch {
	case database.IsUniqueViolation(err):
		return scrapper.ErrChatAlreadyExists
	case err != nil:
		return fmt.Errorf("error adding chat: %w", err)
	}

	return nil
}

func (r *ChatRepository) DeleteChat(ctx context.Context, chatID int64) error {
	q := database.GetQuerier(ctx, r.db)

	startTime := time.Now()
	commandTag, err := q.Exec(ctx, `delete from chats where chat_id = $1`, chatID)
	metrics.ObserveRequestDuration(startTime, databaseScope, "chats")
	if err != nil {
		return fmt.Errorf("error deleting chat: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return scrapper.ErrChatNotFound
	}

	return nil
}

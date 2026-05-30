package builderrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/database"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/metrics"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

const databaseScope = "database"

type ChatRepository struct {
	db      *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	b := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &ChatRepository{db: db, builder: b}
}

func (r *ChatRepository) ChatExists(ctx context.Context, chatID int64) (bool, error) {
	q := database.GetQuerier(ctx, r.db)

	var exists bool

	subquery := r.builder.Select("1").
		From("chats").
		Where(squirrel.Eq{"chat_id": chatID})

	query, args, err := r.builder.Select().
		Column(squirrel.Expr("exists (?)", subquery)).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("error building query %w", err)
	}

	startTime := time.Now()
	err = q.QueryRow(ctx, query, args...).Scan(&exists)
	metrics.ObserveRequestDuration(startTime, databaseScope, "chats")

	if err != nil {
		return false, fmt.Errorf("error checking existence of chat: %w", err)
	}
	return exists, nil
}

func (r *ChatRepository) AddChat(ctx context.Context, chatID int64) error {
	q := database.GetQuerier(ctx, r.db)

	query, args, err := r.builder.Insert("chats").
		Values(chatID).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	startTime := time.Now()
	_, err = q.Exec(ctx, query, args...)
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

	query, args, err := r.builder.Delete("chats").
		Where(squirrel.Eq{"chat_id": chatID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	startTime := time.Now()
	commandTag, err := q.Exec(ctx, query, args...)
	metrics.ObserveRequestDuration(startTime, databaseScope, "chats")

	if err != nil {
		return fmt.Errorf("error deleting chat: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return scrapper.ErrChatNotFound
	}

	return nil
}

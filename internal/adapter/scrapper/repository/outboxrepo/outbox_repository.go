package outboxrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/pkg/database"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/metrics"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/scrapper"
)

const databaseScope = "database"

type OutboxRepository struct {
	db *pgxpool.Pool
}

func NewOutboxRepository(db *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{
		db: db,
	}
}

func (r *OutboxRepository) SaveUpdate(ctx context.Context, update pkg.LinkUpdate) error {
	q := database.GetQuerier(ctx, r.db)
	updateJSON, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("could not marshal update: %w", err)
	}

	startTime := time.Now()
	_, err = q.Exec(ctx, `
	insert into outbox(event_type, payload) values ($1, $2::jsonb)`,
		update.Description,
		updateJSON,
	)
	metrics.ObserveRequestDuration(startTime, databaseScope, "outbox")

	if err != nil {
		return fmt.Errorf("could not insert update: %w", err)
	}
	return nil
}

func (r *OutboxRepository) GetUpdates(ctx context.Context) ([]scrapper.OutboxEvent, error) {
	q := database.GetQuerier(ctx, r.db)

	startTime := time.Now()
	rows, err := q.Query(ctx, `
	SELECT id,  event_id::text, payload FROM outbox
	WHERE sent_at IS NULL
	ORDER BY created_at
	LIMIT 10
	FOR UPDATE SKIP LOCKED`)
	metrics.ObserveRequestDuration(startTime, databaseScope, "outbox")

	if err != nil {
		return nil, fmt.Errorf("could not get outbox rows: %w", err)
	}
	defer rows.Close()

	var events []scrapper.OutboxEvent
	for rows.Next() {
		var (
			event   scrapper.OutboxEvent
			payload []byte
		)

		if err = rows.Scan(&event.ID, &event.EventID, &payload); err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}

		var update pkg.LinkUpdate
		if err = json.Unmarshal(payload, &update); err != nil {
			return nil, fmt.Errorf("could not unmarshal update: %w", err)
		}

		event.Payload = update
		events = append(events, event)
	}

	return events, nil
}

func (r *OutboxRepository) UpdateSendTime(ctx context.Context, id int64) error {
	q := database.GetQuerier(ctx, r.db)

	_, err := q.Exec(ctx, `
	UPDATE outbox SET sent_at = $1
	WHERE id = $2`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("could not update status: %w", err)
	}
	return nil
}

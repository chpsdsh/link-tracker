package metricsrepository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/metrics"
)

const databaseScope = "database"

type LinksMetricRepository struct {
	db *pgxpool.Pool
}

func NewLinksMetricRepository(db *pgxpool.Pool) *LinksMetricRepository {
	return &LinksMetricRepository{db: db}
}

func (r LinksMetricRepository) CountLinksOnTrack(ctx context.Context) (int, int, error) {
	var gitCount int
	startTime := time.Now()
	err := r.db.QueryRow(ctx, `
	SELECT COUNT(*) FROM links l
	WHERE l.url LIKE $1`, "%github%").Scan(&gitCount)
	metrics.ObserveRequestDuration(startTime, databaseScope, "links")
	if err != nil {
		return 0, 0, fmt.Errorf("error counting github links: %w", err)
	}

	var stackOverflowCount int
	startTime = time.Now()
	err = r.db.QueryRow(ctx, `
	SELECT COUNT(*) FROM links l
	WHERE l.url LIKE $1`, "%stackoverflow%").Scan(&stackOverflowCount)
	metrics.ObserveRequestDuration(startTime, databaseScope, "links")
	if err != nil {
		return 0, 0, fmt.Errorf("error counting stackOverflow links: %w", err)
	}
	return gitCount, stackOverflowCount, nil
}

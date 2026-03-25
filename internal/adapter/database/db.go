package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/scrapper/config"
)

const (
	postgresDialect     = "postgres"
	migrationsDirectory = "migrations"
	minConnections      = 2
	maxConnections      = 10
	maxConnIdleTime     = 100 * time.Millisecond
	maxConnLifetime     = time.Second
)

var (
	ErrCreatingDatabaseFromUrl = errors.New("migrating database error")
	ErrMigrating               = errors.New("migrating database error")
	ErrCreatingConnectionPool  = errors.New("creating connection pool error")
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(config config.PostgresConfig) (*DB, error) {
	dsn := createDataBaseUrlFromConfig(config)

	connConf, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, errors.Join(err, ErrCreatingDatabaseFromUrl)
	}

	if err = migrate(connConf); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMigrating, err)
	}

	poolConf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.Join(err, ErrCreatingDatabaseFromUrl)
	}
	poolConf.MinConns = minConnections
	poolConf.MaxConns = maxConnections
	poolConf.MaxConnLifetime = maxConnLifetime
	poolConf.MaxConnIdleTime = maxConnIdleTime

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConf)
	if err != nil {
		return nil, errors.Join(err, ErrCreatingConnectionPool)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) CloseConnectionPool() {
	db.pool.Close()
}

func migrate(cfg *pgx.ConnConfig) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(postgresDialect); err != nil {
		return errors.Join(err, ErrMigrating)
	}
	db := stdlib.OpenDB(*cfg)
	defer func() { _ = db.Close() }()

	if err := goose.Up(db, migrationsDirectory); err != nil {
		return errors.Join(err, ErrMigrating)
	}
	return nil
}

func createDataBaseUrlFromConfig(cfg config.PostgresConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)
}

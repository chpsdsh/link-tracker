package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	ErrCreatingDatabaseFromURL = errors.New("migrating database error")
	ErrMigrating               = errors.New("migrating database error")
	ErrCreatingConnectionPool  = errors.New("creating connection db error")
)

type Querier interface {
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func GetQuerier(ctx context.Context, defaultQuerier Querier) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return defaultQuerier
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

type DB struct {
	db *pgxpool.Pool
}

func NewDB(config config.PostgresConfig) (*DB, error) {
	dsn := createDataBaseURLFromConfig(config)

	connConf, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, errors.Join(err, ErrCreatingDatabaseFromURL)
	}

	if err = migrate(connConf); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMigrating, err)
	}

	poolConf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.Join(err, ErrCreatingDatabaseFromURL)
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

	return &DB{db: pool}, nil
}

func (db *DB) GetDBPool() *pgxpool.Pool {
	return db.db
}

func (db *DB) CloseConnectionPool() {
	db.db.Close()
}

func (db *DB) Transaction(ctx context.Context, txFunc func(ctx context.Context) error) error {
	tx, err := db.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	err = txFunc(injectTx(ctx, tx))
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (db *DB) TransactionWithReturn(ctx context.Context, txFunc func(ctx context.Context) (any, error)) (any, error) {
	tx, err := db.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	value, err := txFunc(injectTx(ctx, tx))
	if err != nil {
		return nil, fmt.Errorf("get transaction from ctx: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return value, nil
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

func createDataBaseURLFromConfig(cfg config.PostgresConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

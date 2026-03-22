package config

import (
	"errors"
	"os"
)

var (
	ErrNoPostgresHost     = errors.New("postgres host should be set with " + postgresHost + " environment variable")
	ErrNoPostgresPort     = errors.New("postgres port should be set with " + postgresPort + " environment variable")
	ErrNoPostgresUser     = errors.New("postgres user should be set with " + postgresUser + " environment variable")
	ErrNoPostgresPassword = errors.New("postgres password should be set with " + postgresPassword + " environment variable")
	ErrNoPostgresDB       = errors.New("postgres db name should be set with " + postgresDatabaseName + " environment variable")
)

const (
	postgresHost         = "POSTGRES_HOST"
	postgresPort         = "POSTGRES_PORT"
	postgresUser         = "POSTGRES_USER"
	postgresPassword     = "POSTGRES_PASSWORD"
	postgresDatabaseName = "POSTGRES_DB"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func ParsePostgresConfig() (PostgresConfig, error) {
	pHost := os.Getenv(postgresHost)
	if pHost == "" {
		return PostgresConfig{}, ErrNoPostgresHost
	}
	pPort := os.Getenv(postgresPort)
	if pPort == "" {
		return PostgresConfig{}, ErrNoPostgresPort
	}
	pUser := os.Getenv(postgresUser)
	if pUser == "" {
		return PostgresConfig{}, ErrNoPostgresUser
	}
	pPassword := os.Getenv(postgresPassword)
	if pPassword == "" {
		return PostgresConfig{}, ErrNoPostgresPassword
	}
	pDB := os.Getenv(postgresDatabaseName)
	if pDB == "" {
		return PostgresConfig{}, ErrNoPostgresDB
	}
	return PostgresConfig{
		Host:     pHost,
		Port:     pPort,
		User:     pUser,
		Password: pPassword,
		DBName:   pPassword,
	}, nil
}

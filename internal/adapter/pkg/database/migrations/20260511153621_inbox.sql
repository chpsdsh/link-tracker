-- +goose Up
SELECT 'up SQL query';

CREATE TABLE inbox
(
    event_id     uuid primary key,
    received_at  timestamptz not null default now(),
    processed_at timestamptz
);

-- +goose Down
SELECT 'down SQL query';

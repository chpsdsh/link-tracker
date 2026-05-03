-- +goose Up
create table outbox
(
    id         bigserial primary key,
    event_type text not null,
    payload    json not null,
    created_at timestamptz default now(),
    sent_at    timestamptz
);

-- +goose Down
SELECT 'down SQL query';

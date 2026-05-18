-- +goose Up
create table outbox
(
    id         bigserial primary key,
    event_type text  not null,
    event_id   uuid  not null default gen_random_uuid(),
    payload    jsonb not null,
    created_at timestamptz    default now(),
    sent_at    timestamptz
);

-- +goose Down
drop table outbox;

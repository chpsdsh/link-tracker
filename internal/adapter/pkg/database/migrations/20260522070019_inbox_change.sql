-- +goose Up
alter table inbox
    add column consumer_name text;

alter table inbox
    DROP constraint inbox_pkey;

alter table inbox
    add primary key (consumer_name, event_id);


-- +goose Down
SELECT 'down SQL query';

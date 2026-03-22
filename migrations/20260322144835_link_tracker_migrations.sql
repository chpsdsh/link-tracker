-- +goose Up
-- +goose StatementBegin
create table chats
(
    chat_id bigint primary key
);

create table links
(
    id  bigserial primary key,
    url text unique not null
);

create table tags
(
    id  bigserial primary key,
    tag text unique not null
);

create table link_chat
(
    chat_id bigint references chats (chat_id) on delete cascade,
    link_id bigint references links (id) on delete cascade,
    primary key (chat_id, link_id)
);

create table link_tag
(
    link_id bigint references links (id) on delete cascade,
    tag_id bigint references tags(id) on delete cascade,
    primary key (link_id, tag_id)
);
-- +goose StatementEnd
    
-- +goose Down
drop table link_tag;
drop table link_chat;
drop table tags;
drop table links;
drop table chats;
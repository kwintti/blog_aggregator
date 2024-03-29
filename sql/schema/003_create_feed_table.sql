-- +goose Up
create table feeds (
    id uuid primary key,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    name text not null,
    url text not null unique,
    user_id uuid not null references users (id) on delete cascade
);

-- +goose Down
drop table feeds;

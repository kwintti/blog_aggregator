-- +goose Up
create table feed_follow (
    id uuid primary key,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    user_id uuid not null references users (id) on delete cascade,
    feed_id uuid not null references feeds (id) on delete cascade
);

-- +goose Down
drop table feed_follow;

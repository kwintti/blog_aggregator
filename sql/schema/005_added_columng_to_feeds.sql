-- +goose Up
alter table feeds
add column last_fetch_at timestamptz;

-- +goose Down
alter table feeds
drop column last_fetch_at;

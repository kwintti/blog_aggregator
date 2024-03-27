-- +goose Up
CREATE TABLE users (
    id uuid Primary Key,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    name TEXT not null
);

-- +goose Down
DROP TABLE users; 

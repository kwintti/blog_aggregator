-- name: CreateUser :one 
insert into users (id, created_at, updated_at, name, apikey)
values($1, $2, $3, $4, encode(sha256(random()::text::bytea), 'hex'))
returning *;

-- name: RetriveUser :one
select * from users
where apikey = $1;

-- name: AddFeed :one
insert into feeds (id, created_at, updated_at, name, url, user_id)
values($1, $2, $3, $4, $5, $6)
returning *;

-- name: ListFeeds :many
select * from feeds;

-- name: AddFeedFollow :one
insert into feed_follow (id, created_at, updated_at, user_id, feed_id)
values($1, $2, $3, $4, $5)
returning *;

-- name: GetFollowFeed :one
select * from feed_follow
where feed_id = $1;

-- name: AllUserFeedFollows :many
select * from feed_follow
where user_id = $1;

-- name: DeleteFeed :exec 
delete from feed_follow
where user_id = $1
and id = $2;

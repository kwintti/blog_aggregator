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

-- name: GetNextFeedsToFetch :many
select * from feeds
order by last_fetch_at nulls first
limit $1;

-- name: MarkFeedFetched :one
update feeds
set updated_at = $1, last_fetch_at = $2
where id = $3 
returning *;

-- name: CreatePost :one
insert into posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
values($1, $2, $3, $4, $5, $6, $7, $8)
returning *;

-- name: GetPostByUser :many
select * from posts
inner join feeds
on posts.feed_id = feeds.id
where feeds.user_id = $1
order by posts.updated_at desc
limit $2;

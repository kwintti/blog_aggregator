# Blog aggregator

This blog aggregator is written in Go as part of guided project from boot.dev.

- You can add users and feeds
- You can follow and unfollow feeds
- Automatic scraping of feeds every 60s
- We are using Postgres as db here

## API

### Add users

POST "/v1/users"

Body:
```json
{
    "name": "MeUser"
}
```

RESPONSE:

```json
{
    "feed": {
        "created_at": "2024-04-02T22:15:41.570887+03:00",
        "id": "68b1c5...4f00c9ce7",
        "last_fetched_at": null,
        "name": "Mashable",
        "updated_at": "2024-04-02T22:15:41.570887+03:00",
        "url": "https://mashable.com/feeds/rss/all",
        "user_id": "576b13...c6c713ffdad"
    },
    "feed_follow": {
        "created_at": "2024-04-02T22:15:41.570887+03:00",
        "feed_id": "68b1c...f00c9ce7",
        "id": "60883b...c85dd4ec03a",
        "updated_at": "2024-04-02T22:15:41.570887+03:00",
        "user_id": "576b13...c6c713ffdad"
    }
}
```
### List users

GET "/v1/users"

Header:
```json
{
    "Authorizarion": "ApiKey 34c233...3b44ba4e0843e1e953e34 "
}
```

RESPONSE:

```json

[{
    "id": "576...6c713ffdad",
    "created_at": "2024-04-02T22:15:41.570887+03:00", 
    "updated_at": "2024-04-02T22:15:41.570887+03:00", 
    "name": "MeUser",
    "apikey": "34c233e53...0843e1e953e34",
},
{
    "id": "576b...b-7c6c713ffdad"
    "created_at": "2024-04-02T22:15:41.570887+03:00" 
    "updated_at": "2024-04-02T22:15:41.570887+03:00" 
    "name": "SecondUser"
    "apikey": "34c233...953e34"
}]
```
### Add new feed

POST "/v1/feeds"

Header:
```json
{
    "Authorizarion": "ApiKey 34c233e5...0843e1e953e34 "
}
```

RESPONSE:

```json

{
    "id": "576...713ffdad",
    "created_at": "2024-04-02T22:15:41.570887+03:00",
    "updated_at": "2024-04-02T22:15:41.570887+03:00",
    "url": "www.example.com/index.xml",
    "user_id": "777b...c713ffdad",
    "last_fetched_at": "2024-04-02T22:15:41.570887+03:00" 
}
```
### List all feeds

GET "/v1/feeds"

RESPONSE:

```json

[
    {
        "created_at": "2024-04-02T22:15:41.570887+03:00",
        "id": "68b1c...00c9ce7",
        "last_fetched_at": "2024-04-02T23:53:06.684776+03:00",
        "name": "Mashable",
        "updated_at": "2024-04-02T23:53:06.684776+03:00",
        "url": "https://mashable.com/feeds/rss/all",
        "user_id": "576...13ffdad"
    },
    {
        "created_at": "2024-03-30T01:34:36.87611+02:00",
        "id": "0a41b223...a8fbaf4ac",
        "last_fetched_at": null,
        "name": "facebook",
        "updated_at": "2024-04-02T23:53:07.123745+03:00",
        "url": "https://wagslane.dev/index.xml",
        "user_id": "ec18d6cd-4...9e6628"
    }
]
```
### Follow a feed

POST "/v1/feed_follows"

Header:
```json
{
    "Authorizarion": "ApiKey 34c233e5361bb647de91...834023b44ba4e0843e1e953e34 "
}
```

Body:
```json
{
    "feed_id": "68b1c543-0...0c9ce7"
}
```

RESPONSE:

```json
{
    "id": "7771c5...44f00c9ce7",
    "created_at": "2024-04-02T23:53:07.123745+03:00",
    "updated_at": "2024-04-02T23:53:07.123745+03:00",
    "user_id": "7771c5...00c9ce7",
    "feed_id": "68b1c543...f00c9ce7"
}
```
### List all follows for user

POST "/v1/feed_follows"

Header:
```json
{
    "Authorizarion": "ApiKey 34c233e5361bb647de91d643fb...b44ba4e0843e1e953e34 "
}
```

RESPONSE:

```json
[
    {
        "created_at": "2024-03-30T00:13:31.125034+02:00",
        "feed_id": "ec18d6cd-...99e6628",
        "id": "2478e...b06a30bf976",
        "updated_at": "2024-03-30T00:13:31.125034+02:00",
        "user_id": "ec18d6c...2c99e6628"
    },
    {
        "created_at": "2024-03-30T01:34:36.87611+02:00",
        "feed_id": "ec18d6...32c99e6628",
        "id": "0c1d05...7cb78c70b6",
        "updated_at": "2024-03-30T01:34:36.87611+02:00",
        "user_id": "ec18d6cd...c99e6628"
    }
]
```
### Unfollow a feed

DELETE /v1/feed_follows/{feedFollowID}

Header:
```json
{
    "Authorizarion": "ApiKey 34c233e5361bb647de91d643f...3b44ba4e0843e1e953e34 "
}
```

### List all posts in the feeds followed by an user

GET "/v1/posts"

You can limit results with `limit=10` parameter

Header:
```json
{
    "Authorizarion": "ApiKey 34c233e5361bb647de91d6...4023b44ba4e0843e1e953e34 "
}
```

RESPONSE:

```json
[
    {
        "created_at": "2024-04-02T15:10:45.579289Z",
        "description": "Pythogoras escaped this month. The community rallied against the Serpent God, and while he was wounded and beaten back, he escaped.",
        "feed_id": "aa63085a...16f8dfa5",
        "id": "a2fe8eb...c4b438cf1",
        "publsihed_ad": "2024-02-28T00:00:00Z",
        "title": "The Boot.dev Beat. March 2024",
        "updated_at": "2024-04-02T15:10:45.579289Z",
        "url": "https://blog.boot.dev/news/bootdev-beat-2024-03/"
    },
    {
        "created_at": "2024-04-02T15:10:45.579289Z",
        "description": "609,179. That&rsquo;s the number of lessons you crazy folks have completed on Boot.dev in the last 30 days.",
        "feed_id": "aa63085...216f8dfa5",
        "id": "2bfd7fc9-...ce53679",
        "publsihed_ad": "2024-01-31T00:00:00Z",
        "title": "The Boot.dev Beat. February 2024",
        "updated_at": "2024-04-02T15:10:45.579289Z",
        "url": "https://blog.boot.dev/news/bootdev-beat-2024-02/"
    },
    {
        "created_at": "2024-04-02T15:10:45.579289Z",
        "description": "Theo has this great video on Kubernetes, currently titled &ldquo;You Don&rsquo;t Need Kubernetes&rdquo;. I&rsquo;m a Kubernetes enjoyer, but I&rsquo;m not here to argue about
 that.",
        "feed_id": "aa630...2a216f8dfa5",
        "id": "fb1ca8c...7fbb953b5",
        "publsihed_ad": "2024-03-08T00:00:00Z",
        "title": "Maybe You Do Need Kubernetes",
        "updated_at": "2024-04-02T15:10:45.579289Z",
        "url": "https://blog.boot.dev/education/maybe-you-do-need-kubernetes/"
    }
]

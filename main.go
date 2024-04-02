package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kwintti/blog_aggregator/internal/database"
	"github.com/lib/pq"
)

func main(){
    godotenv.Load()
    CONN := os.Getenv("CONN")
    db, err := sql.Open("postgres", CONN)
    if err != nil {
        log.Print("Error connecting to database", err)
    }
    defer db.Close()
    dbQueries := database.New(db)
    apiConfig := &apiConfig{}
    apiConfig.DB = dbQueries
    port := os.Getenv("PORT")
    mux := http.NewServeMux()
    mux.HandleFunc("GET /v1/readiness", readinessHandler)
    mux.HandleFunc("GET /v1/err", errorHandler)
    mux.HandleFunc("POST /v1/users", apiConfig.handlerAddUser)
    mux.HandleFunc("GET /v1/users", apiConfig.middlewareAuth(apiConfig.handlerGetUserInfo))
    mux.HandleFunc("POST /v1/feeds", apiConfig.middlewareAuth(apiConfig.handlerCreateFeed))
    mux.HandleFunc("GET /v1/feeds", apiConfig.handlerGetFeeds)
    mux.HandleFunc("POST /v1/feed_follows", apiConfig.middlewareAuth(apiConfig.handlerCreateFeedFollow))
    mux.HandleFunc("GET /v1/feed_follows", apiConfig.middlewareAuth(apiConfig.handlerGetFeedFollows))
    mux.HandleFunc("GET /v1/posts", apiConfig.middlewareAuth(apiConfig.handlerGetPosts))
    mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", apiConfig.middlewareAuth(apiConfig.handlerDeleteFeed))
    corsMux := middlewareCors(mux)
    server := &http.Server{
        Addr: ":"+port,
        Handler: corsMux,
    }
    fmt.Printf("Serving on port %s ...", port)
    ctx, cancelFunc := context.WithCancel(context.Background())
    apiConfig.startScheduledFetching(ctx)
    log.Fatal(server.ListenAndServe())
    cancelFunc()
}


type apiConfig struct {
	DB *database.Queries
}

type params struct {
    Name string `json:"name"`
}

type user struct {
    Id uuid.UUID `json:"id"`
    Created_at time.Time `json:"created_at"`
    Updated_at time.Time `json:"updated_at"`
    Name string `json:"name"`
    ApiKey string `json:"apikey"`
}

type FeedFollows struct {
    Id uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    UserId uuid.UUID `json:"user_id"`
    FeedId uuid.UUID `json:"feed_id"`
} 

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) handlerGetFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {
    ctx := r.Context()
    rows, err := cfg.DB.AllUserFeedFollows(ctx, user.ID)
    if err != nil {
        log.Print("Couldn't get feeds: ", err)
        respondWithError(w, 500, "Couldn't get feeds")
        return
    }
    var outpuFeeds []FeedFollows
    for _, row := range rows {
        outpuFeeds = append(outpuFeeds, FeedFollows{
                                                    Id: row.ID,
                                                    CreatedAt: row.CreatedAt,
                                                    UpdatedAt: row.UpdatedAt,
                                                    UserId: row.UserID,
                                                    FeedId: row.UserID,
                                                    })
    }
    respondWithJSON(w, 200, outpuFeeds)

    
}

type addFeedFollow struct {
    Id uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    UserId uuid.UUID `json:"user_id"`
    FeedId uuid.UUID `json:"feed_id"`
}

func (cfg *apiConfig) handlerCreateFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
    type paramsFeed struct {
        FeedId uuid.UUID `json:"feed_id"`
    }
    decoder := json.NewDecoder(r.Body) 
    params := paramsFeed{}
    err := decoder.Decode(&params)
    if err != nil {
        log.Print("Couldn't parse params ", err)
        respondWithError(w, 500, "Couldn't parse params")
        return
    }
    timeNow := time.Now().UTC()
    ctx := r.Context()
    createdFollowFeed, err := cfg.DB.AddFeedFollow(ctx, database.AddFeedFollowParams{
                                                                                    ID: uuid.New(),
                                                                                    CreatedAt: timeNow,
                                                                                    UpdatedAt: timeNow,
                                                                                    UserID: user.ID,
                                                                                    FeedID: params.FeedId,
                                                                                    })
    if err != nil {
        if err == sql.ErrNoRows {
            fmt.Print("hello")
        }
        log.Print("Couldn't create feed follow ", err)
        respondWithError(w, 500, "Couldn't create feed follow")
        return
    }
    responseFeedFollow := addFeedFollow{Id: createdFollowFeed.ID,
                                        CreatedAt: createdFollowFeed.CreatedAt,
                                        UpdatedAt: createdFollowFeed.UpdatedAt,
                                        UserId: createdFollowFeed.UserID, 
                                        FeedId: createdFollowFeed.FeedID,}

    respondWithJSON(w, 201, responseFeedFollow)
}

func (cfg *apiConfig) handlerDeleteFeed(w http.ResponseWriter, r *http.Request, user database.User) {
    id := r.PathValue("feedFollowID")
    idUUID, err := uuid.Parse(id)
    if err != nil {
        log.Print("Couldn't transfer string to UUID: ", err)
        respondWithError(w, 500, "Couldn't parse uuid")
        return
    }
    ctx := r.Context()
    err = cfg.DB.DeleteFeed(ctx, database.DeleteFeedParams{UserID: user.ID, ID: idUUID,})
    if err != nil {
        log.Print("Couldn't unfollow feed: ", err)
        respondWithError(w, 500, "Couldn't unfollow the feed")
        return
    }
    w.WriteHeader(204)
}


func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        apikey := r.Header.Get("Authorization")
        apikeyParsed := strings.TrimPrefix(apikey, "ApiKey ")
        ctx := r.Context()
        user, err := cfg.DB.RetriveUser(ctx, apikeyParsed)
        if err != nil {
            if err == sql.ErrNoRows {
                log.Print("No results found ", err)
                respondWithError(w, http.StatusUnauthorized, "User results found")
            } else {
                log.Print("Error getting user info ", err)
                respondWithError(w, http.StatusInternalServerError, "Error getting user from db")
            }
            return

        }
        handler(w, r, user)  
    }
}

func (cfg *apiConfig) handlerGetFeeds(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    rows, err := cfg.DB.ListFeeds(ctx)
    if err != nil {
        log.Print("Couldn't retrive feeds")
        respondWithError(w, 500, "Couldn't retrive feeds")
        return
    }
    var feeds []feed
    for _,row := range rows {
        feeds = append(feeds, feed{ Id: row.ID,
                                    CreatedAt: row.CreatedAt,
                                    UpdateAt: row.UpdatedAt,
                                    Name: row.Name,
                                    Url: row.Url,
                                    UserId: row.UserID,
                                })
    }
    if feeds != nil {
        respondWithJSON(w, 200, feeds)
    } else {
        respondWithError(w, 500, "No feeds available") 
    }
}


func readinessHandler(w http.ResponseWriter, r *http.Request) {
    type resp struct {
        Status string `json:"status"`
    }
    respondWithJSON(w, 200, resp{Status: "ok"})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
    respondWithError(w, 500, "Internal Server Error")    
}

func middlewareCors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "*")
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)

        })
}

func (cfg *apiConfig) getNextFeedsToFetch(numberOfRows int32) []feed {    
    ctx := context.TODO()
    rows, err := cfg.DB.GetNextFeedsToFetch(ctx, numberOfRows) 
    if err != nil {
        if err == sql.ErrNoRows {
            log.Print("No rows found: ", err) 
        } else {
            log.Print("Couldn't retrive from database: ", err)
        }
    }
    var rowsOut []feed
    for _, row := range rows {
        rowsOut = append(rowsOut, databaseFeedToFeed(row)) 
    }
    return rowsOut
}

func (cfg *apiConfig) markFeedFetched(feedId uuid.UUID) database.Feed{
    var convertedToNullTime sql.NullTime

    ctx := context.TODO() 
    
    timeNow := time.Now().UTC()
    convertedToNullTime.Time = timeNow
    convertedToNullTime.Valid = true

    updatedFeed, err := cfg.DB.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{ID: feedId, LastFetchAt: convertedToNullTime, UpdatedAt: timeNow})
    if err != nil {
        log.Print("Couldn't update feed")
        return database.Feed{}
    }
    return updatedFeed

}
type xmlParse struct {
    Channel struct {
        Title string `xml:"title"`
        Link  struct {
            Text string `xml:",chardata"`
            Href string `xml:"href,attr"`
            Rel  string `xml:"rel,attr"`
            Type string `xml:"type,attr"`
        } `xml:"link"`
        Description   string `xml:"description"`
        Generator     string `xml:"generator"`
        Language      string `xml:"language"`
        LastBuildDate string `xml:"lastBuildDate"`
        Items         []Item `xml:"item"`
    } `xml:"channel"`
}    

type Item struct {
    Title       string `xml:"title"`
    Link        string `xml:"link"`
    PubDate     string `xml:"pubDate"`
    Guid        string `xml:"guid"`
    Description string `xml:"description"`
}


func fetchXmlData(url string) xmlParse{
    
    resp, err := http.Get(url)
    if err != nil {
        log.Print("Couldn't fetch from url: ", err)
        return xmlParse{}
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil { 
        log.Print("Couldn't read body from response: ", err)
    }
    resp.Body.Close()
    v := xmlParse{} 
    err = xml.Unmarshal([]byte(body), &v)
    if err != nil {
        log.Print("Couldn't parse xml: ", err)
        return xmlParse{}
    }
    return v 
}

func (cfg *apiConfig) fetchingFeeds(wg *sync.WaitGroup, ctx context.Context) {
    feeds := cfg.getNextFeedsToFetch(10) 
    timeNow := time.Now().UTC()
    for _, feed := range feeds {
        feedData := fetchXmlData(feed.Url) 
        for _, item := range feedData.Channel.Items {
            layout := "Mon, 02 Jan 2006 15:04:05 +0000"
            timeOfPub, err := time.Parse(layout, strings.TrimSpace(item.PubDate)) 
            if err != nil {
                log.Print("Couldn't parse publication time: ", err)
            }
            var timeOfPubNullable sql.NullTime 
            if timeOfPub.IsZero() {
                timeOfPubNullable.Valid = false
            } else {
                timeOfPubNullable.Time = timeOfPub
                timeOfPubNullable.Valid = true
            }

            var titleNullable sql.NullString
            if len(item.Title) == 0 {
                titleNullable.Valid = false
            } else {
                titleNullable.String = item.Title
                titleNullable.Valid = true
            }

            var urlNullable sql.NullString
            if len(item.Link) == 0 {
                urlNullable.Valid = false
            } else {
                urlNullable.String = item.Link
                urlNullable.Valid = true
            }

            var descrNullable sql.NullString
            if len(item.Description) == 0 {
                descrNullable.Valid = false
            } else {
                descrNullable.String = item.Description
                descrNullable.Valid = true
            }

            _, err = cfg.DB.CreatePost(ctx, database.CreatePostParams{
                                                                ID: uuid.New(),
                                                                CreatedAt: timeNow,
                                                                UpdatedAt: timeNow,
                                                                Title: titleNullable,
                                                                Url: urlNullable,
                                                                Description: descrNullable,
                                                                PublishedAt: timeOfPubNullable,
                                                                FeedID: feed.Id,
                                                            })
            if err != nil {
                pqErr := err.(*pq.Error) 
                if pqErr.Code.Name() == "unique_violation" {
                    continue
                } else {
                    log.Println("Couldn't create post: ", err)
                }
            } 
        }
        cfg.markFeedFetched(feed.Id)    
    }
}

func (cfg *apiConfig) startScheduledFetching(ctx context.Context) {
    var wg sync.WaitGroup
    ticker := time.NewTicker(60 * time.Second)
    go func() {
        for {
            wg.Add(1)
            go func() {
                defer wg.Done()
                cfg.fetchingFeeds(&wg, ctx)
            }()
            select {
            case <-ticker.C:
                continue
            case <-ctx.Done():
                ticker.Stop()
                wg.Wait()
                return
            }
        }
    }()
}

type PostJSON struct {
    Id  uuid.UUID   `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Title   *string      `json:"title"`
    Url     *string      `json:"url"`
    Description *string  `json:"description"`
    PublishedAt *time.Time `json:"publsihed_ad"`
    FeedId      uuid.UUID `json:"feed_id"`
}

func databasePostToPost(inputFeed database.GetPostByUserRow) PostJSON {
    var convertedPubAt *time.Time
    
    if inputFeed.PublishedAt.Valid {
        convertedPubAt = &inputFeed.PublishedAt.Time
    } else {
        convertedPubAt = nil
    }

    var convertedTitle *string
    if inputFeed.Title.Valid {
        convertedTitle = &inputFeed.Title.String
    } else {
        convertedTitle = nil
    }

    var convertedUrl *string
    if inputFeed.Url.Valid {
        convertedUrl = &inputFeed.Url.String
    } else {
        convertedUrl = nil 
    }

    var convertedDescription *string
    if inputFeed.Url.Valid {
        convertedDescription = &inputFeed.Description.String
    } else {
        convertedUrl = nil
    }

    ouputFeed := PostJSON{
                    Id: inputFeed.ID,
                    CreatedAt: inputFeed.CreatedAt,
                    UpdatedAt: inputFeed.UpdatedAt,
                    Title: convertedTitle,
                    Url: convertedUrl,
                    Description: convertedDescription,
                    PublishedAt: convertedPubAt, 
                    FeedId: inputFeed.FeedID,
                }

    return ouputFeed 
}




func (cfg *apiConfig) handlerGetPosts(w http.ResponseWriter, r *http.Request, user database.User) {
    ctx := r.Context()
    var limit int
    var err error
    param := r.URL.Query().Get("limit")
    if len(param) == 0 {
        limit = 100
    } else {
        limit, err = strconv.Atoi(param)
        if err != nil {
            log.Println("Couldn't parse url argument: ", err)
        }
    }
    rows, err := cfg.DB.GetPostByUser(ctx, database.GetPostByUserParams{UserID: user.ID, Limit: int32(limit) })
    if err != nil {
        log.Print(err)
    }
    var output []PostJSON
    for _, item := range rows {
        output = append(output, databasePostToPost(item))
    }
    respondWithJSON(w, 200, output)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    ansMSG, err := json.Marshal(payload)
    if err != nil {
        log.Print("Error with enconding json occured", err)
        w.WriteHeader(500)
        return
    }
    w.Write(ansMSG)
    
}

func respondWithError(w http.ResponseWriter, status int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    type respond struct {
        Error    string `json:"error"`
    }
    ansMSG, err := json.Marshal(respond{Error: msg})
    if err != nil {
        log.Print("Error with enconding json occured", err)
        w.WriteHeader(500)
        return
    }
    w.Write(ansMSG)
    
}


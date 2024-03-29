package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kwintti/blog_aggregator/internal/database"
	_ "github.com/lib/pq"
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
    mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", apiConfig.middlewareAuth(apiConfig.handlerDeleteFeed))
    corsMux := middlewareCors(mux)
    server := &http.Server{
        Addr: ":"+port,
        Handler: corsMux,
    }
    fmt.Printf("Serving on port %s ...", port)
    log.Fatal(server.ListenAndServe())

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


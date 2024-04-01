package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kwintti/blog_aggregator/internal/database"
	_ "github.com/lib/pq"
)

func (cfg *apiConfig) handlerAddUser(w http.ResponseWriter, r *http.Request) {
    decoder := json.NewDecoder(r.Body)
    params := params{}
    err := decoder.Decode(&params)
    if err != nil {
        log.Print("Error decoding body from json", err)
        respondWithError(w, http.StatusInternalServerError, "Error decoding params.")
    }
    timeNow := time.Now().UTC()
    newUser := user {
                    Id: uuid.New(),
                    Created_at: timeNow,
                    Updated_at: timeNow,
                    Name: params.Name,}
     
    ctx := r.Context()
    createdUser, err := cfg.DB.CreateUser(ctx, database.CreateUserParams{
                                ID: newUser.Id,
                                CreatedAt: newUser.Created_at,
                                UpdatedAt: newUser.Updated_at,
                                Name: newUser.Name,
                                })
    if err != nil {
        log.Print("Error creating user in database ", err)
        respondWithError(w, http.StatusInternalServerError, "Error creating user.")
    }
    userFromDb := user {
        Id: createdUser.ID,
        Created_at: createdUser.CreatedAt,
        Updated_at: createdUser.UpdatedAt,
        Name: createdUser.Name,
    }
    respondWithJSON(w, 201, userFromDb)
}

type paramsToken struct{
    Apikey string `json:"apikey"`
}

func (cfg *apiConfig) handlerGetUserInfo(w http.ResponseWriter, r *http.Request, infoUser database.User) {
    returnUserInfo := user {
        Id: infoUser.ID, 
        Created_at: infoUser.CreatedAt,
        Updated_at: infoUser.UpdatedAt,
        Name: infoUser.Name,
        ApiKey: infoUser.Apikey,
    }
    respondWithJSON(w, 200, returnUserInfo)
    
}
type feed struct {
    Id uuid.UUID `json:"id"` 
    Name string `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdateAt time.Time `json:"updated_at"`
    Url string `json:"url"`
    UserId uuid.UUID `json:"user_id"`
    LastFetchAt *time.Time `json:"last_fetched_at"`
}


func databaseFeedToFeed(inputFeed database.Feed) feed {
    var convertedTime *time.Time
    
    if inputFeed.LastFetchAt.Valid {
        convertedTime = &inputFeed.LastFetchAt.Time
    } else {
        convertedTime = nil
    }

    ouputFeed := feed{
                    Id: inputFeed.ID,
                    CreatedAt: inputFeed.CreatedAt,
                    UpdateAt: inputFeed.UpdatedAt,
                    Name: inputFeed.Name,
                    Url: inputFeed.Url,
                    UserId: inputFeed.UserID,
                    LastFetchAt: convertedTime, 
                }

    return ouputFeed 
}


func (cfg *apiConfig) handlerCreateFeed(w http.ResponseWriter, r *http.Request, user database.User) {
    type paramsFeed struct {
        Name string `json:"name"`
        Url string `json:"url"`
    }
    decoder := json.NewDecoder(r.Body)
    params := paramsFeed{}
    err := decoder.Decode(&params)
    if err != nil {
        log.Print("Error decoding json ", err)
        respondWithError(w, 500, "Error decoding json")
        return
    }
    ctx := r.Context()
    timeNow := time.Now().UTC()
    feedID := uuid.New()
    newFeed, err := cfg.DB.AddFeed(ctx, database.AddFeedParams{
                                                            ID: feedID,
                                                            CreatedAt: timeNow,
                                                            UpdatedAt: timeNow,
                                                            Name: params.Name,
                                                            Url: params.Url,
                                                            UserID: user.ID,})
    if err != nil {
        log.Print("Cannot add to database ", err)
        respondWithError(w, 500, "Cannot add to database")
        return 
    }
    feedFollow, err := cfg.DB.AddFeedFollow(ctx, database.AddFeedFollowParams{
                                                                            ID: uuid.New(),
                                                                            CreatedAt: timeNow,
                                                                            UpdatedAt: timeNow,
                                                                            UserID: user.ID,
                                                                            FeedID: feedID,
                                                                            })
    if err != nil {
        log.Print("Couldn't add a feed follow: ", err)
        respondWithError(w, 500, "Couldn't add a feed follow")
        return
    }
    addedFeed := databaseFeedToFeed(newFeed)
    addedFeedFollow := addFeedFollow{
                    Id: feedFollow.ID,
                    CreatedAt: feedFollow.CreatedAt,
                    UpdatedAt: feedFollow.UpdatedAt,
                    UserId: feedFollow.UserID,
                    FeedId: feedFollow.FeedID,
                    }
    type feedAndFollow struct {
        Feed feed `json:"feed"`
        FeedFollow addFeedFollow `json:"feed_follow"`
    }

    outputFeedAndFollow := feedAndFollow{
                    Feed: addedFeed,
                    FeedFollow: addedFeedFollow,
                    }


    respondWithJSON(w, 201, outputFeedAndFollow)

}

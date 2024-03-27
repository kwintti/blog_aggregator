package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
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

func (cfg *apiConfig) handlerGetUserInfo(w http.ResponseWriter, r *http.Request) {
    apikey := r.Header.Get("Authorization")
    apikeyParsed := strings.TrimPrefix(apikey, "ApiKey ")
    ctx := r.Context()
    getUserInfo, err := cfg.DB.RetriveUser(ctx, apikeyParsed)
    if err != nil {
        log.Print("Error getting user info ", err)
        respondWithError(w, http.StatusInternalServerError, "Error getting user from db")
    }
    returnUserInfo := user {
        Id: getUserInfo.ID, 
        Created_at: getUserInfo.CreatedAt,
        Updated_at: getUserInfo.UpdatedAt,
        Name: getUserInfo.Name,
        ApiKey: getUserInfo.Apikey,
    }
    respondWithJSON(w, 200, returnUserInfo)
    
}

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

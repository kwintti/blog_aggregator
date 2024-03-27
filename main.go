package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
    mux.HandleFunc("GET /v1/users", apiConfig.handlerGetUserInfo)
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


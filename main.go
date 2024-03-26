package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main(){
    godotenv.Load()
    port := os.Getenv("PORT")
    mux := http.NewServeMux()
    corsMux := middlewareCors(mux)
    server := &http.Server{
        Addr: ":"+port,
        Handler: corsMux,
    }
    fmt.Printf("Serving on port %s ...", port)
    log.Fatal(server.ListenAndServe())

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

func respondWithJSON(w http.ResponseWriter, r *http.Request, status int, payload interface{}) {
    
}

func respondWithError(w http.ResponseWriter, r *http.Request, status int, msg string) {
    w.WriteHeader(status)
    
}


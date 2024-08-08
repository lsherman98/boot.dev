package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lsherman98/boot.dev/aggregator/internal/database"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Couldnt connect to the database: %d\n", err)
		return
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := apiConfig{
		DB: dbQueries,
	}

	mux.HandleFunc("POST /v1/users", apiCfg.handlerUsersCreate)
	mux.HandleFunc("GET /v1/users", apiCfg.middlewareAuth(apiCfg.handlerUsersGet))

	mux.HandleFunc("POST /v1/feeds", apiCfg.middlewareAuth(apiCfg.handlerFeedCreate))
	mux.HandleFunc("GET /v1/feeds", apiCfg.handlerFeedsGet)

	mux.HandleFunc("GET /v1/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerFeedFollowsGet))
	mux.HandleFunc("POST /v1/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerFeedFollowCreate))
	mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerFeedFollowDelete))

	mux.HandleFunc("GET /v1/posts", apiCfg.middlewareAuth(apiCfg.handlerPostsGet))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	const collectionConcurrency = 10
	const collectionInterval = time.Minute
	go startScraping(dbQueries, collectionConcurrency, collectionInterval)

	log.Fatal(srv.ListenAndServe())
}

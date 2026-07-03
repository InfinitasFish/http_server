package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
	"database/sql"
	"internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	root := http.Dir(".")
	dbUrl := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Unable to use data source name", err)
	}
	apiCfg := &apiConfig{dbQueries: database.New(db), platform: platform}

	// ServeMux is an HTTP request multiplexer. It matches the URL of each incoming request against a list 
	// of registered patterns and calls the handler for the pattern that most closely matches the URL.
	mux := http.NewServeMux()

	// "index.html" is served from "/" by convention
	root_handler := http.StripPrefix("/app", http.FileServer(root))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(root_handler))
	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetAllUsers)
	mux.HandleFunc("POST /admin/reset_metrics", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	mux.HandleFunc("POST /api/validate_chirp", chirpValidater)

	fmt.Println("starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	fmt.Println(err)
}


package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"internal/database"
	"encoding/json"
	"time"
	"log"
	"github.com/google/uuid"
	)

type apiConfig struct {
	// atomic is thread-safe type for multiple goroutines
	fileserverHits atomic.Int32
	dbQueries *database.Queries
	platform string
}

type CreateUserBody struct {
	Email string `json:"email"`
}

type ResponseUserBody struct {
	ID string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email string `json:"email"`
}

type CreateChirpBody struct {
	Body string `json:"body"`
	UserID string `json:"user_id"`
}

type ResponseChirpBody struct {
	ID string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body string `json:"body"`
	UserID string `json:"user_id"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// injecting some code before serving original Handler
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	metricsHtml := fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
	`, cfg.fileserverHits.Load())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metricsHtml))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	newUser := CreateUserBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newUser)
	if err != nil {
		// Print + os.Exit(1)
		log.Fatal("Invalid json for creating a user", err)
	}

	userParams := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email: newUser.Email,
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Fatal("Error while creating user in db", err)
	}
	newUserResponse := ResponseUserBody{
		ID: user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email: user.Email,
	}
	
	data, _ := json.Marshal(newUserResponse)
	w.WriteHeader(201)
	// w.Write([]byte("HTTP 201 Created\n"))
	w.Write(data)
}

func (cfg *apiConfig) resetAllUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("403 Forbidden\n"))
		return
	}
	
	err := cfg.dbQueries.ResetAllUsers(r.Context())
	if err != nil {
		log.Fatal("Error while deleting users", err)
	}
	
	w.WriteHeader(200)
	// w.Write([]byte("200 Deleted all users\n"))
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	newChirp := CreateChirpBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newChirp)
	if err != nil {
		log.Fatal("Invalid json for creating a chirp", err)
	}

	// validate chirp
	filteredChirp := filterChirp(newChirp.Body)
	isValid := validateChirp(filteredChirp)
	if !isValid {
		w.WriteHeader(400)
		w.Write([]byte("Invalid chirp body"))
		return
	}
	
	// add chirp to db
	userID, err := uuid.Parse(newChirp.UserID)
	if err != nil {
		log.Fatalf("failed to parse UUID %q: %v", newChirp.UserID, err)
	}
	chirpParams := database.CreateChirpParams{
		ID: uuid.New(), 
		CreatedAt: time.Now(), 
		UpdatedAt: time.Now(), 
		Body: filteredChirp,
		UserID: userID,
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		log.Fatal("Error while adding chirp to db", err)
	}

	chirpResponse := ResponseChirpBody{
		ID: chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.String(),
		UpdatedAt: chirp.UpdatedAt.String(),
		Body: chirp.Body,
		UserID: chirp.UserID.String(),
	}
	data, _ := json.Marshal(chirpResponse)

	w.WriteHeader(201)
	w.Write([]byte(data))
}

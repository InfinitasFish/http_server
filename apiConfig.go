package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"encoding/json"
	"time"
	"log"
	"internal/database"
	"internal/auth"
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
	Password string `json:"password"`
}

type LoginUserBody struct {
	Email string `json:"email"`
	Password string `json:"password"`
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
		w.WriteHeader(404)
		log.Println("Invalid json for creating a user", err)
		return
	}

	// hash password before saving to database
	hashedPassword, err := auth.HashPassword(newUser.Password)
	if err != nil {
		w.WriteHeader(404)
		log.Println("Error hashing user's password", err)
		return
	}

	userParams := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email: newUser.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), userParams)
	if err != nil {
		w.WriteHeader(404)
		log.Println("Error while creating user in db", err)
		return
	}
	newUserResponse := ResponseUserBody{
		ID: user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email: user.Email,
	}
	
	data, _ := json.Marshal(newUserResponse)
	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	// get presumably existing user email and password
	userLoginBody := LoginUserBody{}
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&userLoginBody)
	if err != nil {
		w.WriteHeader(404)
		log.Println("Invalid json for a user to login", err)
		return
	}

	// find this user in db
	user, err := cfg.dbQueries.FindUser(r.Context(), userLoginBody.Email)
	if err != nil {
		w.WriteHeader(404)
		log.Println("User with this email isn't found", err)
		return
	}

	// check if given password compares to one in database
	match, err := auth.CheckPasswordHash(userLoginBody.Password, user.HashedPassword)
	if match != true {
		w.WriteHeader(401)
		log.Println("Wrong password for this user", err)
		return
	}

	newUserResponse := ResponseUserBody{
		ID: user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email: user.Email,
	}
	log.Println("Successful login!")

	data, _ := json.Marshal(newUserResponse)
	w.WriteHeader(200)
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
		w.WriteHeader(404)
		log.Println("Error while deleting users", err)
		return
	}
	
	w.WriteHeader(200)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	newChirp := CreateChirpBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newChirp)
	if err != nil {
		w.WriteHeader(404)
		log.Println("Invalid json for creating a chirp", err)
		return
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
		log.Printf("failed to parse UUID %q: %v", newChirp.UserID, err)
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
		w.WriteHeader(404)
		log.Println("Error while adding chirp to db", err)
		return
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

func (cfg *apiConfig) listChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetAllChrips(r.Context())
	if err != nil {
		w.WriteHeader(404)
		log.Println("Error retrieving chirps from database", err)
		return
	}
	
	var chirpsResponse []ResponseChirpBody
	for _, chirp := range chirps {
		chirpResponse := ResponseChirpBody{
			ID: chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.String(),
			UpdatedAt: chirp.UpdatedAt.String(),
			Body: chirp.Body,
			UserID: chirp.UserID.String(),
		}
		chirpsResponse = append(chirpsResponse, chirpResponse)
	}

	data, err := json.Marshal(chirpsResponse)
	if err != nil {
		w.WriteHeader(404)
		log.Println("Error encoding chirps", err)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(data))
}

func (cfg *apiConfig) getChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		w.WriteHeader(404)
		log.Println("Invalid UUID while retrieving chirp", err)
		return
	}

	fmt.Println(chirpID)
	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(404)
		log.Println("Error while retrieving chirp by ID", err)
		return
	}

	chirpResponse := ResponseChirpBody{
		ID: chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.String(),
		UpdatedAt: chirp.UpdatedAt.String(),
		Body: chirp.Body,
		UserID: chirp.UserID.String(),
	}

	data, _ := json.Marshal(chirpResponse)
	w.WriteHeader(200)
	w.Write([]byte(data))
}

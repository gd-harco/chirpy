package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/gd-harco/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func getReadyStatus(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(200)
	resp.Write([]byte("OK"))
}

func (cfg *apiConfig) getHitCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write(
		[]byte(
			fmt.Sprintf(`<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>
				`, cfg.fileserverHits.Load(),
			),
		),
	)
}

func (cfg *apiConfig) resetAPI(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write(nil)
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	cfg.db.DeleteUser(r.Context())
}

func (cfg *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type userDesc struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	user := userDesc{}
	err := decoder.Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	createdUser, err := cfg.db.CreateUser(r.Context(), user.Email)
	if err != nil {
		log.Fatal(err)
	}
	respondWithJSON(w, http.StatusCreated, createdUser)
	return
}

func (cfg *apiConfig) createChirps(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	requestPayload := database.CreateChirpsParams{}
	err := decoder.Decode(&requestPayload)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}
	toSave, err := validateChirp(requestPayload.Body)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}
	requestPayload.Body = toSave
	createdChirp, err := cfg.db.CreateChirps(r.Context(), requestPayload)
	if err != nil {
		respondWithError(w, 500, err)
		return
	}
	respondWithJSON(w, 201, createdChirp)

}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	target := r.PathValue("chirpID")

	if target == "" {
		resp, err := cfg.db.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, 500, err)
			return
		}
		respondWithJSON(w, 200, resp)
		return
	}

	id, err := uuid.Parse(target)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	resp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, 404, err)
			return
		}
		respondWithError(w, 500, err) // ou 500 selon ce que retourne ta query si non trouvé
		return
	}
	respondWithJSON(w, 200, resp)
}
func validateChirp(initial string) (string, error) {

	if len(initial) > 140 {
		return "", errors.New("chirps is too long")
	}

	cleanedInput := sanitizeInput(initial)
	return cleanedInput, nil
}

func sanitizeInput(userInput string) string {
	forbidden := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Fields(userInput)

	for i, word := range words {
		cleanedWord := strings.Trim(word, ".,!?\"';:")
		for _, target := range forbidden {
			if strings.EqualFold(cleanedWord, target) {
				words[i] = "****"
				break
			}
		}
	}
	return strings.Join(words, " ")
}

func respondWithError(w http.ResponseWriter, status int, err error) {
	respondWithJSON(w, status, struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	response, err := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to marshal JSON response"}`))
		return
	}

	w.WriteHeader(status)
	w.Write(response)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{}
	apiCfg.platform = os.Getenv("PLATFORM")
	apiCfg.db = database.New(db)
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetrics(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", getReadyStatus)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getHitCount)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetAPI)
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirps)
	mux.HandleFunc("POST /api/users", apiCfg.createUser)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirps)

	serv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	_ = serv.ListenAndServe()
}

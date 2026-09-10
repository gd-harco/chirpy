package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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

func (cfg *apiConfig) resetHitCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func validateJSON(w http.ResponseWriter, r *http.Request) {
	type chirpStruct struct {
		Content string `json:"body"`
	}
	w.Header().Add("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	chirp := chirpStruct{}
	err := decoder.Decode(&chirp)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	userInput := chirp.Content
	if len(userInput) > 140 {
		respondWithError(w, http.StatusBadRequest, errors.New("chirp is too long"))
		return
	}
	cleanedInput := sanitizeInput(userInput)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\"cleaned_body\":\"%s\"}", cleanedInput)))
}

func sanitizeInput(userInput string) string {
	forbidden := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Fields(userInput)
	for i, word := range words {
		for _, target := range forbidden {
			if strings.EqualFold(word, target) {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}

func respondWithError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	body := struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	data, _ := json.Marshal(body)
	w.Write(data)
}

func main() {
	apiCfg := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetrics(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", getReadyStatus)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getHitCount)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHitCount)
	mux.HandleFunc("POST /api/validate_chirp", validateJSON)
	serv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	_ = serv.ListenAndServe()
}

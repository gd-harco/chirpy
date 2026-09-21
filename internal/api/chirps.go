package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gd-harco/chirpy/internal/auth"
	"github.com/gd-harco/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *Config) createChirps(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err)
		return
	}
	userUUID, err := auth.JWTToUserUUID(bearer, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err)
		return
	}
	decoder := json.NewDecoder(r.Body)
	requestPayload := database.CreateChirpsParams{}
	err = decoder.Decode(&requestPayload)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	requestPayload.UserID = userUUID
	toSave, err := validateChirp(requestPayload.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	requestPayload.Body = toSave
	createdChirp, err := cfg.db.CreateChirps(r.Context(), requestPayload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, createdChirp)

}
func (cfg *Config) getChirps(w http.ResponseWriter, r *http.Request) {
	var userId uuid.NullUUID
	userIdString := r.URL.Query().Get("author_id")
	if userIdString == "" {
		userId = uuid.NullUUID{
			UUID: uuid.UUID{},
			Valid: false,
		}
	} else {
		parsedID, err := uuid.Parse(userIdString)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err)
			return
		}
		userId = uuid.NullUUID{
			UUID:  parsedID,
			Valid: true,
		}
	}
	resp, err := cfg.db.GetAllChirps(r.Context(), userId)
		if err != nil {
			respondWithError(w, 500, err)
			return
		}
	respondWithJSON(w, 200, resp)
}

func (cfg *Config) getChirp(w http.ResponseWriter, r *http.Request) {
	target := r.PathValue("chirpID")
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

func (cfg *Config) deleteChirp(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err)
		return
	}
	userUUID, err := auth.JWTToUserUUID(bearer, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if chirp.UserID != userUUID {
		respondWithError(w, http.StatusForbidden, errors.New("user is not the author of this chirp"))
		return
	}

	if err := cfg.db.DeleteChirp(r.Context(), chirpID); err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

package api

import (
	"encoding/json"
	"net/http"
)

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

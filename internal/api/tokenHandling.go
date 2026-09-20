package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gd-harco/chirpy/internal/auth"
)

func (cfg *Config) refreshJWT(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("unexpected body in request"))
		return
	}
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), bearer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, errors.New("invalid refresh token"))
			return
		}
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	if refreshToken.RevokedAt.Valid || time.Now().Compare(refreshToken.ExpiresAt) >= 0 {
		respondWithError(w, http.StatusUnauthorized, errors.New("refresh token expired or revoked"))
		return
	}
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	newToken, err := auth.MakeJWT(user.ID, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{Token: newToken})
}

/*
    Create a new POST /api/revoke endpoint. This new endpoint does not accept a request body,
	but does require a refresh token to be present in the headers, in the same Authorization: Bearer <refresh-token> format.
    Revoke the refresh token record in the database that matches the refresh token
	passed in the request header by setting the revoked_at to the current timestamp.
	Remember that any time you update a record, you should also be updating the updated_at timestamp.
    Respond with a 204 status code. A 204 status means the request was successful but no body is returned.
*/

func (cfg *Config) revokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("unexpected body in request"))
		return
	}
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), bearer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, errors.New("invalid refresh token"))
			return
		}
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	cfg.db.RevokeRefreshToken(r.Context(), refreshToken.Token)
	respondWithJSON(w, http.StatusNoContent, struct{}{})
}

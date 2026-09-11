package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gd-harco/chirpy/internal/auth"
	"github.com/gd-harco/chirpy/internal/database"
	"github.com/google/uuid"
)

type userResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func newUserResponse(u database.User) userResponse {
	return userResponse{
		Id:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
		Token:     "",
	}
}

func (cfg *Config) createUser(w http.ResponseWriter, r *http.Request) {
	type userDesc struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	user := userDesc{}
	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, 500, err)
		return
	}
	hash, err := auth.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, 500, err)
		return
	}
	createdUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: user.Email, HashedPassword: hash})
	if err != nil {
		respondWithError(w, 500, err)
		return
	}
	responseUser := newUserResponse(createdUser)
	respondWithJSON(w, http.StatusCreated, responseUser)
	return
}

func (cfg *Config) login(w http.ResponseWriter, r *http.Request) {
	type userDesc struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	decoder := json.NewDecoder(r.Body)
	user := userDesc{}
	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, 500, err)
		return
	}
	dbUser, err := cfg.db.GetUser(r.Context(), user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	valid := auth.CheckPasswordHash(user.Password, dbUser.HashedPassword)
	if !valid {
		respondWithError(w, 401, errors.New("Incorrect email or password"))
		return
	}
	responseUser := newUserResponse(dbUser)
	if user.ExpiresInSeconds == 0 || user.ExpiresInSeconds > 3600 {
		user.ExpiresInSeconds = 3600
	}
	jwt, err := auth.MakeJWT(responseUser.Id, cfg.secretKey, time.Duration(user.ExpiresInSeconds)*time.Second)
	respondWithJSON(w, 200, responseUser)
}

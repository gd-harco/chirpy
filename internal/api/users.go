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
	Id           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type userDesc struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func newUserResponse(u database.User) userResponse {
	return userResponse{
		Id:           u.ID,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		Email:        u.Email,
		Token:        "",
		RefreshToken: "",
		IsChirpyRed:  u.IsChirpyRed,
	}
}

func (cfg *Config) createUser(w http.ResponseWriter, r *http.Request) {
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
}

func (cfg *Config) updateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	userID, err := auth.JWTToUserUUID(token, cfg.secretKey)
	if err != nil{
		respondWithError(w, http.StatusUnauthorized, errors.New(""))
		return
	}
	decoder := json.NewDecoder(r.Body)
	userInfo := userDesc{}
	err = decoder.Decode(&userInfo)
	if err != nil {
		respondWithError(w, 500, err)
		return
	}
	hash, err := auth.HashPassword(userInfo.Password)
	if err != nil {
		respondWithError(w, 500, err)
		return
	}
	data := database.UpdateUserParams{
		ID: userID,
		Email: userInfo.Email,
		HashedPassword: hash,
	}
	updatedUser, err := cfg.db.UpdateUser(r.Context(), data)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, newUserResponse(updatedUser))
}

func (cfg *Config) login(w http.ResponseWriter, r *http.Request) {
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
	jwt, err := auth.MakeJWT(responseUser.Id, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	refresh := auth.MakeRefreshToken()
	stored, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{Token: refresh, UserID: responseUser.Id})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	responseUser.Token = jwt
	responseUser.RefreshToken = stored.Token
	respondWithJSON(w, 200, responseUser)
}

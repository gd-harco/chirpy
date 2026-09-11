package auth

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:   "chirpy-access",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()),
		Subject: userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	res, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Print(err)
		return "", err
	}
	return res, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Print(err)
		return uuid.Nil, err
	}
	uuidString, err := token.Claims.GetSubject()
	if err != nil {
		log.Print(err)
		return uuid.Nil, err
	}
	userUuid, err := uuid.Parse(uuidString)
	if err != nil {
		log.Print(err)
		return uuid.Nil, err
	}
	return userUuid, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	content := headers.Get("Authorization")
	if content == "" {
		return "", errors.New("No Authorization found in header")
	}
	split := strings.Split(content, " ")
	if len(split) != 2 {
		return "", errors.New("Malformed Bearer")
	}
	return split[1], nil
}

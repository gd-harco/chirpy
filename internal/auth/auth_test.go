package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	tokenString, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	gotUserID, err := ExtractUserUUIDFromJWT(tokenString, secret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	if gotUserID != userID {
		t.Fatalf("ValidateJWT() = %v, want %v", gotUserID, userID)
	}
}

func TestValidateJWTRejectsExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	tokenString, err := MakeJWT(userID, secret, -time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	_, err = ExtractUserUUIDFromJWT(tokenString, secret)
	if err == nil {
		t.Fatal("ValidateJWT() error = nil, want expired token error")
	}
}

func TestValidateJWTRejectsWrongSecret(t *testing.T) {
	userID := uuid.New()
	tokenString, err := MakeJWT(userID, "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	_, err = ExtractUserUUIDFromJWT(tokenString, "wrong-secret")
	if err == nil {
		t.Fatal("ValidateJWT() error = nil, want signature error")
	}
}

func TestNoAuthorizationHeader(t *testing.T) {
	headers := make(map[string][]string)
	httpHeaders := http.Header(headers)

	_, err := GetBearerToken(httpHeaders)
	if err == nil {
		t.Fatal("GetBearerToken() error = nil, want error for missing Authorization header")
	}
}

func TestMalformedAuthorizationHeader(t *testing.T) {
	headers := make(map[string][]string)
	httpHeaders := http.Header(headers)
	httpHeaders.Set("Authorization", "BearerTokenWithoutSpace")

	_, err := GetBearerToken(httpHeaders)
	if err == nil {
		t.Fatal("GetBearerToken() error = nil, want error for malformed Authorization header")
	}
}

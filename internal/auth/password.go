package auth

import (
	"log"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hashed, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		log.Print(err)
		return "", err
	}
	return hashed, nil
}

func CheckPasswordHash(password, hash string) bool {
	result, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		log.Print(err)
		return false
	}
	return result
}

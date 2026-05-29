package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func GeneratePasswordHash(password string) ([]byte, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return passHash, nil
}

func VerifyPassword(dbpass []byte, password string) bool {
	if err := bcrypt.CompareHashAndPassword(dbpass, []byte(password)); err != nil {
		return false
	}
	return true
}

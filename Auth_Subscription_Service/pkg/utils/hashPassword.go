package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {

	HashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(HashedPassword)
}

func CompareWithHashedPassword(hashedPassword string, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return errors.New("passwords does not match")
	}
	return nil
}

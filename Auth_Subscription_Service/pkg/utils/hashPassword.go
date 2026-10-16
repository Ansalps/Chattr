package utils

import (
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
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
		return domain.ErrPasswordMismatch
	}
	return nil
}

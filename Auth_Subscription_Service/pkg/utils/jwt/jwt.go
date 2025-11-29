package jwt

import (
	"fmt"
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils/jwt/interfacesJwt"
	"github.com/golang-jwt/jwt/v5"
)

type JwtUtil struct{}

func NewJwtUtil() interfacesJwt.Jwt {
	return &JwtUtil{}
}

func (ju *JwtUtil) GenerateToken(securityKey string, id uint64, email, role string,tokenType string,duration time.Duration) (string, error) {

	claims := &requestmodels.JwtClaims{
		ID:    id,
		Email: email,
		Role:  role,
		Type: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "Chattr",
		},
	}
	fmt.Println("jwt key", securityKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(securityKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

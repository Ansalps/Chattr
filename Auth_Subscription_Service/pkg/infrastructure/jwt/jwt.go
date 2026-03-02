package jwt

import (
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtProvider struct{}

func NewJwtProvider() *JwtProvider {
	return &JwtProvider{}
}

func (ju *JwtProvider) GenerateToken(securityKey string, id uint64, email, role string, tokenType string, duration time.Duration) (string, error) {
	jti := uuid.NewString() // secure unique token id
	claims := &requestmodels.JwtClaims{
		ID:    id,
		Email: email,
		Role:  role,
		Type:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "Chattr",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(securityKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

package requestmodels

import "github.com/dgrijalva/jwt-go"

type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required" validat:"required"`
	Password string `json:"password" binding:"required" validate:"min=6 max=20"`
}

type JwtClaims struct {
	ID    uint64
	Email string
	Role  string
	jwt.StandardClaims
}

type UserSignUpRequest struct {
    Name            string `json:"Name" binding:"required,min=3,max=30"`
    UserName        string `json:"UserName" binding:"required,min=3,max=30"`
    Email           string `json:"Email" binding:"required,email"`
    Password        string `json:"Password" binding:"required,min=3,max=30"`
    ConfirmPassword string `json:"ConfirmPassword" binding:"required,eqfield=Password"`
}
package helper

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/dgrijalva/jwt-go"
)



func GenerateToken(id uint64,email,role string)(string,error){
	config,err:=config.LoadConfig()
	if err!=nil{
		return "",errors.New("error loading config in GenerateToken Method")
	}
	claims:=&requestmodels.JwtClaims{
		ID: id,
		Email: email,
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24*time.Hour).Unix(),
			IssuedAt: time.Now().Unix(),
			Issuer: "Chattr",
		},
	}
	fmt.Println("jwt key",config.JwtKey)
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	tokenString,err:=token.SignedString([]byte(config.JwtKey))
	if err!=nil{
		return "",err
	}
	return tokenString,nil
}
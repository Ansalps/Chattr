package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func VerifyJwt(requiredRoles []string, tokenType string, tokenSecurityKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// // Get the token from the Authorization header
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Authorization header missing", nil))
			c.Abort()
			return
		}
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		if tokenSecurityKey == "" {
			c.JSON(http.StatusInternalServerError, response.ClientResponse(http.StatusInternalServerError, "Internal Server Error-missing token securtiy key", nil))
			c.Abort()
			return
		}
		fmt.Println("token security key",tokenSecurityKey)
		var jwtClaims responsemodels.JwtClaims
		token, err := jwt.ParseWithClaims(tokenString, &jwtClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecurityKey), nil
		})
		if err != nil {
			//log the error
			log.Printf("Error while parsing token : %v /n", err)
			// Check if the token is expired
			if errors.Is(err, jwt.ErrTokenExpired) {
				switch tokenType {
				case "access":
					// Expired access tokens
					switch jwtClaims.Role {
					case "user", "admin":
						c.JSON(http.StatusUnauthorized, gin.H{
							"message":                 "Access token expired",
							"is_refresh_token_needed": true,
						})
					case "otpverification":
						c.JSON(http.StatusUnauthorized, gin.H{
							"message": "Session expired. Please sign up or request a new OTP again.",
						})
					case "resetpassword":
						c.JSON(http.StatusUnauthorized, gin.H{
							"message": "Session expired. Please initiate the forgot password process again.",
						})
					default:
						c.JSON(http.StatusUnauthorized, gin.H{
							"message": "Session expired. Please login again.",
						})
					}

				case "refresh":
					// Expired refresh tokens (user or admin)
					c.JSON(http.StatusUnauthorized, gin.H{
						"message": "Session expired. Please login again to continue.",
					})
				}
				c.Abort()
				return
			} else {
				fmt.Println("is error reaching here not the error you thought print the error", err)
			}
			
			// Check for an invalid signature error
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				log.Println("Token signature is invalid.")
				c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid token signature", nil))
				c.Abort()
				return
			}
		
			// Check for an invalid claims error (e.g., unexpected claims)
			if err == jwt.ErrSignatureInvalid {
				log.Println("Invalid token signature error.")
				c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid token signature", nil))
				c.Abort()
				return
			}
		
			// Check for an unknown JWT parsing error
			log.Printf("Unexpected error: %v\n", err)
			
			// Any other JWT parsing error
			c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid token", nil))
			// c.JSON(http.StatusUnauthorized, gin.H{
			// 	"message": "Invalid token",
			// })
			c.Abort()
			return
		}
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid token", nil))
			// c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		if jwtClaims.Type != tokenType {
			c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid token type", nil))
			//c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token type"})
			c.Abort()
			return
		}
		fmt.Println("please print role",jwtClaims.Role)
		// Role check
		authorized := false
		for _, r := range requiredRoles {
			fmt.Println("please print roles inside",r)
			if jwtClaims.Role == r {
				authorized = true
				break
			}
		}
		if !authorized {
			c.JSON(http.StatusForbidden, response.ClientResponse(http.StatusForbidden, "Insufficient privileges", nil))
			//c.JSON(http.StatusForbidden, gin.H{"message": "Insufficient privileges"})
			c.Abort()
			return
		}
		fmt.Println("jwt claims",jwtClaims)
		c.Set("claims", jwtClaims)
		c.Next()
	}
}

package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Ansalps/Chattr_Api_Gateway/infrastructure/logger"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	interfacesRepository "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	RedisRepo interfacesRepository.RedisRepository
}

func NewAuthMiddlware(redisRepo interfacesRepository.RedisRepository) *AuthMiddleware {
	return &AuthMiddleware{
		RedisRepo: redisRepo,
	}
}

// func (m *AuthMiddleware) VerifyJwt(requiredRoles []string, tokenType string, tokenSecurityKey string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// // Get the token from the Authorization header
// 		tokenString := c.GetHeader("Authorization")
// 		if tokenString == "" {
// 			c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Authorization header missing", nil))
// 			c.Abort()
// 			return
// 		}
// 		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
// 		if tokenSecurityKey == "" {
// 			c.JSON(http.StatusInternalServerError, response.ClientResponse(http.StatusInternalServerError, "Internal Server Error-missing token securtiy key", nil))
// 			c.Abort()
// 			return
// 		}
// 		var jwtClaims responsemodels.JwtClaims
// 		token, err := jwt.ParseWithClaims(tokenString, &jwtClaims, func(token *jwt.Token) (interface{}, error) {
// 			return []byte(tokenSecurityKey), nil
// 		})
// 		if err != nil {
// 			//log the error
// 			log.Printf("Error while parsing token : %v /n", err)
// 			// Check if the token is expired
// 			if errors.Is(err, jwt.ErrTokenExpired) {
// 				switch tokenType {
// 				case "access":
// 					// Expired access tokens
// 					switch jwtClaims.Role {
// 					case "user", "admin":
// 						c.JSON(http.StatusUnauthorized, gin.H{
// 							"message":                 "Access token expired",
// 							"is_refresh_token_needed": true,
// 						})
// 					case "otpverification":
// 						c.JSON(http.StatusUnauthorized, gin.H{
// 							"message": "Session expired. Please sign up or request a new OTP again.",
// 						})
// 					case "resetpassword":
// 						c.JSON(http.StatusUnauthorized, gin.H{
// 							"message": "Session expired. Please initiate the forgot password process again.",
// 						})
// 					default:
// 						c.JSON(http.StatusUnauthorized, gin.H{
// 							"message": "Session expired. Please login again.",
// 						})
// 					}

// 				case "refresh":
// 					// Expired refresh tokens (user or admin)
// 					c.JSON(http.StatusUnauthorized, gin.H{
// 						"message": "Session expired. Please login again to continue.",
// 					})
// 				}
// 				c.Abort()
// 				return
// 			} else {
// 				fmt.Println("is error reaching here not the error you thought print the error", err)
// 			}

// 			// Check for an invalid signature error
// 			if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
// 				log.Println("Token signature is invalid.")
// 				c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid token signature", nil))
// 				c.Abort()
// 				return
// 			}

// 			// Check for an invalid claims error (e.g., unexpected claims)
// 			// if err == jwt.ErrSignatureInvalid {
// 			// 	log.Println("Invalid token signature error.")
// 			// 	c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid token signature", nil))
// 			// 	c.Abort()
// 			// 	return
// 			// }

// 			// Check for an unknown JWT parsing error
// 			log.Printf("Unexpected error: %v\n", err)

//			// Any other JWT parsing error
//			c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid token", nil))
//			// c.JSON(http.StatusUnauthorized, gin.H{
//			// 	"message": "Invalid token",
//			// })
//			c.Abort()
//			return
//		}
//		if !token.Valid {
//			c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid token", nil))
//			// c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
//			c.Abort()
//			return
//		}
//		if jwtClaims.Type != tokenType {
//			c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid token type", nil))
//			//c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token type"})
//			c.Abort()
//			return
//		}
//		jti := jwtClaims.RegisteredClaims.ID
//		//exp:=jwtClaims.RegisteredClaims.ExpiresAt.Time
//		if jti == "" {
//			c.JSON(http.StatusBadRequest, gin.H{"error": "token missing jti"})
//			c.Abort()
//			return
//		}
//		blacklisted, err := m.RedisRepo.IsTokenBlacklisted(jti)
//		if err != nil {
//			log.Printf("redis error: %v", err)
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
//			c.Abort()
//			return
//		}
//		if blacklisted {
//			c.JSON(http.StatusUnauthorized, gin.H{"message": "session logged out,please login again to continue"})
//			c.Abort()
//			return
//		}
//		// Role check
//		authorized := false
//		for _, r := range requiredRoles {
//			//fmt.Println("please print roles inside", r)
//			if jwtClaims.Role == r {
//				authorized = true
//				break
//			}
//		}
//		if !authorized {
//			c.JSON(http.StatusForbidden, response.ClientResponse(http.StatusForbidden, "Insufficient privileges", nil))
//			//c.JSON(http.StatusForbidden, gin.H{"message": "Insufficient privileges"})
//			c.Abort()
//			return
//		}
//			c.Set("claims", jwtClaims)
//			c.Next()
//		}
//	}
func (m *AuthMiddleware) VerifyJwt(requiredRoles []string, tokenType string, tokenSecurityKey string) gin.HandlerFunc {
	return func(c *gin.Context) {

		//log := utils.GetLogger(c)

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			//log.Warn("missing authorization header")
			response.AbortWithError(c, http.StatusUnauthorized, "Authorization header missing")
			return
		}

		if !strings.HasPrefix(tokenString, "Bearer ") {
			//log.Warn("invalid token format")
			response.AbortWithError(c, http.StatusUnauthorized, "Invalid token format")
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		var jwtClaims responsemodels.JwtClaims
		token, err := jwt.ParseWithClaims(tokenString, &jwtClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecurityKey), nil
		})

		if err != nil {
			fmt.Println("jwstClaims.Role", jwtClaims.Role)
			if errors.Is(err, jwt.ErrTokenExpired) {
				//log.Warn("token expired")
				response.AbortWithError(c, http.StatusUnauthorized, "Session expired, repeat the process to get new otp again")
				return
			}
			if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				//log.Warn("token signature invalid")
				response.AbortWithError(c, http.StatusUnauthorized, "token signature invalid")
				return
			}
			// log.Error("jwt parsing failed",
			// 	logger.Field{Key: "error", Value: err},
			// )

			response.AbortWithError(c, http.StatusUnauthorized, "Invalid token")
			return
		}

		if !token.Valid {
			//log.Warn("invalid token")
			response.AbortWithError(c, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Check token type
		if jwtClaims.Type != tokenType {
			// log.Warn("invalid token type",
			// 	logger.Field{Key: "expected_type", Value: tokenType},
			// 	logger.Field{Key: "actual_type", Value: jwtClaims.Type},
			// )

			response.AbortWithError(c, http.StatusUnauthorized, "Invalid token type")
			return
		}

		// Extract JTI
		jti := jwtClaims.RegisteredClaims.ID
		if jti == "" {
			//log.Error("missing jti in token")

			c.JSON(http.StatusBadRequest, gin.H{"error": "token missing jti"})
			c.Abort()
			return
		}

		// Check blacklist in Redis
		blacklisted, err := m.RedisRepo.IsTokenBlacklisted(jti)
		if err != nil {
			// log.Error("redis blacklist check failed",
			// 	logger.Field{Key: "error", Value: err},
			// )

			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		if blacklisted {
			// log.Warn("blacklisted token used",
			// 	logger.Field{Key: "jti", Value: jti},
			// )

			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "session logged out, please login again to continue",
			})
			c.Abort()
			return
		}

		// Role authorization
		authorized := false
		for _, r := range requiredRoles {
			fmt.Println("jwstClaims.Role", jwtClaims.Role)
			if jwtClaims.Role == r {
				authorized = true
				break
			}
		}

		if !authorized {
			// log.Warn("insufficient privileges",
			// 	logger.Field{Key: "role", Value: jwtClaims.Role},
			// 	logger.Field{Key: "required_roles", Value: requiredRoles},
			// )

			response.AbortWithError(c, http.StatusForbidden, "Insufficient privileges")
			return
		}

		// SUCCESS — attach claims to context
		c.Set("claims", jwtClaims)

		// log.Info("jwt verification successful",
		// 	logger.Field{Key: "user_id", Value: jwtClaims.ID},
		// 	logger.Field{Key: "role", Value: jwtClaims.Role},
		// )

		c.Next()
	}
}

const RequestIDKey = "request_id"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")

		if requestID == "" {
			requestID = uuid.New().String()
		}

		// store in standard context (IMPORTANT)
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

// func LoggerMiddleware(baseLogger logger.Logger) gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		start := time.Now()

// 		// 1. Get request_id (set by RequestIDMiddleware)
// 		requestID, _ := c.Get(RequestIDKey)

// 		// type assertion (safe)
// 		reqID, _ := requestID.(string)

// 		// 2. Create child logger with request_id
// 		reqLogger := baseLogger.With(
// 			logger.Field{Key: "request_id", Value: reqID},
// 		)

// 		// 3. Store logger in context
// 		utils.SetLogger(c, reqLogger)

// 		// 4. Process next middleware / handler
// 		c.Next()

// 		// 5. After request completes → log response
// 		latency := time.Since(start)
// 		status := c.Writer.Status()

// 		// Determine log level
// 		var logFunc func(string, ...logger.Field)

// 		switch {
// 		case status >= 500:
// 			logFunc = reqLogger.Error
// 		case status >= 400:
// 			logFunc = reqLogger.Warn
// 		default:
// 			logFunc = reqLogger.Info
// 		}

// 		// Prepare fields
// 		fields := []logger.Field{
// 			{Key: "method", Value: c.Request.Method},
// 			{Key: "path", Value: c.Request.URL.Path},
// 			{Key: "status", Value: status},
// 			{Key: "latency", Value: latency.String()},
// 			{Key: "client_ip", Value: c.ClientIP()},
// 		}

// 		// Add gin errors if any
// 		if len(c.Errors) > 0 {
// 			fields = append(fields, logger.Field{
// 				Key:   "errors",
// 				Value: c.Errors.String(),
// 			})
// 		}

// 		// Log message
// 		message := "request completed"
// 		if status >= 500 {
// 			message = "request failed"
// 		}

//			logFunc(message, fields...)
//		}
//	}
func LoggerMiddleware(baseLogger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		//start := time.Now()

		// 1. Get request_id (set by RequestIDMiddleware)
		requestID, _ := c.Get(RequestIDKey)

		// 2. Create child logger with request_id
		reqLogger := baseLogger.With(
			logger.Field{Key: "request_id", Value: requestID},
		)

		// 3. Store logger in context
		utils.SetLogger(c, reqLogger)

		// 4. Process next middleware / handler
		c.Next()

		// 5. After request completes → log response
		// latency := time.Since(start)

		// reqLogger.Info("request completed",
		// 	logger.Field{Key: "method", Value: c.Request.Method},
		// 	logger.Field{Key: "path", Value: c.Request.URL.Path},
		// 	logger.Field{Key: "status", Value: c.Writer.Status()},
		// 	logger.Field{Key: "latency", Value: latency.String()},
		// )
	}
}

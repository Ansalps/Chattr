package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

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

func (m *AuthMiddleware) VerifyJwt(requiredRole string, tokenType string, tokenSecurityKey string) gin.HandlerFunc {
	return func(c *gin.Context) {

		//log := utils.GetLogger(c)
		cookieName := "jwt_user_token"
		switch requiredRole {
		case "admin":
			switch tokenType {
			case "access":
				cookieName = "jwt_admin_access_token"
			case "refresh":
				cookieName = "jwt_admin_refresh_token"
			}
		case "user":
			switch tokenType {
			case "access":
				cookieName = "jwt_user_access_token"
			case "refresh":
				cookieName = "jwt_user_refresh_token"
			}
		case "otpverification":

			cookieName = "jwt_otpverification_token"
		case "resetpassword":
			cookieName = "jwt_resetpassword_token"
		}

		tokenString, err := c.Cookie(cookieName)
		if err != nil {

			tokenString, err = c.Cookie("jwt_token")
			if err != nil {

				c.JSON(http.StatusUnauthorized, gin.H{"message": "Please Log In"})
				c.Abort()
				return
			}
		}

		if tokenString == "" {
			//log.Warn("missing authorization header")

			response.AbortWithError(c, http.StatusUnauthorized, "Authorization cookie missing")
			return
		}

		var jwtClaims responsemodels.JwtClaims
		token, err := jwt.ParseWithClaims(tokenString, &jwtClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecurityKey), nil
		})

		if err != nil {
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

		if jwtClaims.Role == requiredRole {
			authorized = true

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
func getCallerInfo() (string, int) {
	// Skip 2 levels up to find the actual handler function that called the logger
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", 0
	}
	return filepath.Base(file), line
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

func LoggerMiddleware(baseLogger logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        fileName, lineNum := getCallerInfo()
        start := time.Now()

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

        // 5. After request completes → log response details
        latency := time.Since(start)
        status := c.Writer.Status()

        var logFunc func(string, ...logger.Field)
        switch {
        case status >= 500:
            logFunc = reqLogger.Error
        case status >= 400:
            logFunc = reqLogger.Warn
        default:
            logFunc = reqLogger.Info
        }

        // Prepare base fields
        fields := []logger.Field{
            {Key: "file", Value: fmt.Sprintf("%s:%d", fileName, lineNum)},
            {Key: "method", Value: c.Request.Method},
            {Key: "path", Value: c.Request.URL.Path},
            {Key: "status", Value: status},
            {Key: "latency", Value: latency.String()},
            {Key: "client_ip", Value: c.ClientIP()},
        }

        // Include Gin errors if any occurred during the request lifecycle
        if len(c.Errors) > 0 {
            fields = append(fields, logger.Field{
                Key:   "errors",
                Value: c.Errors.String(),
            })
        }

        message := "HTTP Request Completed"
        if status >= 500 {
            message = "HTTP Request Failed"
        }

        logFunc(message, fields...)
    }
}
package utils

import (
	"errors"
	"net/http"

	"github.com/Ansalps/Chattr_Chat_Service/logger"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/domain"
	"github.com/gin-gonic/gin"
)

type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

func ClientResponse(statusCode int, message string, data interface{}) Response {

	return Response{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
}

func MapError(log logger.Logger, err error, c *gin.Context) {
	switch {
	case errors.Is(err, domain.ErrInternal):
		log.Error("internal server Error",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusInternalServerError, ClientResponse(400, domain.ErrInternal.Error(), nil))
	case errors.Is(err, domain.ErrUserNotFound):
		log.Warn("User Not Found",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusNotFound, ClientResponse(404, domain.ErrUserNotFound.Error(), nil))
	case errors.Is(err, domain.ErrDatabase):
		log.Error("Database error",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusInternalServerError, ClientResponse(500, domain.ErrDatabase.Error(), nil))
	default:
		log.Error("internal server Error",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusInternalServerError, ClientResponse(400, domain.ErrInternal.Error(), nil))
	}
}

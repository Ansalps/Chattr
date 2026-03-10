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

func MapDomainError(c *gin.Context, log logger.Logger, err error) {
	var userErr *domain.NonExistingUsersError
	switch {
	case errors.Is(err,domain.ErrUserNotFound):
		log.Warn("User Not found",
			logger.Field{Key: "error",Value: err})
		c.JSON(404,ClientResponse(404,domain.ErrUserNotFound.Error(),nil))
	case errors.Is(err, domain.ErrContentTypeNil):
		log.Warn("Content type is nil",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusBadRequest, ClientResponse(400, domain.ErrContentTypeNil.Error(), nil))
	case errors.Is(err, domain.ErrUserNotInConversation):
		log.Warn("user not present in conversation",
			logger.Field{Key: "error", Value: err})
		c.JSON(403, ClientResponse(403, domain.ErrUserNotInConversation.Error(), nil))
	case errors.As(err, &userErr):
		log.Warn("Invalid userids persent",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusBadRequest, gin.H{
			"error":            "some users do not exist",
			"missing_user_ids": userErr.UserIDs,
		})
	case errors.Is(err, domain.ErrUserNotPresent):
		log.Warn("Member to remove not present in the group",
			logger.Field{Key: "error", Value: err})
		c.JSON(404, ClientResponse(404, domain.ErrUserNotPresent.Error(), nil))
	case errors.Is(err, domain.ErrGroupNotFound):
		log.Warn("Invalid group id",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusNotFound, ClientResponse(404, domain.ErrGroupNotFound.Error(), nil))
	case errors.Is(err, domain.ErrNotGroupMember):
		log.Warn("Not a Group Member",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusForbidden, ClientResponse(403, domain.ErrNotGroupMember.Error(), nil))
	case errors.Is(err, domain.ErrDatabaseTimeout):
		log.Error("Database Connection timed out",
			logger.Field{Key: "error", Value: err})
		c.JSON(500, ClientResponse(500, domain.ErrDatabaseTimeout.Error(), nil))
	case errors.Is(err, domain.ErrInternal):
		log.Error("Internal server error",
			logger.Field{Key: "error", Value: err})
		c.JSON(500, ClientResponse(500, domain.ErrInternal.Error(), nil))
	default:
		log.Error("internal server errot",
			logger.Field{Key: "error", Value: err})
		// ✅ fallback (VERY IMPORTANT)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}

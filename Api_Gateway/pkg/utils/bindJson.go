package utils

import (
	"net/http"

	"github.com/Ansalps/Chattr_Api_Gateway/infrastructure/logger"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

func BindingJson(c *gin.Context, val any, log logger.Logger) error {
	if err := c.ShouldBindJSON(&val); err != nil {
		if validationErrors := FormatValidationError(err); validationErrors != nil {
			// log.Warn("Validation error:",
			// logger.Field{Key: "error",Value: validationErrors})
			WarnValidation(log, validationErrors)
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return err
		}
		// log.Warn("Bind error:",
		// logger.Field{Key: "error",Value: err.Error()})
		WarnBind(log, err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return err
	}
	return nil
}

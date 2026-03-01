package utils

import (
	"github.com/Ansalps/Chattr_Api_Gateway/infrastructure/logger"
	"github.com/gin-gonic/gin"
)

func SetLogger(c *gin.Context, log logger.Logger) {
	c.Set("logger", log)
}

func GetLogger(c *gin.Context) logger.Logger {
	log, exists := c.Get("logger")
	if !exists {
		panic("logger not found in context")
	}
	return log.(logger.Logger)
}

func WarnValidation(log logger.Logger, validationErrors []string) {
	log.Warn("Validation error:",
		logger.Field{Key: "error", Value: validationErrors})
}
func WarnBind(log logger.Logger, err error) {
	log.Warn("Bind error:",
		logger.Field{Key: "error", Value: err.Error()})
}

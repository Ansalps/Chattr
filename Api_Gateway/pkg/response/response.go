package response

import (
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

func AbortWithError(c *gin.Context, status int, msg string) {
	c.JSON(status, ClientResponse(status, msg, nil))
	c.Abort()
}

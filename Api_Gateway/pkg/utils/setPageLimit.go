package utils

import (
	"net/http"
	"strconv"

	"github.com/Ansalps/Chattr_Api_Gateway/infrastructure/logger"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

func SetPageLimit(c *gin.Context, log logger.Logger) (int, int,int,error) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		if err != nil {
			//log.Printf("Error while string to int conversion(page), error: %v", err)
			LogAdminApi(log, 400, "Error while string to int conversion(page), error:"+err.Error())
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page value", nil))
		return 0,0,0,err
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1 || limit > 100 {
		if err != nil {
			//log.Printf("Error while string to int conversion(limit), error: %v", err)
			LogAdminApi(log, 400, "Error while string to int conversion(limit), error:"+err.Error())
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid limit value, must be between 1 and 100", nil))
		return 0,0,0,err
	}

	offset := (page - 1) * limit

	return limit, offset,page,nil
}

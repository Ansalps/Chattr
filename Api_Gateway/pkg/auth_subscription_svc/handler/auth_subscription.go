package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client/interfaces"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/utils"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthSubscriptionHandler struct {
	GPPC_Client interfaces.AuthSubscriptionClient
}

func NewAuthSubscriptionHandler(authSubscriptionClient interfaces.AuthSubscriptionClient) *AuthSubscriptionHandler {
	return &AuthSubscriptionHandler{
		GPPC_Client: authSubscriptionClient,
	}
}

func (as *AuthSubscriptionHandler) AdminLogin(c *gin.Context) {
	var adminDetails requestmodels.AdminLoginRequest
	if err := c.ShouldBindJSON(&adminDetails); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	fmt.Println("is call reaching here")
	admin, err := as.GPPC_Client.AdminLogin(adminDetails)
	fmt.Println("what about here")
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				obj = response.ClientResponse(http.StatusNotFound, "user not found", nil)
			case codes.Unauthenticated:
				obj = response.ClientResponse(http.StatusUnauthorized, "Invalide Email or Password", nil)
			default:
				obj = response.ClientResponse(http.StatusInternalServerError, "Internal Server Error", nil)
			}
		} else {
			// Unexpected non-gRPC error
			obj = response.ClientResponse(http.StatusInternalServerError, "Unexpected Error", nil)
		}
		c.JSON(obj.StatusCode, obj)
		return
	}
	success := response.ClientResponse(http.StatusOK, "Admin authenticated successfully", admin)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) UserSignUp(c *gin.Context) {
	var userSignup requestmodels.UserSignUpRequest
	if err := c.ShouldBindJSON(&userSignup); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	userResponse,err:=as.GPPC_Client.UserSignUp(userSignup)
	if err!=nil{
		
	}
}
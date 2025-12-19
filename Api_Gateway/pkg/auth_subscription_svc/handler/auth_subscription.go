package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client/interfaces"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/pb/auth_subscription"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/pb/post_relation"
	postClient "github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/client"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/utils"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthSubscriptionHandler struct {
	GPPC_Client      interfaces.AuthSubscriptionClientInterface
	config           *config.Config
	DirectClient     *client.AuthSubscriptionClient
	PostDirectClient *postClient.PostRelationClient
}

func NewAuthSubscriptionHandler(authSubscriptionClient interfaces.AuthSubscriptionClientInterface, cfg *config.Config, authSubClient *client.AuthSubscriptionClient, postDirectClient *postClient.PostRelationClient) *AuthSubscriptionHandler {
	return &AuthSubscriptionHandler{
		GPPC_Client:      authSubscriptionClient,
		config:           cfg,
		DirectClient:     authSubClient,
		PostDirectClient: postDirectClient,
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
				obj = response.ClientResponse(http.StatusUnauthorized, "Invalide Email or Password", nil)
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
	validUserName, msg1 := utils.IsValidUsername(userSignup.UserName)
	if !validUserName {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "validation failed", msg1))
		return
	}
	validPassword, msg2 := utils.IsValidPassword(userSignup.ConfirmPassword)
	if !validPassword {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "validation failed", msg2))
		return
	}
	userResponse, err := as.GPPC_Client.UserSignUp(userSignup)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				obj = response.ClientResponse(http.StatusConflict, st.Message(), nil)
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
	success := response.ClientResponse(http.StatusOK, "Otp Sent Successfully to email address provided, verify your otp within 5 minutes before getting expired", userResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) VerifyOtp(c *gin.Context) {
	var otpRequest requestmodels.OtpRequest
	if err := c.ShouldBindJSON(&otpRequest); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}

	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	fmt.Println("will call even comes here?????")
	otpRequest.Email = jwtClaims.Email
	otpRequest.UserId = jwtClaims.ID
	fmt.Println("inside verify otp handler ", jwtClaims.Email, jwtClaims.ID)
	fmt.Println("print otp request", otpRequest)
	otpResponse, err := as.GPPC_Client.VerifyOtp(otpRequest)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound, codes.InvalidArgument:
				obj = response.ClientResponse(http.StatusBadRequest, "Invalid otp", nil)
			case codes.FailedPrecondition:
				obj = response.ClientResponse(http.StatusPreconditionFailed, "Expired otp", nil)
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
	success := response.ClientResponse(http.StatusOK, "Otp verifeid successfully", otpResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) ResendOtp(c *gin.Context) {
	var resendOtpReq requestmodels.ResendOtpRequest
	if err := c.ShouldBindJSON(&resendOtpReq); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	resendOtpResponse, err := as.GPPC_Client.ResendOtp(resendOtpReq)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
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
	success := response.ClientResponse(http.StatusOK, "Otp resend Successfully to email address provided, verify your otp within 5 minutes before getting expired", resendOtpResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) AccessRegenerator(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	fmt.Printf("Claims type: %T\n", claims)
	fmt.Println(claims)
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}

	var accessRegenerator requestmodels.AccessRegeneratorRequest
	accessRegenerator.ID = jwtClaims.ID
	accessRegenerator.Email = jwtClaims.Email
	accessRegenerator.Role = jwtClaims.Role
	fmt.Println("inside handler access regeneration", jwtClaims.ID, jwtClaims.Email, jwtClaims.Role)
	accessRegeneratorResponse, err := as.GPPC_Client.AccessRegenerator(accessRegenerator)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
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
	success := response.ClientResponse(http.StatusOK, "New Access token generated", accessRegeneratorResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) ForgotPassord(c *gin.Context) {
	var forgetPasswordReq requestmodels.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&forgetPasswordReq); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	forgotPasswordRes, err := as.GPPC_Client.ForgotPassword(forgetPasswordReq)
	if err != nil {

	}
	success := response.ClientResponse(http.StatusOK, "Otp code sent successully to the email provided", forgotPasswordRes)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) ResetPassword(c *gin.Context) {
	var resetPassword requestmodels.ResetPasswordRequest
	if err := c.ShouldBindJSON(&resetPassword); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	validPassword, msg2 := utils.IsValidPassword(resetPassword.ConfirmPassword)
	if !validPassword {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "validation failed", msg2))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	resetPassword.Email = jwtClaims.Email
	resetPasswordResponse, err := as.GPPC_Client.ResetPassword(resetPassword)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				obj = response.ClientResponse(http.StatusUnauthorized, st.Message(), nil)
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
	success := response.ClientResponse(http.StatusOK, "password reset successful, please login again with new password", resetPasswordResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) BlockUser(c *gin.Context) {
	var blockUser requestmodels.BlockUserRequest
	if err := c.ShouldBindJSON(&blockUser); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	blockUserResponse, err := as.GPPC_Client.BlockUser(blockUser)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.FailedPrecondition:
				obj = response.ClientResponse(http.StatusUnauthorized, st.Message(), nil)
			case codes.NotFound:
				obj = response.ClientResponse(http.StatusUnauthorized, st.Message(), nil)
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
	success := response.ClientResponse(http.StatusOK, "Block user by user id successful ", blockUserResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) UnblockUser(c *gin.Context) {
	var unblockUser requestmodels.UnblockUserRequest
	if err := c.ShouldBindJSON(&unblockUser); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	unblockUserResponse, err := as.GPPC_Client.UnblockUser(unblockUser)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.FailedPrecondition:
				obj = response.ClientResponse(http.StatusUnauthorized, st.Message(), nil)
			case codes.NotFound:
				obj = response.ClientResponse(http.StatusUnauthorized, st.Message(), nil)
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
	success := response.ClientResponse(http.StatusOK, "Unblock user by user id successful ", unblockUserResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) UserLogin(c *gin.Context) {
	var userLogin requestmodels.UserLoginRequest
	if err := c.ShouldBindJSON(&userLogin); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	user, err := as.GPPC_Client.UserLogin(userLogin)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				obj = response.ClientResponse(http.StatusUnauthorized, "Invalid Email or Password", nil)
			case codes.Unauthenticated:
				obj = response.ClientResponse(http.StatusUnauthorized, "Invalid Email or Password", nil)
			case codes.PermissionDenied:
				obj = response.ClientResponse(http.StatusForbidden, st.Message(), nil)
			case codes.FailedPrecondition:
				obj = response.ClientResponse(http.StatusPreconditionFailed, st.Message(), nil)
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
	success := response.ClientResponse(http.StatusOK, "User authenticated successfully", user)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) GetAllUsers(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		if err != nil {
			log.Printf("Error while string to int conversion(page), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page value", nil))
		return
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1 || limit > 100 {
		if err != nil {
			log.Printf("Error while string to int conversion(limit), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid limit value, must be between 1 and 100", nil))
		return
	}

	offset := (page - 1) * limit

	var getAllUsers requestmodels.GetAllUsersRequest
	getAllUsers.Limit = uint64(limit)
	getAllUsers.Offset = uint64(offset)
	users, err := as.GPPC_Client.GetAllUsers(getAllUsers)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
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
	success := response.ClientResponse(http.StatusOK, "Get All users successully", users)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) CreateSubscriptionPlan(c *gin.Context) {
	var creatSubscriptionPlanReq requestmodels.CreateSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&creatSubscriptionPlanReq); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	createSubscriptionPlanResponse, err := as.GPPC_Client.CreateSubscriptionPlan(creatSubscriptionPlanReq)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
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
	success := response.ClientResponse(http.StatusOK, "Subscritption plan created successfully", createSubscriptionPlanResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) ActivateSubscriptionPlan(c *gin.Context) {
	var activateSubscriptionPlanReq requestmodels.ActivateSubscriptionPlanRequest
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid Subscription Plan Id", nil))
		return
	}
	activateSubscriptionPlanReq.ID = id
	activateSubscriptionPlanResponse, err := as.GPPC_Client.ActivateSubscriptionPlan(activateSubscriptionPlanReq)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.FailedPrecondition:
				obj = response.ClientResponse(http.StatusPreconditionFailed, st.Message(), nil)
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
	success := response.ClientResponse(http.StatusOK, "Subscritption plan activated successfully", activateSubscriptionPlanResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) DeactivateSubscriptionPlan(c *gin.Context) {
	var deactivateSubscriptionPlanReq requestmodels.DeactivateSubscriptionPlanRequest
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid Subscription Plan Id", nil))
		return
	}
	deactivateSubscriptionPlanReq.ID = id
	deactivateSubscriptionPlanResponse, err := as.GPPC_Client.DeactivateSubscriptionPlan(deactivateSubscriptionPlanReq)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
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
	success := response.ClientResponse(http.StatusOK, "Subscritption plan deactivated successfully", deactivateSubscriptionPlanResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) GetAllSubscriptionPlans(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		if err != nil {
			log.Printf("Error while string to int conversion(page), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page value", nil))
		return
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1 || limit > 100 {
		if err != nil {
			log.Printf("Error while string to int conversion(limit), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid limit value, must be between 1 and 100", nil))
		return
	}

	offset := (page - 1) * limit

	var getAllSubscriptionPlans requestmodels.GetAllSubscriptionPlansRequest
	getAllSubscriptionPlans.Limit = uint64(limit)
	getAllSubscriptionPlans.Offset = uint64(offset)
	subscriptionPlans, err := as.GPPC_Client.GetAllSubscriptionPlans(getAllSubscriptionPlans)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.FailedPrecondition:
				obj = response.ClientResponse(http.StatusPreconditionFailed, st.Message(), nil)
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
	success := response.ClientResponse(http.StatusOK, "Get All subscription plans successully", subscriptionPlans)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) GetAllActiveSubscriptionPlans(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		if err != nil {
			log.Printf("Error while string to int conversion(page), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page value", nil))
		return
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1 || limit > 100 {
		if err != nil {
			log.Printf("Error while string to int conversion(limit), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid limit value, must be between 1 and 100", nil))
		return
	}

	offset := (page - 1) * limit

	var getAllActiveSubscriptionPlans requestmodels.GetAllActiveSubscriptionPlansRequest
	getAllActiveSubscriptionPlans.Limit = uint64(limit)
	getAllActiveSubscriptionPlans.Offset = uint64(offset)
	subscriptionPlans, err := as.GPPC_Client.GetAllActiveSubscriptionPlans(getAllActiveSubscriptionPlans)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
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
	success := response.ClientResponse(http.StatusOK, "Get All Active subscription plans successully", subscriptionPlans)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) Subscribe(c *gin.Context) {
	var subscribeReq requestmodels.SubscribeRequest
	//planID:=c.Param("plan_id")
	PlanIdStr := c.Param("plan_id")
	planID, err := strconv.ParseUint(PlanIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid Subscription Plan Id", nil))
		return
	}
	subscribeReq.PlanId = planID
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	fmt.Println("jwt claims", jwtClaims)
	subscribeReq.UserId = jwtClaims.ID
	fmt.Println("user id", subscribeReq.UserId)
	subscribeResponse, err := as.GPPC_Client.Subscribe(subscribeReq)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
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

	fmt.Println("razorpay subscription it", subscribeResponse.RazorpaySubscriptionId)
	data := gin.H{
		"SubscriptionID": subscribeResponse.RazorpaySubscriptionId,
		"KeyID":          as.config.Razorpay.KeyId,
	}
	c.HTML(http.StatusOK, "subscription_checkout.html", data)
	// success := response.ClientResponse(http.StatusOK, "User subscribe to the plan successfully", subscribeResponse)
	// c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) VerifySubscriptionPayment(c *gin.Context) {
	var verifySubscriptionPaymentReq requestmodels.VerifySubscriptionPaymentRequest
	if err := c.ShouldBindJSON(&verifySubscriptionPaymentReq); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	// Validate signature
	if !utils.VerifyRazorpaySignature(verifySubscriptionPaymentReq.RazorpayPaymentId, verifySubscriptionPaymentReq.RazorpaySubscriptionId, verifySubscriptionPaymentReq.RazorpaySignature, as.config.Razorpay.KeySecret) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid payment signature. Request cannot be authenticated.",
		})
		return
	}
	verifySubscriptionPaymentRes, err := as.GPPC_Client.VerifySubscriptionPayment(verifySubscriptionPaymentReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server internal error",
		})
		return
	}
	c.JSON(http.StatusOK, verifySubscriptionPaymentRes)
}

func (as *AuthSubscriptionHandler) Unsubscribe(c *gin.Context) {
	var unsubscribeReq requestmodels.UnsubscribeRequest
	if err := c.ShouldBindJSON(&unsubscribeReq); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}
	subIdStr := c.Param("sub_id")
	subID, err := strconv.ParseUint(subIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid Subscription Plan Id", nil))
		return
	}
	unsubscribeReq.SubId = subID
	unsubscribeResponse, err := as.GPPC_Client.Unsubscribe(unsubscribeReq)
	if err != nil {
		var obj response.Response
		// Check if it’s a gRPC status error
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			// case codes.NotFound:
			// 	obj = response.ClientResponse(http.StatusUnauthorized, "Invalide Email or Password", nil)
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
	success := response.ClientResponse(http.StatusOK, "unsubscribed successully", unsubscribeResponse)
	c.JSON(success.StatusCode, success)
}
func (as *AuthSubscriptionHandler) SetProfileImage(c *gin.Context) {
	var setProfileImageReq requestmodels.SetProfileImageRequest
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	setProfileImageReq.UserId = jwtClaims.ID
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{"error": "Image is required"})
		return
	}
	// Check file size < 2 MB
	if file.Size > 2*1024*1024 {
		c.JSON(400, gin.H{"error": "Image must be less than 2MB"})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "Cannot open image"})
		return
	}
	defer src.Close()

	// Read first 512 bytes to detect content type
	buf := make([]byte, 512)
	_, err = src.Read(buf)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid image"})
		return
	}

	// Detect MIME type
	contentType := http.DetectContentType(buf)

	// Allowed types
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}

	if !allowed[contentType] {
		c.JSON(400, gin.H{"error": "Only JPG, PNG, or WebP images are allowed"})
		return
	}

	// Reset file pointer (since we read 512 bytes)
	src.Seek(0, 0)

	// Read full bytes
	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(500, gin.H{"error": "Cannot read image"})
		return

	}
	setProfileImageReq.Image = data
	setProfileImageReq.ContentType = contentType
	setProfileImageResponse, err := as.GPPC_Client.SetProfileImage(setProfileImageReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "error in resonse"})
		return
	}
	c.JSON(http.StatusOK, setProfileImageResponse)
}

func (as *AuthSubscriptionHandler) GetProfileInformation(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalid claims", nil))
		return
	}
	var req requestmodels.GetProfileInformationRequest
	req.UserId = jwtClaims.ID
	res, err := as.GPPC_Client.GetProfileInformation(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	fmt.Println("resp in api gateway", res)
	c.JSON(http.StatusOK, res)

}

func (as *AuthSubscriptionHandler) EditProfileInformation(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalid claims", nil))
		return
	}
	var editProfile requestmodels.EditProfile
	if err := c.ShouldBindJSON(&editProfile); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	//fmt.Println("**",*editProfile.Bio,"&&",*editProfile.Name,"!!",*editProfile.Links)
	if editProfile.Links == nil {
		fmt.Println("just checking on the firs")
	}
	if editProfile.Bio != nil && editProfile.Links == nil && editProfile.Name == nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Need any one data to update", nil))
		return
	}
	resp, err := as.DirectClient.Client.EditProfileInfromation(context.Background(), &auth_subscription.EditProfileReq{
		UserId: jwtClaims.ID,
		Name:   editProfile.Name,
		Bio:    editProfile.Bio,
		Links:  editProfile.Links,
	})
	if err != nil {
		log.Println("error from grpc calling editp profile information,error: ", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "edited profile information successfully", resp))
}
func (as *AuthSubscriptionHandler) ChangePassword(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalid claims", nil))
		return
	}
	var req requestmodels.ChangePassword
	req.UserID = jwtClaims.ID
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "validation failed", validationErrors))
			return
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "bind error", err))
		return
	}
	validPassword, msg2 := utils.IsValidPassword(req.ConfirmNewPassword)
	if !validPassword {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "validation failed", msg2))
		return
	}
	resp, err := as.DirectClient.Client.ChangePassword(context.Background(), &auth_subscription.ChangePasswordRequest{
		UserId:             req.UserID,
		OldPassword:        req.OldPassword,
		NewPasswrod:        req.NewPassword,
		ConfirmNewPassword: req.ConfirmNewPassword,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "changed password successfully", resp))
}
func (as *AuthSubscriptionHandler) SearchUser(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		if err != nil {
			log.Printf("Error while string to int conversion(page), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page value", nil))
		return
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1 || limit > 100 {
		if err != nil {
			log.Printf("Error while string to int conversion(limit), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid limit value, must be between 1 and 100", nil))
		return
	}

	offset := (page - 1) * limit

	var req requestmodels.SearchUser
	req.Limit = limit
	req.Offset = offset
	searchText := c.Query("username")
	req.SearchText = searchText
	resp, err := as.DirectClient.Client.SearchUser(context.Background(), &auth_subscription.SearchUserRequest{
		SearchText: req.SearchText,
		Limit:      int64(req.Limit),
		Offset:     int64(req.Offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "users retrieved successfully", resp))
}
func (as *AuthSubscriptionHandler) GetPublicProfile(c *gin.Context) {
	userIdStr := c.Param("user_id")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid user id", nil))
		return
	}
	var req requestmodels.GetPublicProfile
	req.UserID = userId
	//var resp *auth_subscription.GetPublicProfileResponse
	authChan := make(chan *auth_subscription.UserPublicDataResponse, 1)
	postChan := make(chan *post_relation.PostFollowCountResponse, 1)
	errChan := make(chan error, 2)
	go func() {
		authresp, err := as.DirectClient.Client.UserPublicData(context.Background(), &auth_subscription.UserPublicDataRequest{
			UserId: req.UserID,
		})
		if err != nil {
			errChan <- err
		}
		authChan <- authresp
	}()
	go func() {
		postresp, err := as.PostDirectClient.Client.PostFollowCount(context.Background(), &post_relation.PostFollowCountRequest{
			UserId: req.UserID,
		})
		if err != nil {
			errChan <- err
		}
		postChan <- postresp
	}()
	// 3. Collect results using variables
	var authData *auth_subscription.UserPublicDataResponse
	var postData *post_relation.PostFollowCountResponse
	// We need to wait for exactly 2 "events"
	for i := 0; i < 2; i++ {
		select {
		case res := <-authChan:
			authData = res
		case res := <-postChan:
			postData = res
		case <-errChan:
			// Handle error or just ignore to allow partial success
		case <-c.Done():
			c.JSON(http.StatusGatewayTimeout, "Service took too long")
			return
		}
	}

	// 1. Mandatory Check: Did we get the profile?
	if authData == nil {
		// If Auth failed, we can't show a profile at all.
		c.JSON(http.StatusNotFound, response.ClientResponse(http.StatusNotFound, "User not found", nil))
		return
	}

	// 2. Optional Data: Did we get stats?
	var followers, following, posts uint64
	if postData != nil {
		followers = postData.FollowerCount
		following = postData.FollowingCount
		posts = postData.PostCount
	} else {
		// Log that the post service is down, but don't stop the request
		log.Println("Warning: post_relation service unavailable for user", userId)
	}

	// 3. Construct the response
	c.JSON(http.StatusOK, gin.H{
		"user_info": authData,
		"social_stats": gin.H{
			"followers": followers,
			"following": following,
			"posts":     posts,
			"is_stale":  postData == nil, // Helpful for frontend to know data might be old
		},
	})
}

// func (as *AuthSubscriptionHandler) Webhook(c *gin.Context) {
// 	fmt.Println("is it reaching in webhook")
// 	body, err := io.ReadAll(c.Request.Body)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, response.ClientResponse(400, "Invalid body", nil))
// 		return
// 	}
// 	signature := c.GetHeader("X-Razorpay-Signature")
// 	fmt.Println("is it getting signature", signature)
// 	if !utils.VerifyRazorpayWebhookSignature(body, as.config.Razorpay.WebhookSecret, signature) {
// 		fmt.Println("invalid signature is the problem")
// 		c.JSON(http.StatusForbidden, response.ClientResponse(http.StatusForbidden, "invalid signature", nil))
// 	}
// 	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
// 	fmt.Println("signature is verfied")
// 	var webhookReq requestmodels.WebhookRequest
// 	fmt.Println("after signature verifcation")
// 	if err := c.ShouldBindJSON(&webhookReq); err != nil {
// 		fmt.Println("understand")
// 		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
// 			fmt.Println("What woul be", validationErrors)
// 			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
// 			return
// 		}
// 		log.Printf("Bind error: %v", err)
// 		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
// 		return
// 	}
// 	fmt.Println("hello hi verification")
// 	fmt.Println("print the webhook event", webhookReq.Event)
// 	if webhookReq.Event != "subscription.completed" {
// 		c.JSON(http.StatusPreconditionFailed, response.ClientResponse(http.StatusPreconditionFailed, "Not the expected event", nil))
// 		return
// 	}
// 	WebhookResponse, err := as.GPPC_Client.Webhook(webhookReq)
// 	if err != nil {

// 	}
// 	fmt.Println(WebhookResponse)
// 	c.JSON(http.StatusOK, WebhookResponse)
// }

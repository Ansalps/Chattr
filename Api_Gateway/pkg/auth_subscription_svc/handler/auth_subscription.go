package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client/interfaces"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/utils"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthSubscriptionHandler struct {
	GPPC_Client interfaces.AuthSubscriptionClient
	config      *config.Config
}

func NewAuthSubscriptionHandler(authSubscriptionClient interfaces.AuthSubscriptionClient, cfg *config.Config) *AuthSubscriptionHandler {
	return &AuthSubscriptionHandler{
		GPPC_Client: authSubscriptionClient,
		config:      cfg,
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
	fmt.Println("inside verify otp handler ",jwtClaims.Email,jwtClaims.ID)
	fmt.Println("print otp request",otpRequest)
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
	fmt.Println("inside handler access regeneration",jwtClaims.ID,jwtClaims.Email,jwtClaims.Role)
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
	fmt.Println("jwt claims",jwtClaims)
	subscribeReq.UserId = jwtClaims.ID
	fmt.Println("user id",subscribeReq.UserId)
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
	
	fmt.Println("razorpay subscription it",subscribeResponse.RazorpaySubscriptionId)
	data := gin.H{
		"SubscriptionID": subscribeResponse.RazorpaySubscriptionId,
		"KeyID":          as.config.Razorpay.KeyId,
	}
	c.HTML(http.StatusOK, "subscription_checkout.html", data)
	// success := response.ClientResponse(http.StatusOK, "User subscribe to the plan successfully", subscribeResponse)
	// c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler)VerifySubscriptionPayment(c *gin.Context){
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
    if !utils.VerifyRazorpaySignature(verifySubscriptionPaymentReq.RazorpayPaymentId, verifySubscriptionPaymentReq.RazorpaySubscriptionId, verifySubscriptionPaymentReq.RazorpaySignature,as.config.Razorpay.KeySecret) {
        c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid payment signature. Request cannot be authenticated.",
		})
		return
    }
	verifySubscriptionPaymentRes,err:=as.GPPC_Client.VerifySubscriptionPayment(verifySubscriptionPaymentReq)
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"error":"server internal error",
		})
		return
	}
	c.JSON(http.StatusOK,verifySubscriptionPaymentRes)
}

func (as *AuthSubscriptionHandler)Unsubscribe(c *gin.Context){
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
	unsubscribeResponse,err:=as.GPPC_Client.Unsubscribe(unsubscribeReq)
	if err!=nil{
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

// func (as *AuthSubscriptionHandler)Webhook(c *gin.Context){
// 	signature:=c.GetHeader("X-Razorpay-Signature")

// 	if !utils.VerifyRazorpayWebhookSignature(as.config.Razorpay.WebhookSecret,signature){
// 		c.JSON(http.StatusForbidden,response.ClientResponse(http.StatusForbidden,"invalid signature",nil))
// 	}
// 	var webhookReq requestmodels.WebhookRequest
// 	if err := c.ShouldBindJSON(&webhookReq); err != nil {
// 		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
// 			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
// 			return
// 		}
// 		log.Printf("Bind error: %v", err)
// 		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
// 		return
// 	}
	
	
// }
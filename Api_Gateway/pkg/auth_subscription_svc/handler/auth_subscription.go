package handler

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Ansalps/Chattr_Api_Gateway/infrastructure/logger"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client/interfaces"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	interfacesrepository "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/repository/interfacesRepository"
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
	RedisRepository  interfacesrepository.RedisRepository
}

func NewAuthSubscriptionHandler(authSubscriptionClient interfaces.AuthSubscriptionClientInterface, cfg *config.Config, authSubClient *client.AuthSubscriptionClient, postDirectClient *postClient.PostRelationClient, redisRepository interfacesrepository.RedisRepository) *AuthSubscriptionHandler {
	return &AuthSubscriptionHandler{
		GPPC_Client:      authSubscriptionClient,
		config:           cfg,
		DirectClient:     authSubClient,
		PostDirectClient: postDirectClient,
		RedisRepository:  redisRepository,
	}
}

func (as *AuthSubscriptionHandler) AdminLogin(c *gin.Context) {
	log:=utils.GetLogger(c)
	var adminDetails requestmodels.AdminLoginRequest
	err:=utils.BindingJson(c,&adminDetails,log)
	if err!=nil{
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	admin, err := as.GPPC_Client.AdminLogin(ctx,adminDetails)
	if err != nil {
		code, msg := utils.GRPCtoHTTP(err)
		// Log 4xx errors as WARN
        utils.LogPublicApiError(log,adminDetails.Email,code,msg)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Admin authenticated successfully", admin)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) UserSignUp(c *gin.Context) {
	log:=utils.GetLogger(c)
	var userSignup requestmodels.UserSignUpRequest
	
	err:=utils.BindingJson(c,&userSignup,log)
	if err!=nil{
		return
	}
	validUserName, msg1 := utils.IsValidUsername(userSignup.UserName)
	if !validUserName {
		log.Warn("Client side error on sign up",
		logger.Field{Key: "error",Value: "username validation failed"})
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "validation failed", msg1))
		return
	}
	validPassword, msg2 := utils.IsValidPassword(userSignup.ConfirmPassword)
	if !validPassword {
		log.Warn("Client side error on sign up",
		logger.Field{Key: "error",Value: "password validation failed"})
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "validation failed", msg2))
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	userResponse, err := as.GPPC_Client.UserSignUp(ctx,userSignup)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		// Log 4xx errors as WARN
        utils.LogPublicApiError(log,userSignup.Email,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Otp Sent Successfully to email address provided, verify your otp within 5 minutes before getting expired", userResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) VerifyOtp(c *gin.Context) {
	log:=utils.GetLogger(c)
	var otpRequest requestmodels.OtpRequest
	err:=utils.BindingJson(c,&otpRequest,log)
	if err!=nil{
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		utils.LogPublicApiError(log,otpRequest.Email,401,"Claims not found")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}

	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		utils.LogPublicApiError(log,otpRequest.Email,401,"Invalid claims")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	otpRequest.Email = jwtClaims.Email
	otpRequest.UserId = jwtClaims.ID
	ctx,cancel:=context.WithTimeout(c.Request.Context(),10*time.Second)
	defer cancel()
	otpResponse, err := as.GPPC_Client.VerifyOtp(ctx,otpRequest)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogPublicApiError(log,otpRequest.Email,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Otp verifeid successfully", otpResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) ResendOtp(c *gin.Context) {
	log:=utils.GetLogger(c)
	var resendOtpReq requestmodels.ResendOtpRequest
	err:=utils.BindingJson(c,&resendOtpReq,log)
	if err!=nil{
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resendOtpResponse, err := as.GPPC_Client.ResendOtp(ctx,resendOtpReq)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogPublicApiError(log,resendOtpReq.Email,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Otp resend Successfully to email address provided, verify your otp within 5 minutes before getting expired", resendOtpResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) AccessRegenerator(c *gin.Context) {
	log := utils.GetLogger(c)
	claims, exists := c.Get("claims")
	if !exists {
		utils.LogPublicApiError(log,"",401,"Claims not found")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}

	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		utils.LogPublicApiError(log,"",401,"Invlaid claims")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}

	var accessRegenerator requestmodels.AccessRegeneratorRequest
	accessRegenerator.ID = jwtClaims.ID
	accessRegenerator.Email = jwtClaims.Email
	accessRegenerator.Role = jwtClaims.Role

	accessRegeneratorResponse, err := as.GPPC_Client.AccessRegenerator(accessRegenerator)
	if err != nil {
		// log.Error("grpc access regenerator call failed",
		// 	logger.Field{Key: "error", Value: err},
		// 	logger.Field{Key: "user_id", Value: accessRegenerator.ID},
		// 	logger.Field{Key: "email", Value: accessRegenerator.Email},
		// )
		code, msg := utils.GRPCtoHTTP(err)
		utils.LogApiWithUserID(log,accessRegenerator.Email,accessRegenerator.ID,code,msg)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "New Access token generated", accessRegeneratorResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) ForgotPassword(c *gin.Context) {
	log:=utils.GetLogger(c)
	var forgetPasswordReq requestmodels.ForgotPasswordRequest
	err:=utils.BindingJson(c,&forgetPasswordReq,log)
	if err!=nil{
		return
	}
	ctx,cancel:=context.WithTimeout(c.Request.Context(),10*time.Second)
	defer cancel()
	forgotPasswordRes, err := as.GPPC_Client.ForgotPassword(ctx,forgetPasswordReq)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogPublicApiError(log,forgetPasswordReq.Email,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Otp code sent successully to the email provided", forgotPasswordRes)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) ResetPassword(c *gin.Context) {
	log:=utils.GetLogger(c)
	var resetPassword requestmodels.ResetPasswordRequest
	err:=utils.BindingJson(c,&resetPassword,log)
	if err!=nil{
		return
	}
	validPassword, msg2 := utils.IsValidPassword(resetPassword.ConfirmPassword)
	if !validPassword {
		utils.LogPublicApiError(log,resetPassword.Email,400,"vailidation failed:"+msg2)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "validation failed", msg2))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		utils.LogPublicApiError(log,resetPassword.Email,401,"Claims not found")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		utils.LogPublicApiError(log,resetPassword.Email,401,"Invalid claims")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	resetPassword.Email = jwtClaims.Email
	resetPasswordResponse, err := as.GPPC_Client.ResetPassword(resetPassword)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogPublicApiError(log,resetPassword.Email,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "password reset successful, please login again with new password", resetPasswordResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) BlockUser(c *gin.Context) {
	log:=utils.GetLogger(c)
	var blockUser requestmodels.BlockUserRequest
	err:=utils.BindingJson(c,&blockUser,log)
	if err!=nil{
		return
	}
	blockUserResponse, err := as.GPPC_Client.BlockUser(blockUser)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogAdminApi(log,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Block user by user id successful ", blockUserResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) UnblockUser(c *gin.Context) {
	log:=utils.GetLogger(c)
	var unblockUser requestmodels.UnblockUserRequest
	err:=utils.BindingJson(c,&unblockUser,log)
	if err!=nil{
		return
	}
	unblockUserResponse, err := as.GPPC_Client.UnblockUser(unblockUser)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogAdminApi(log,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Unblock user by user id successful ", unblockUserResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) UserLogin(c *gin.Context) {
	log:=utils.GetLogger(c)
	var userLogin requestmodels.UserLoginRequest
	err:=utils.BindingJson(c,&userLogin,log)
	if err!=nil{
		return
	}
	ctx,cancel:=context.WithTimeout(c.Request.Context(),10*time.Second)
	defer cancel()
	user, err := as.GPPC_Client.UserLogin(ctx,userLogin)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogPublicApiError(log,userLogin.Email,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "User authenticated successfully", user)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) GetAllUsers(c *gin.Context) {
	log:=utils.GetLogger(c)
	//pageStr := c.Query("page")
	// limitStr := c.Query("limit")

	// page, err := strconv.Atoi(pageStr)
	// if err != nil || page < 1 {
	// 	if err != nil {
	// 		//log.Printf("Error while string to int conversion(page), error: %v", err)
	// 		utils.LogAdminApi(log,400,"Error while string to int conversion(page), error:"+err.Error())
	// 	}
	// 	c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page value", nil))
	// 	return
	// }

	// limit, err := strconv.Atoi(limitStr)

	// if err != nil || limit < 1 || limit > 100 {
	// 	if err != nil {
	// 		//log.Printf("Error while string to int conversion(limit), error: %v", err)
	// 		utils.LogAdminApi(log,400,"Error while string to int conversion(limit), error:"+err.Error())
	// 	}
	// 	c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid limit value, must be between 1 and 100", nil))
	// 	return
	// }

	// offset := (page - 1) * limit
	limit,offset,page,err:=utils.SetPageLimit(c,log)
	if err!=nil{
		return
	}

	var getAllUsers requestmodels.GetAllUsersRequest
	getAllUsers.Limit = uint64(limit)
	getAllUsers.Offset = uint64(offset)
	users, err := as.GPPC_Client.GetAllUsers(getAllUsers, page)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogAdminApi(log,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Get All users successully", users)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) CreateSubscriptionPlan(c *gin.Context) {
	log:=utils.GetLogger(c)
	var creatSubscriptionPlanReq requestmodels.CreateSubscriptionPlanRequest
	err:=utils.BindingJson(c,&creatSubscriptionPlanReq,log)
	if err!=nil{
		return
	}
	createSubscriptionPlanResponse, err := as.GPPC_Client.CreateSubscriptionPlan(creatSubscriptionPlanReq)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogAdminApi(log,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Subscritption plan created successfully", createSubscriptionPlanResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) ActivateSubscriptionPlan(c *gin.Context) {
	log:=utils.GetLogger(c)
	var activateSubscriptionPlanReq requestmodels.ActivateSubscriptionPlanRequest
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.LogAdminApi(log,400,"Invalid Subscription Plan Id")
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid Subscription Plan Id", nil))
		return
	}
	activateSubscriptionPlanReq.ID = id
	activateSubscriptionPlanResponse, err := as.GPPC_Client.ActivateSubscriptionPlan(activateSubscriptionPlanReq)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogAdminApi(log,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Subscritption plan activated successfully", activateSubscriptionPlanResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) DeactivateSubscriptionPlan(c *gin.Context) {
	log:=utils.GetLogger(c)
	var deactivateSubscriptionPlanReq requestmodels.DeactivateSubscriptionPlanRequest
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.LogAdminApi(log,400,"Invalid Subscription Plan Id")
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid Subscription Plan Id", nil))
		return
	}
	deactivateSubscriptionPlanReq.ID = id
	deactivateSubscriptionPlanResponse, err := as.GPPC_Client.DeactivateSubscriptionPlan(deactivateSubscriptionPlanReq)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogAdminApi(log,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Subscritption plan deactivated successfully", deactivateSubscriptionPlanResponse)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) GetAllSubscriptionPlans(c *gin.Context) {
	// pageStr := c.Query("page")
	// limitStr := c.Query("limit")

	// page, err := strconv.Atoi(pageStr)
	// if err != nil || page < 1 {
	// 	if err != nil {
	// 		log.Printf("Error while string to int conversion(page), error: %v", err)
	// 	}
	// 	c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page value", nil))
	// 	return
	// }

	// limit, err := strconv.Atoi(limitStr)

	// if err != nil || limit < 1 || limit > 100 {
	// 	if err != nil {
	// 		log.Printf("Error while string to int conversion(limit), error: %v", err)
	// 	}
	// 	c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid limit value, must be between 1 and 100", nil))
	// 	return
	// }

	// offset := (page - 1) * limit
	log:=utils.GetLogger(c)
	limit,offset,page,err:=utils.SetPageLimit(c,log)
	if err!=nil{
		return
	}

	var getAllSubscriptionPlans requestmodels.GetAllSubscriptionPlansRequest
	getAllSubscriptionPlans.Limit = uint64(limit)
	getAllSubscriptionPlans.Offset = uint64(offset)
	subscriptionPlans, err := as.GPPC_Client.GetAllSubscriptionPlans(getAllSubscriptionPlans, page)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogAdminApi(log,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Get All subscription plans successully", subscriptionPlans)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) GetAllActiveSubscriptionPlans(c *gin.Context) {
	log:=utils.GetLogger(c) 
	limit,offset,page,err:=utils.SetPageLimit(c,log)
	if err!=nil{
		return
	}
	var getAllActiveSubscriptionPlans requestmodels.GetAllActiveSubscriptionPlansRequest
	getAllActiveSubscriptionPlans.Limit = uint64(limit)
	getAllActiveSubscriptionPlans.Offset = uint64(offset)
	subscriptionPlans, err := as.GPPC_Client.GetAllActiveSubscriptionPlans(getAllActiveSubscriptionPlans, page)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogAdminApi(log,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "Get All Active subscription plans successully", subscriptionPlans)
	c.JSON(success.StatusCode, success)
}

func (as *AuthSubscriptionHandler) Subscribe(c *gin.Context) {
	log:=utils.GetLogger(c)
	var subscribeReq requestmodels.SubscribeRequest
	//planID:=c.Param("plan_id")
	PlanIdStr := c.Param("plan_id")
	planID, err := strconv.ParseUint(PlanIdStr, 10, 64)
	if err != nil {
		utils.LogAdminApi(log,400,"Invalid Subscription Plan Id")
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid Subscription Plan Id", nil))
		return
	}
	subscribeReq.PlanId = planID
	claims, exists := c.Get("claims")
	if !exists {
		utils.LogAdminApi(log,401,"Claims not found")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		utils.LogAdminApi(log,401,"Invalid claims")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	//fmt.Println("jwt claims", jwtClaims)
	subscribeReq.UserId = jwtClaims.ID
	subscribeReq.UserEmail = jwtClaims.Email
	err=utils.BindingJson(c,&subscribeReq,log)
	if err!=nil{
		return
	}

	subscribeResponse, err := as.GPPC_Client.Subscribe(subscribeReq)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogApiWithUserID(log,subscribeReq.UserEmail,subscribeReq.UserId,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}

	//fmt.Println("razorpay subscription id", subscribeResponse.RazorpaySubscriptionId)
	data := gin.H{
		"SubscriptionID": subscribeResponse.RazorpaySubscriptionId,
		"KeyID":          as.config.Razorpay.KeyId,
	}
	c.HTML(http.StatusOK, "razorpaydoc.html", data)
	success := response.ClientResponse(http.StatusOK, "User subscribe to the plan successfully", subscribeResponse)
	c.JSON(success.StatusCode, success)
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
	log:=utils.GetLogger(c)
	claims, exists := c.Get("claims")
	if !exists {
		utils.LogAdminApi(log,401,"Calims not found")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		utils.LogAdminApi(log,401,"Invalid claims")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	var unsubscribeReq requestmodels.UnsubscribeRequest
	unsubscribeReq.UserID = jwtClaims.ID
	err:=utils.BindingJson(c,&unsubscribeReq,log)
	if err!=nil{
		return
	}
	// subIdStr := c.Param("sub_id")
	// subID, err := strconv.ParseUint(subIdStr, 10, 64)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid Subscription Plan Id", nil))
	// 	return
	// }
	//unsubscribeReq.SubId = subID

	unsubscribeResponse, err := as.GPPC_Client.Unsubscribe(unsubscribeReq)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogApiWithUserID(log,"",unsubscribeReq.UserID,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	success := response.ClientResponse(http.StatusOK, "unsubscribed successully", unsubscribeResponse)
	c.JSON(success.StatusCode, success)
}
func (as *AuthSubscriptionHandler) SetProfileImage(c *gin.Context) {
	log:=utils.GetLogger(c)
	var setProfileImageReq requestmodels.SetProfileImageRequest
	claims, exists := c.Get("claims")
	if !exists {
		utils.LogAdminApi(log,401,"Claims not found")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		utils.LogAdminApi(log,401,"Invalid claims")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	setProfileImageReq.UserId = jwtClaims.ID
	file, err := c.FormFile("image")
	if err != nil {
		utils.LogApiWithUserID(log,"",setProfileImageReq.UserId,400,"Image is required"+err.Error())
		c.JSON(400, gin.H{"error": "Image is required"})
		return
	}
	str := as.config.ProfileImgSize
	num, err := strconv.Atoi(str) // returns (int, error)
	if err != nil {
		//fmt.Println("Error:", err)
		utils.LogApiWithUserID(log,"",setProfileImageReq.UserId,500,"error in converting prorile image size from string to int"+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error in converting prorile image size from string to int"})
		return
	}
	// Check file size < 2 MB
	if file.Size > int64(num)*1024*1024 {
		utils.LogApiWithUserID(log,"",setProfileImageReq.UserId,500,"Image must be less than 2MB")
		c.JSON(400, gin.H{"error": "Image must be less than 2MB"})
		return
	}
	src, err := file.Open()
	if err != nil {
		utils.LogApiWithUserID(log,"",setProfileImageReq.UserId,500,"Cannot open image"+err.Error())
		c.JSON(500, gin.H{"error": "Cannot open image"})
		return
	}
	defer src.Close()

	// Read first 512 bytes to detect content type
	buf := make([]byte, 512)
	_, err = src.Read(buf)
	if err != nil {
		//log.Println(err)
		utils.LogApiWithUserID(log,"",setProfileImageReq.UserId,400,"Invalid image"+err.Error())
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
		utils.LogApiWithUserID(log,"",setProfileImageReq.UserId,400,"Only JPG, PNG, or WebP images are allowed")
		c.JSON(400, gin.H{"error": "Only JPG, PNG, or WebP images are allowed"})
		return
	}

	// Reset file pointer (since we read 512 bytes)
	src.Seek(0, 0)

	// Read full bytes
	data, err := io.ReadAll(src)
	if err != nil {
		//log.Println(err)
		utils.LogApiWithUserID(log,"",setProfileImageReq.UserId,500,"Cannot read image"+err.Error())
		c.JSON(500, gin.H{"error": "Cannot read image"})
		return

	}
	setProfileImageReq.Image = data
	setProfileImageReq.ContentType = contentType
	setProfileImageResponse, err := as.GPPC_Client.SetProfileImage(setProfileImageReq)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogApiWithUserID(log,"",setProfileImageReq.UserId,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}
	c.JSON(http.StatusOK, setProfileImageResponse)
}

func (as *AuthSubscriptionHandler) GetProfileInformation(c *gin.Context) {
	log:=utils.GetLogger(c)
	claims, exists := c.Get("claims")
	if !exists {
		utils.LogAdminApi(log,401,"Claims not found")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		utils.LogAdminApi(log,401,"invalid claims")
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalid claims", nil))
		return
	}
	var req requestmodels.GetProfileInformationRequest
	req.UserId = jwtClaims.ID
	res, err := as.GPPC_Client.GetProfileInformation(req)
	if err != nil {
		code,msg:=utils.GRPCtoHTTP(err)
		utils.LogApiWithUserID(log,res.Email,req.UserId,code,msg)
		c.JSON(code,response.ClientResponse(code,msg,nil))
		return
	}

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
	// if editProfile.Links == nil {
	// 	fmt.Println("just checking on the firs")
	// }
	if editProfile.Bio == nil && editProfile.Links == nil && editProfile.Name == nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Need any one data to update", nil))
		return
	}
	resp, err := as.DirectClient.Client.EditProfileInfromation(context.Background(), &auth_subscription.EditProfileReq{
		UserId: jwtClaims.ID,
		Name:   editProfile.Name,
		Bio:    editProfile.Bio,
		Links:  editProfile.Links,
		Phone:  editProfile.Phone,
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
	// var resp *auth_subscription.SearchUserResponse
	// var r  []*auth_subscription.UserMetaData
	resp, err := as.DirectClient.Client.SearchUser(context.Background(), &auth_subscription.SearchUserRequest{
		SearchText: req.SearchText,
		Limit:      int64(req.Limit),
		Offset:     int64(req.Offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	resp2 := make([]responsemodels.UserMetaData, len(resp.UserMetaData))
	//var resp1 responsemodels.SearchUserResponse
	for i, v := range resp.UserMetaData {
		resp2[i].UserId = v.UserId
		resp2[i].UserName = v.UserName
		resp2[i].Name = v.Name
		resp2[i].ProfileImgUrl = v.ProfileImgUrl
	}
	resp1 := responsemodels.SearchUserResponse{
		UserMeataData: resp2,
		Pagingation: responsemodels.PaginationDetails{
			CurrentPage: page,
			PageSize:    limit,
		},
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "users retrieved successfully", resp1))
}
func (as *AuthSubscriptionHandler) GetPublicProfile(c *gin.Context) {
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
	userIdStr := c.Param("user_id")
	if userIdStr == "" {
		str := strconv.FormatUint(jwtClaims.ID, 10)
		userIdStr = str
	}
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
	//fmt.Println("user id print here please ", req.UserID)
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
			//fmt.Println("res00000",res)
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
	//fmt.Println("auth datat", authData)
	// if authData.BlueTick==nil{

	// }
	// 1. Mandatory Check: Did we get the profile?
	if authData == nil {
		// If Auth failed, we can't show a profile at all.
		c.JSON(http.StatusNotFound, response.ClientResponse(http.StatusNotFound, "User not found", nil))
		return
	}
	user_info_resp := responsemodels.AuthData{
		UserId:        authData.UserId,
		UserName:      authData.UserName,
		Name:          authData.Name,
		ProfileImgUrl: authData.ProfileImgUrl,
		BlueTick:      authData.BlueTick,
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
		"user_info": user_info_resp,
		"social_stats": gin.H{
			"followers": followers,
			"following": following,
			"posts":     posts,
			"is_stale":  postData == nil, // Helpful for frontend to know data might be old
		},
	})
}

func (as *AuthSubscriptionHandler) Logout(c *gin.Context) {
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
	// Extract jti and exp
	jti := jwtClaims.RegisteredClaims.ID
	exp := jwtClaims.RegisteredClaims.ExpiresAt.Time
	if jti == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token missing jti"})
		return
	}
	err := as.RedisRepository.BlacklistToken(jti, exp)
	if err != nil {
		// Decide your policy here
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
		return
	}

	// 5️⃣ Success
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

func (as *AuthSubscriptionHandler) Webhook(c *gin.Context) {
	// 1. Read the raw body for signature verification
	// body, err := io.ReadAll(c.Request.Body)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, response.ClientResponse(400, "Invalid body", nil))
	// 	return
	// }

	// 2. Verify Signature
	// signature := c.GetHeader("X-Razorpay-Signature")
	// if signature!="postman-bypass"{
	// 	if !utils.VerifyRazorpayWebhookSignature(body, as.config.Razorpay.WebhookSecret, signature) {
	// 		fmt.Println("Security Alert: Invalid Webhook Signature")
	// 		c.JSON(http.StatusForbidden, response.ClientResponse(403, "invalid signature", nil))
	// 		return // IMPORTANT: Don't forget to return here!
	// 	}
	// }

	// 3. Unmarshal the body we already read
	var webhookReq requestmodels.RazorpayEvent
	// if err := json.Unmarshal(body, &webhookReq); err != nil {
	// 	log.Printf("Unmarshal error: %v", err)
	// 	c.JSON(http.StatusBadRequest, response.ClientResponse(400, "Invalid request JSON", nil))
	// 	return
	// }
	if err := c.ShouldBindJSON(&webhookReq); err != nil {
		log.Println("error in binding", err)
		return
	}
	//fmt.Printf("Type: %T, Value: %v\n", webhookReq.Payload.Subscription.Entity.Notes["user_id"], webhookReq.Payload.Subscription.Entity.Notes["user_id"])
	UserIdStr := webhookReq.Payload.Subscription.Entity.Notes["user_id"]
	UserID, err := strconv.ParseUint(UserIdStr, 10, 64)
	if err != nil {
		//fmt.Println("Error converting string to uint64:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error in conver string user id to uint64"})
		return
	}
	//log.Println("printing webhook event", webhookReq.Event)
	//log.Println("webhook req", webhookReq)
	// 4. Validate Event Type
	var res *auth_subscription.WebhookSubscriptionActivatedResponse
	switch webhookReq.Event {
	case "subscription.activated":
		//fmt.Println("is it coming here man?")
		res, err = as.DirectClient.Client.WebhookSubscriptionActivated(context.Background(), &auth_subscription.WebhookSubscriptionActivatedRequest{
			RazorpaySubscriptionId: webhookReq.Payload.Subscription.Entity.ID,
			Status:                 webhookReq.Payload.Subscription.Entity.Status,
			PaidCount:              int64(webhookReq.Payload.Subscription.Entity.PaidCount),
			RemainingCount:         int64(webhookReq.Payload.Subscription.Entity.RemainingCount),
			StartAt:                utils.UnixToProto(webhookReq.Payload.Subscription.Entity.StartAt),
			EndAt:                  utils.UnixToProto(webhookReq.Payload.Subscription.Entity.EndAt),
			UserId:                 UserID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	case "subscription.charged":
		//fmt.Printf("The type of UserID is: %T\n", UserID)
		_, err := as.DirectClient.Client.WebhookSubscriptionCharged(context.Background(), &auth_subscription.WebhookSubscriptionChargedRequest{
			RazorpaySubscriptionId: webhookReq.Payload.Subscription.Entity.ID,
			RazorpayPlanId:         webhookReq.Payload.Subscription.Entity.PlanID,
			NextChargeAt:           utils.UnixToProto(webhookReq.Payload.Subscription.Entity.CurrentEnd),
			InvoiceId:              webhookReq.Payload.Payment.Entity.InvoiceID,
			Amount:                 webhookReq.Payload.Payment.Entity.Amount,
			Currency:               webhookReq.Payload.Payment.Entity.Currency,
			Method:                 webhookReq.Payload.Payment.Entity.Method,
			Status:                 webhookReq.Payload.Payment.Entity.Status,
			PaidCount:              int64(webhookReq.Payload.Subscription.Entity.PaidCount),
			RemainingCount:         int64(webhookReq.Payload.Subscription.Entity.RemainingCount),
			TransactionDate:        utils.UnixToProto(webhookReq.CreatedAt),
			PaymentId:              webhookReq.Payload.Payment.Entity.ID,
			UserId:                 UserID,
		})
		//fmt.Println("res", res)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	case "subscription.halted":
		//fmt.Println("user id ",UserID)
		//fmt.Printf("The type of UserID is: %T\n", UserID)
		_, err := as.DirectClient.Client.WebhookSubscriptionHalted(context.Background(), &auth_subscription.WebhookSubscriptionHaltedRequest{
			RazorpaySubscriptionId: webhookReq.Payload.Subscription.Entity.ID,
			Status:                 webhookReq.Payload.Subscription.Entity.Status,
			UserId:                 UserID,
		})
		//fmt.Println("resp", resp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	case "subscription.cancelled":
		_, err := as.DirectClient.Client.WebhookSubscriptionCancelled(context.Background(), &auth_subscription.WebhookSubscriptionCancelledRequest{
			RazorpaySubscriptionId: webhookReq.Payload.Subscription.Entity.ID,
			Status:                 webhookReq.Payload.Subscription.Entity.Status,
			CancelledAt:            utils.UnixToProto(webhookReq.Payload.Subscription.Entity.EndedAt),
			UserId:                 UserID,
		})
		//fmt.Println("resp", resp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	case "subscription.completed":
		//fmt.Println("is it here in completed")
		_, err := as.DirectClient.Client.WebhookSubscriptionCompleted(context.Background(), &auth_subscription.WebhookSubscriptionCompletedRequest{
			RazorpaySubscriptionId: webhookReq.Payload.Subscription.Entity.ID,
			Status:                 webhookReq.Payload.Subscription.Entity.Status,
			UserId:                 UserID,
		})
		//fmt.Println("resp", resp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	default:
		c.JSON(http.StatusOK, "ignored event")
	}

	// if webhookReq.Event != "subscription.completed" {
	// 	c.JSON(http.StatusOK, response.ClientResponse(200, "Event ignored", nil)) // Better to return 200 for ignored events
	// 	return
	// }

	// 5. Logic execution (gRPC or Usecase)
	// WebhookResponse, err := as.GPPC_Client.WebhookSubsciptionCompleted(webhookReq)
	// if err != nil {
	// 	log.Printf("Internal processing error: %v", err)
	// 	c.JSON(http.StatusInternalServerError, response.ClientResponse(500, "Internal error", nil))
	// 	return
	// }

	c.JSON(http.StatusOK, res)
}

func (as *AuthSubscriptionHandler) GetSubscriptionDetails(c *gin.Context) {
	// SubIdStr:=c.Param("sub_id")
	// SubID,err:=strconv.ParseUint(SubIdStr,10,64)
	// if err!=nil{
	// 	log.Println("error converting string to uint sub id")
	// 	c.JSON(500,gin.H{"error":"internal error in conversion of string to uint"})
	// 	return
	// }
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
	var req requestmodels.GetSubscriptionDetails
	req.UserID = jwtClaims.ID
	//req.SubId=SubID
	resp, err := as.DirectClient.Client.GetSubscriptionDetails(context.Background(), &auth_subscription.GetSubscriptionDetailsRequest{
		UserId: req.UserID,
		//SubId: req.SubId,
	})
	if err != nil {
		log.Println(err)
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}
	resp2 := responsemodels.SubscriptionPlan{
		ID:             resp.SubsciptionPlan.Id,
		CreatedAt:      resp.SubsciptionPlan.CreatedAt.AsTime(),
		UpdatedAt:      resp.SubsciptionPlan.UpdatedAt.AsTime(),
		RazorpayPlanId: resp.SubsciptionPlan.RazorpayPlanId,
		Name:           resp.SubsciptionPlan.Name,
		Price:          resp.SubsciptionPlan.Price / 100,
		Currency:       resp.SubsciptionPlan.Currency,
		Period:         resp.SubsciptionPlan.Period,
		Interval:       resp.SubsciptionPlan.Interval,
		Description:    resp.SubsciptionPlan.Description,
		IsActive:       resp.SubsciptionPlan.IsActive,
	}
	var cancelledAt, startAt, endAt, nextChargeAt *time.Time
	if resp.CancelledAt != nil {
		t := resp.CancelledAt.AsTime()
		cancelledAt = &t
	}
	if resp.StartAt != nil {
		t := resp.StartAt.AsTime()
		startAt = &t
	}
	if resp.EndAt != nil {
		t := resp.EndAt.AsTime()
		endAt = &t
	}
	if resp.NextChargeAt != nil {
		t := resp.NextChargeAt.AsTime()
		nextChargeAt = &t
	}
	//fmt.Println("cancelld at",resp.CancelledAt.AsTime())
	resp1 := responsemodels.GetSubscriptionDetails{
		SubscriptionPlan:       resp2,
		ID:                     resp.SubscriptionId,
		CreatedAt:              resp.CreatedAt.AsTime(),
		UpdatedAt:              resp.UpdatedAt.AsTime(),
		UserID:                 resp.UserId,
		RazorpaySubscriptionId: resp.RazorpaySubscriptionId,
		Status:                 resp.Status,
		ShortUrl:               resp.ShortUrl,
		StartAt:                startAt,
		EndAt:                  endAt,
		NextChargeAt:           nextChargeAt,
		TotalCount:             int(resp.TotalCount),
		RemainingCount:         int(resp.RemainingCount),
		PaidCount:              int(resp.PaidCount),
		CancelledAt:            cancelledAt,
		CancelReason:           resp.CancelReason,
	}
	c.JSON(http.StatusOK, resp1)
}

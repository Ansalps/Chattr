package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/responsemodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/usecase/interfacesUsecase"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils/jwt/interfacesJwt"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils/randomNumber/interfacesRandomNumber"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils/smtp/interfacesSmtp"
	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)

type AuthSubscriptionUsecase struct {
	SmtpUtil                   interfacesSmtp.Smtp
	AuthSubscriptionRepository interfacesRepository.AuthSubscriptionRepository
	RandomUtil                 interfacesRandomNumber.RandomNumber
	TokenSecurityKey		*config.Token
	JwtUtil	interfacesJwt.Jwt
	//RazorpayCredentials	*config.Razorpay
	RazorpayClient *razorpay.Client
}

func NewAuthSubscriptionUsecase(repository interfacesRepository.AuthSubscriptionRepository,randomUtil interfacesRandomNumber.RandomNumber, 
	smtpUtil interfacesSmtp.Smtp,tokenSecurityKey *config.Token,jwtUtil interfacesJwt.Jwt,/*razorpayCredentials *config.Razorpay,*/razorpayClient *razorpay.Client) interfacesUsecase.AuthSubscriptionUsecase {
	return &AuthSubscriptionUsecase{
		AuthSubscriptionRepository: repository,
		SmtpUtil:                   smtpUtil,
		RandomUtil:randomUtil,
		TokenSecurityKey: tokenSecurityKey,
		JwtUtil: jwtUtil,
		//RazorpayCredentials: razorpayCredentials,
		RazorpayClient: razorpayClient,
	}
}

var (
	ErrInvalidCredentials          = errors.New("invalid credentials")
	ErrUserNotFound                = errors.New("user not found")
	ErrUserAlreadyExistsByEmail    = errors.New("user already exists, try again with another email")
	ErrUserAlreadyExistsByUsername = errors.New("username already taken, try with another username")
	ErrOtpExpired=errors.New("otp expired")
	ErrUserNotActive=errors.New("Cannot block user, email not verified or user alreday blocked")
	ErrUserNotBlocked=errors.New("Cannnot unblock user, unblock allowed for users who are alreday in blocked state")
	ErrBlockedLogin=errors.New("User account blocked, cannot login")
	ErrPendingLogin=errors.New("Otp Verfication Pending, verfiy otp to login")
	ErrSubscriptionPlanAlreadyActive=errors.New("Cannot the activate the subscription plan, subscription plan is already active")
	ErrSubscriptionPlanAlreadyDeactive=errors.New("Cannot the deactivate the subscription plan, subscription plan is already deactive")
	//ErrRazropayApi=errors.New("error calling razorpay api")
)

func (as *AuthSubscriptionUsecase) AdminLogin(admin requestmodels.AdminLoginRequest) (responsemodels.AdminLoginResponse, error) {
	admins, err := as.AuthSubscriptionRepository.CheckAdminExistsByEmail(admin.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.AdminLoginResponse{}, ErrUserNotFound
		}
		return responsemodels.AdminLoginResponse{}, fmt.Errorf("database error: %w", err)
	}
	if admin.Password != admins.Password {
		return responsemodels.AdminLoginResponse{}, ErrInvalidCredentials
	}
	adminAccessTokenString, err := as.JwtUtil.GenerateToken(as.TokenSecurityKey.AdminSecurityKey,uint64(admins.ID), admins.Email, "admin","access",24*time.Hour)
	if err != nil {
		return responsemodels.AdminLoginResponse{}, fmt.Errorf("Failed to generarate access token for admin: %w", err)
	}
	adminRefreshTokenString,err:=as.JwtUtil.GenerateToken(as.TokenSecurityKey.AdminRefreshKey,uint64(admins.ID),admins.Email,"admin","refresh",24*7*time.Hour)
	if err != nil {
		return responsemodels.AdminLoginResponse{}, fmt.Errorf("Failed to generarate refresh token for admin: %w", err)
	}
	fmt.Println("sdflksljflsj")
	fmt.Println("hi hello",as.TokenSecurityKey.AdminSecurityKey,as.TokenSecurityKey.AdminRefreshKey)
	return responsemodels.AdminLoginResponse{
		Admin: responsemodels.AdminDetails{
			ID:    admins.ID,
			Email: admins.Email,
		},
		AccessToken: adminAccessTokenString,
		RefreshToken: adminRefreshTokenString,
	}, nil
}

func (as *AuthSubscriptionUsecase)BlockUser(blockUserReq requestmodels.BlockUserRequest)(responsemodels.BlockUserResponse,error){
	if blockUserReq.UserId==0{
		return responsemodels.BlockUserResponse{},ErrUserNotFound
	}
	status,err:=as.AuthSubscriptionRepository.CheckUserStatus(blockUserReq.UserId)
	if err!=nil{
		return responsemodels.BlockUserResponse{},fmt.Errorf("database error: %w", err)
	}
	if status!="active"{
		return responsemodels.BlockUserResponse{},ErrUserNotActive
	}
	err=as.AuthSubscriptionRepository.ChangeUserStatusToBlockedByUserId(blockUserReq)
	if err!=nil{
		return responsemodels.BlockUserResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.BlockUserResponse{
		UserId: blockUserReq.UserId,
	},nil
}

func (as *AuthSubscriptionUsecase) UnblockUser(unblockUserReq requestmodels.UnblockUserRequest)(responsemodels.UnblockUserResponse,error){
	if unblockUserReq.UserId==0{
		return responsemodels.UnblockUserResponse{},ErrUserNotFound
	}
	status,err:=as.AuthSubscriptionRepository.CheckUserStatus(unblockUserReq.UserId)
	if err!=nil{
		return responsemodels.UnblockUserResponse{},fmt.Errorf("database error: %w", err)
	}
	if status!="blocked"{
		return responsemodels.UnblockUserResponse{},ErrUserNotBlocked
	}
	err=as.AuthSubscriptionRepository.ChangeUserStatusToActiveByUserId(unblockUserReq)
	if err!=nil{
		return responsemodels.UnblockUserResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.UnblockUserResponse{
		UserId: unblockUserReq.UserId,
	},nil
}

func (as *AuthSubscriptionUsecase)GetAllUsers(getAllUsersReq requestmodels.GetAllUsersRequest)(responsemodels.GetAllUsersResponse,error){
	users,err:=as.AuthSubscriptionRepository.GetAllUsers(getAllUsersReq)
	if err!=nil{
		return responsemodels.GetAllUsersResponse{},fmt.Errorf("database error: %w", err)
	}
	return responsemodels.GetAllUsersResponse{
		Users: users.Users,
	},nil
}

func (as *AuthSubscriptionUsecase) UserSignUp(userReq requestmodels.UserSignUpRequest) (responsemodels.UserSignupResponse, error) {
	err:=as.AuthSubscriptionRepository.DeletePendingUser(userReq.Email)
	if err!=nil{
		return responsemodels.UserSignupResponse{}, fmt.Errorf("database error: %w", err)
	}
	user, err := as.AuthSubscriptionRepository.CheckUserExistsByEmail(userReq.Email)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return responsemodels.UserSignupResponse{}, fmt.Errorf("database error: %w", err)
		}
	}
	if user != nil {
		return responsemodels.UserSignupResponse{
			ID:       user.ID,
			UserName: user.UserName,
			Name:     user.Name,
			Email:    user.Email,
		}, ErrUserAlreadyExistsByEmail
	}
	usernameAlredayExists, err := as.AuthSubscriptionRepository.CheckUserExistsByUseraname(userReq.UserName)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return responsemodels.UserSignupResponse{}, fmt.Errorf("database error: %w", err)
		}
	}
	if usernameAlredayExists != nil {
		return responsemodels.UserSignupResponse{
			ID:       usernameAlredayExists.ID,
			UserName: usernameAlredayExists.UserName,
			Name:     usernameAlredayExists.Name,
			Email:    usernameAlredayExists.Email,
		}, ErrUserAlreadyExistsByUsername
	}
	otp := as.RandomUtil.RandomNumber()
	fmt.Println("Otp is ----",otp)
	err=as.AuthSubscriptionRepository.DeleteOtpByEmail(userReq.Email)
	if err!=nil{
		return responsemodels.UserSignupResponse{}, fmt.Errorf("database error: %w", err)
	}
	
	expiration := time.Now().Add(5 * time.Minute)
	err = as.AuthSubscriptionRepository.TemporarySavingUserOtp(otp, userReq.Email, expiration)
	if err != nil {
		fmt.Println("cannont save otp in db")
		return responsemodels.UserSignupResponse{}, fmt.Errorf("database error: %w", err)
	}
	err = as.SmtpUtil.SendVerifcationEmailWithOtp(otp, userReq.Email, userReq.Name)
	if err != nil {
		return responsemodels.UserSignupResponse{}, fmt.Errorf("Error in sending otp to email address: %w", err)
	}
	hashedPassword := utils.HashPassword(userReq.ConfirmPassword)
	userReq.Password = hashedPassword

	userRes,err := as.AuthSubscriptionRepository.CreateUser(&userReq)
	if err != nil {
		return responsemodels.UserSignupResponse{}, fmt.Errorf("database error: %w", err)
	}
	fmt.Println("userRes.ID is ",userRes.ID)
	otpVerificationToken, err := as.JwtUtil.GenerateToken(as.TokenSecurityKey.OtpVerificationSecurityKey,uint64(userRes.ID),userRes.Email,"otpverification","access",5*time.Minute)
	if err != nil {
		return responsemodels.UserSignupResponse{}, fmt.Errorf("Failed to generarate token for otp verfication: %w", err)
	}
	return responsemodels.UserSignupResponse{
		ID: userRes.ID,
		UserName: userRes.UserName,
		Name: userRes.Name,
		Email: userRes.Email,
		OtpVerificationToken: otpVerificationToken,
	},nil
}

func (as *AuthSubscriptionUsecase)VerifyOtp(otpReq requestmodels.OtpRequest)(responsemodels.OtpVerificationResponse,error){
	otp,err:=as.AuthSubscriptionRepository.CheckOtpExistsByEmail(otpReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.OtpVerificationResponse{}, ErrUserNotFound
		}
		return responsemodels.OtpVerificationResponse{}, fmt.Errorf("database error: %w", err)
	}
	if otp.OTP!=otpReq.OtpCode{
		return responsemodels.OtpVerificationResponse{},ErrInvalidCredentials
	}
	if time.Now().After(otp.Expiration){
		return responsemodels.OtpVerificationResponse{},ErrOtpExpired
	}
	err=as.AuthSubscriptionRepository.ChangeOtpStatus(otpReq.Email)
	if err!=nil{
		return responsemodels.OtpVerificationResponse{},fmt.Errorf("database error: %w", err)
	}
	err=as.AuthSubscriptionRepository.ChangeUserStatusByEmail(otpReq.Email)
	if err!=nil{
		return responsemodels.OtpVerificationResponse{},fmt.Errorf("database error: %w", err)
	}
	if otpReq.Purpose=="user-forgot-password"{
		resetPasswordToken, err := as.JwtUtil.GenerateToken(as.TokenSecurityKey.ResetPasswordSecurityKey,uint64(otpReq.UserId),otp.Email,"resetpassword","access",5*time.Minute)
		if err != nil {
			return responsemodels.OtpVerificationResponse{}, fmt.Errorf("Failed to generarate token for otp verfication: %w", err)
		}
		return responsemodels.OtpVerificationResponse{
			Email: otp.Email,
			Status: "verified",
			TempToken: resetPasswordToken,
		},nil
	}
	fmt.Println("in verify otp usercase --",otpReq.UserId)
	userAccessTokenString, err := as.JwtUtil.GenerateToken(as.TokenSecurityKey.UserSecurityKey,uint64(otpReq.UserId), otp.Email, "user","access",15*time.Minute)
	if err != nil {
		return responsemodels.OtpVerificationResponse{}, fmt.Errorf("Failed to generarate access token for user: %w", err)
	}
	userRefreshTokenString,err:=as.JwtUtil.GenerateToken(as.TokenSecurityKey.UserRefreshKey,uint64(otpReq.UserId),otp.Email,"user","refresh",24*7*time.Hour)
	if err != nil {
		return responsemodels.OtpVerificationResponse{}, fmt.Errorf("Failed to generarate refresh token for user: %w", err)
	}
	//fmt.Println("userAccessTokenSting",userAccessTokenString)
	//fmt.Println("userRefreshTokenString",userRefreshTokenString)
	return responsemodels.OtpVerificationResponse{
		Email: otp.Email,
		Status: "verified",
		AccessToken: userAccessTokenString,
		RefreshToken: userRefreshTokenString,
	},nil
}

func(as *AuthSubscriptionUsecase)ResendOtp(resendOtpReq requestmodels.ResendOtpRequest)(responsemodels.ResendOtpResponse,error){
	err:=as.AuthSubscriptionRepository.DeleteOtpByEmail(resendOtpReq.Email)
	if err!=nil{
		return responsemodels.ResendOtpResponse{}, fmt.Errorf("database error: %w", err)
	}
	otp := as.RandomUtil.RandomNumber()
	expiration := time.Now().Add(5 * time.Minute)
	err = as.AuthSubscriptionRepository.TemporarySavingUserOtp(otp, resendOtpReq.Email, expiration)
	if err != nil {
		fmt.Println("cannont save otp in db")
		return responsemodels.ResendOtpResponse{}, fmt.Errorf("database error: %w", err)
	}
	err = as.SmtpUtil.SendVerifcationEmailWithOtp(otp, resendOtpReq.Email, resendOtpReq.Name)
	if err != nil {
		return responsemodels.ResendOtpResponse{}, fmt.Errorf("Error in sending otp to email address: %w", err)
	}
	return responsemodels.ResendOtpResponse{
		Email: resendOtpReq.Email,
	},nil
}

func (as *AuthSubscriptionUsecase) AccessRegenerator(accessRegeneratorReq requestmodels.AccessRegeneratorRequest)(responsemodels.AccessRegeneratorResponse,error){
	var accessTokenString string
	switch accessRegeneratorReq.Role{
	case "admin":
		adminAccessTokenString, err := as.JwtUtil.GenerateToken(as.TokenSecurityKey.AdminSecurityKey,uint64(accessRegeneratorReq.ID), accessRegeneratorReq.Email, "admin","access",24*time.Hour)
		if err != nil {
			return responsemodels.AccessRegeneratorResponse{}, fmt.Errorf("Failed to generarate access token for admin: %w", err)
		}
		accessTokenString=adminAccessTokenString
	case "user":
		userAccessTokenString, err := as.JwtUtil.GenerateToken(as.TokenSecurityKey.UserSecurityKey,uint64(accessRegeneratorReq.ID), accessRegeneratorReq.Email, "user","access",24*time.Hour)
		if err != nil {
			return responsemodels.AccessRegeneratorResponse{}, fmt.Errorf("Failed to generarate access token for user: %w", err)
		}
		accessTokenString=userAccessTokenString
	}
	
	return responsemodels.AccessRegeneratorResponse{
		Id: accessRegeneratorReq.ID,
		Email: accessRegeneratorReq.Email,
		Role: accessRegeneratorReq.Role,
		NewAccessToken: accessTokenString,
	},nil
}
func (as *AuthSubscriptionUsecase)ForgotPassword(forgotPasswordReq requestmodels.ForgotPasswordRequest)(responsemodels.ForgotPassordResponse,error){
	user,err:=as.AuthSubscriptionRepository.CheckUserExistsByEmail(forgotPasswordReq.Email)
	if err!=nil{
		if err==gorm.ErrRecordNotFound{
			return responsemodels.ForgotPassordResponse{},ErrUserNotFound
		} else{
			return responsemodels.ForgotPassordResponse{},fmt.Errorf("database error: %w",err)
		}
	}
	otp := as.RandomUtil.RandomNumber()
	err=as.AuthSubscriptionRepository.DeleteOtpByEmail(user.Email)
	if err!=nil{
		return responsemodels.ForgotPassordResponse{}, fmt.Errorf("database error: %w", err)
	}
	
	expiration := time.Now().Add(5 * time.Minute)
	err = as.AuthSubscriptionRepository.TemporarySavingUserOtp(otp, user.Email, expiration)
	if err != nil {
		fmt.Println("cannont save otp in db")
		return responsemodels.ForgotPassordResponse{}, fmt.Errorf("database error: %w", err)
	}
	err = as.SmtpUtil.SendResetPasswordEmailOtp(otp, user.Email)
	if err != nil {
		return responsemodels.ForgotPassordResponse{}, fmt.Errorf("Error in sending otp to email address: %w", err)
	}
	otpVerificationToken, err := as.JwtUtil.GenerateToken(as.TokenSecurityKey.OtpVerificationSecurityKey,uint64(user.ID),user.Email,"otpverification","access",5*time.Minute)
	if err != nil {
		return responsemodels.ForgotPassordResponse{}, fmt.Errorf("Failed to generarate token for otp verfication: %w", err)
	}
	return responsemodels.ForgotPassordResponse{
		Email: user.Email,
		TempToken: otpVerificationToken,
	},nil
}
func (as *AuthSubscriptionUsecase)ResetPassword(resetPasswordReq requestmodels.ResetPasswordRequest)(responsemodels.ResetPasswordResponse,error){
	if resetPasswordReq.Email==""{
		return responsemodels.ResetPasswordResponse{},ErrUserNotFound
	}
	hashedPassword := utils.HashPassword(resetPasswordReq.Password)
	resetPasswordReq.Password = hashedPassword
	err:=as.AuthSubscriptionRepository.UpdatePassword(resetPasswordReq)
	if err!=nil{
		return responsemodels.ResetPasswordResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.ResetPasswordResponse{
		Email: resetPasswordReq.Email,
	},nil
}



func (as *AuthSubscriptionUsecase)UserLogin(userLoginReq requestmodels.UserLoginRequest)(responsemodels.UserLoginResponse,error){
	user,err:=as.AuthSubscriptionRepository.CheckUserExistsByEmail(userLoginReq.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.UserLoginResponse{}, ErrUserNotFound
		}
		return responsemodels.UserLoginResponse{}, fmt.Errorf("database error: %w", err)
	}
	err=utils.CompareWithHashedPassword(user.Password,userLoginReq.Password)
	if err!=nil{
		return responsemodels.UserLoginResponse{},ErrInvalidCredentials
	}
	if user.Status=="blocked"{
		return responsemodels.UserLoginResponse{},ErrBlockedLogin
	}
	if user.Status=="pending"{
		return responsemodels.UserLoginResponse{},ErrPendingLogin
	}
	fmt.Println("inside user login ",user.ID)
	userAccessTokenString, err := as.JwtUtil.GenerateToken(as.TokenSecurityKey.UserSecurityKey,uint64(user.ID), user.Email, "user","access",24*time.Hour)
	if err != nil {
		return responsemodels.UserLoginResponse{}, fmt.Errorf("Failed to generarate access token for user: %w", err)
	}
	userRefreshTokenString,err:=as.JwtUtil.GenerateToken(as.TokenSecurityKey.UserRefreshKey,uint64(user.ID),user.Email,"user","refresh",24*7*time.Hour)
	if err != nil {
		return responsemodels.UserLoginResponse{}, fmt.Errorf("Failed to generarate refresh token for user: %w", err)
	}
	return responsemodels.UserLoginResponse{
		User: responsemodels.UserDetailsResponse{
			Id: uint64(user.ID),
			Name: user.Name,
			UserName: user.UserName,
			Email: user.Email,
			Status: user.Status,
			BlueTick: user.BlueTick,
		},
		AccessToken: userAccessTokenString,
		RefreshToken: userRefreshTokenString,
	},nil
}

func (as *AuthSubscriptionUsecase)CreateSubscriptionPlan(createSubscriptionPlanReq requestmodels.CreateSubscriptionPlanRequest)(responsemodels.CreateSubscriptionPlanResponse,error){
	// subscriptionPlan,err:=as.AuthSubscriptionRepository.CreateSubscriptionPlan(createSubscriptionPlanReq)
	// if err!=nil{
	// 	return  responsemodels.CreateSubscriptionPlanResponse{},fmt.Errorf("database error: %w", err)
	// }
	//razorpayClient:=utils.NewRazorpayClient(as.RazorpayCredentials.KeyId,as.RazorpayCredentials.KeySecret)
	planData := map[string]interface{}{
		"period":   createSubscriptionPlanReq.Period,
		"interval": createSubscriptionPlanReq.Interval,
		"item": map[string]interface{}{
			"name":        createSubscriptionPlanReq.Name,
			"amount":      createSubscriptionPlanReq.Price*100,
			"currency":    createSubscriptionPlanReq.Currency,
			"description": createSubscriptionPlanReq.Description,
		},
	}
	plan,err:=utils.RazorpayCreatePlan(as.RazorpayClient,planData)
	if err!=nil{
		fmt.Println("i think here is the error",err)
		return responsemodels.CreateSubscriptionPlanResponse{},err
	}
	subscriptionPlanRes,err:=as.AuthSubscriptionRepository.CreateSubscriptionPlan(plan)
	if err!=nil{
		return responsemodels.CreateSubscriptionPlanResponse{},fmt.Errorf("database error: %w", err)
	}
	return responsemodels.CreateSubscriptionPlanResponse{
		ID: subscriptionPlanRes.ID,
		CreatedAt: subscriptionPlanRes.CreatedAt,
		UpdatedAt: subscriptionPlanRes.UpdatedAt,
		Name: subscriptionPlanRes.Name,
		Price: subscriptionPlanRes.Price,
		Currency: subscriptionPlanRes.Currency,
		Period: subscriptionPlanRes.Period,
		Interval: subscriptionPlanRes.Interval,
		Description: subscriptionPlanRes.Description,
		IsActive: subscriptionPlanRes.IsActive,
	},nil
}


func (as *AuthSubscriptionUsecase)ActivateSubscriptionPlan(activateSubscriptionPlanReq requestmodels.ActivateSubscriptionPlanRequest)(responsemodels.ActivateSubscriptionPlanResponse,error){
	status,err:=as.AuthSubscriptionRepository.FetchStatusFromSubcriptionPlan(activateSubscriptionPlanReq.ID)
	if err!=nil{
		return responsemodels.ActivateSubscriptionPlanResponse{},fmt.Errorf("database error: %w",err)
	}
	if status{
		return responsemodels.ActivateSubscriptionPlanResponse{},ErrSubscriptionPlanAlreadyActive
	}
	subscriptionPlan,err:=as.AuthSubscriptionRepository.ActivateSubscriptionPlan(activateSubscriptionPlanReq)
	if err!=nil{
		return responsemodels.ActivateSubscriptionPlanResponse{},fmt.Errorf("database error: %w", err)
	}
	return responsemodels.ActivateSubscriptionPlanResponse{
		ID: subscriptionPlan.ID,
		CreatedAt: subscriptionPlan.CreatedAt,
		UpdatedAt: subscriptionPlan.UpdatedAt,
		RazorpayPlanId: subscriptionPlan.RazorpayPlanId,
		Name: subscriptionPlan.Name,
		Price: subscriptionPlan.Price,
		Currency: subscriptionPlan.Currency,
		Period: subscriptionPlan.Period,
		Interval: subscriptionPlan.Interval,
		Description: subscriptionPlan.Description,
		IsActive: subscriptionPlan.IsActive,
	},nil
}

func (as *AuthSubscriptionUsecase)DeactivateSubscriptionPlan(deactivateSubscriptionPlanReq requestmodels.DeactivateSubscriptionPlanRequest)(responsemodels.DeactivateSubscriptionPlanResponse,error){
	status,err:=as.AuthSubscriptionRepository.FetchStatusFromSubcriptionPlan(deactivateSubscriptionPlanReq.ID)
	if err!=nil{
		return responsemodels.DeactivateSubscriptionPlanResponse{},fmt.Errorf("database error: %w",err)
	}
	if !status{
		return responsemodels.DeactivateSubscriptionPlanResponse{},ErrSubscriptionPlanAlreadyDeactive
	}
	subscriptionPlan,err:=as.AuthSubscriptionRepository.DeactivateSubscriptionPlan(deactivateSubscriptionPlanReq)
	if err!=nil{
		return responsemodels.DeactivateSubscriptionPlanResponse{},fmt.Errorf("database error: %w", err)
	}
	return responsemodels.DeactivateSubscriptionPlanResponse{
		ID: subscriptionPlan.ID,
		CreatedAt: subscriptionPlan.CreatedAt,
		UpdatedAt: subscriptionPlan.UpdatedAt,
		RazorpayPlanId: subscriptionPlan.RazorpayPlanId,
		Name: subscriptionPlan.Name,
		Price: subscriptionPlan.Price,
		Currency: subscriptionPlan.Currency,
		Period: subscriptionPlan.Period,
		Interval: subscriptionPlan.Interval,
		Description: subscriptionPlan.Description,
		IsActive: subscriptionPlan.IsActive,
	},nil
}

func (as *AuthSubscriptionUsecase)GetAllSubscriptionPlans(getAllSubscripionPlansReq requestmodels.GetAllSubscriptionPlansRequest)(responsemodels.GetAllSubscriptionPlansResponse,error){
	subscriptionPlans,err:=as.AuthSubscriptionRepository.GetAllSubscriptionPlans(getAllSubscripionPlansReq)
	if err!=nil{
		return responsemodels.GetAllSubscriptionPlansResponse{},fmt.Errorf("database error: %w", err)
	}
	return responsemodels.GetAllSubscriptionPlansResponse{
		SubscriptionPlans: subscriptionPlans.SubscriptionPlans,
	},nil
}

func (as *AuthSubscriptionUsecase)GetAllActiveSubscriptionPlans(getAllActiveSubscriptionPlansReq requestmodels.GetAllActiveSubscriptionPlansRequest)(responsemodels.GetAllActiveSubscriptionPlansResponse,error){
	subscriptionPlans,err:=as.AuthSubscriptionRepository.GetAllActiveSubscriptionPlans(getAllActiveSubscriptionPlansReq)
	if err!=nil{
		return responsemodels.GetAllActiveSubscriptionPlansResponse{},fmt.Errorf("database error: %w", err)
	}
	return responsemodels.GetAllActiveSubscriptionPlansResponse{
		SubscriptionPlans: subscriptionPlans.SubscriptionPlans,
	},nil
}

func (as *AuthSubscriptionUsecase)Subscribe(subscribeReq requestmodels.SubscribeRequest)(responsemodels.SubscribeResponse,error){
	RazorpayPlanId,err:=as.AuthSubscriptionRepository.FetchRazorpayPlanIdFromId(subscribeReq.PlanId)
	if err!=nil{
		return responsemodels.SubscribeResponse{},fmt.Errorf("database error: %w",err)
	}
	//razorpayClient:=utils.NewRazorpayClient(as.RazorpayCredentials.KeyId,as.RazorpayCredentials.KeySecret)
	subscriptionData := map[string]interface{}{
		"plan_id":         RazorpayPlanId,
		"total_count":     12,
		"quantity":        1,
		"customer_notify": 1,
	}
	subscription,err:=utils.RazorpayCreateSubscription(as.RazorpayClient,subscriptionData)
	if err!=nil{
		fmt.Println("error on subscribing",err)
		return responsemodels.SubscribeResponse{},fmt.Errorf("database error: %w",err)
	}
	//fmt.Println(subscription)
	subcribeRes,err:=as.AuthSubscriptionRepository.CreateSubscription(subscribeReq,subscription)
	if err!=nil{
		fmt.Printf("is there any error returning after createSubscripion %v",err)
		return responsemodels.SubscribeResponse{},err
	}
	//fmt.Println("subscribeRes",subcribeRes)
	return subcribeRes,nil
}
func (as *AuthSubscriptionUsecase) pollRazorpayAndSync(subid string) {
    ticker := time.NewTicker(30 * time.Second) // Poll every 30 seconds
    defer ticker.Stop()
	count:=0
    for range ticker.C{
      
        // Call Razorpay API to get subscription data
        razorpayData, err := as.RazorpayClient.Subscription.Fetch(subid,nil,nil)
        if err != nil {
			count++
			if count==20{
				return
			}
            fmt.Println("Error fetching Razorpay subscription:", err)
            continue//retry on error
        }
		// Check if charge_at exists and is a valid value
		chargeAt, ok := razorpayData["charge_at"].(float64) // or try .(string) or other types if needed
		if !ok || chargeAt == 0 {
			count++
			if count==20{
				return
			}
			fmt.Println("charge_at not populated or invalid")
			continue // Retry if charge_at is null or invalid
		}
		// Sync the data if charge_at is populated
		_, err = as.AuthSubscriptionRepository.UpdateUserSubscripion(subid, razorpayData)
		if err != nil {
			count++
			if count==20{
				return
			}
			fmt.Println("Error syncing Razorpay subscription data:", err)
			continue // Retry on sync error
		}
		// Successfully synced, print success and return
		fmt.Println("Successfully synced subscription data for:", subid)
		return // Stop polling after a successful update
    }
}

// Function to calculate `end_at` and `NextChargeAt` based on period, interval, and total_count
func calculateEndAndNextChargeTime(startAt time.Time,period string,interval uint64, totalCount int) (time.Time, time.Time) {
	// Calculate the total period
	totalInterval := int(interval) * totalCount

	// Variable to store the calculated end_at
	var endAt time.Time
	var nextChargeAt time.Time

	// Calculate the end date based on the interval and total period
	switch period {
	case "day":
		// If the interval is in days, add the total period in days
		endAt = startAt.AddDate(0, 0, totalInterval)
		nextChargeAt = endAt.AddDate(0, 0, int(interval)) // Next charge is after the `end_at` by 1 period (in days)
	case "month":
		// If the interval is in months, add the total period in months
		endAt = startAt.AddDate(0, totalInterval, 0)
		nextChargeAt = endAt.AddDate(0, int(interval), 0) // Next charge is after the `end_at` by 1 period (in months)
	case "year":
		// If the interval is in years, add the total period in years
		endAt = startAt.AddDate(totalInterval, 0, 0)
		nextChargeAt = endAt.AddDate(int(interval), 0, 0) // Next charge is after the `end_at` by 1 period (in years)
	default:
		// Default case if the interval is not recognized
		endAt = startAt
		nextChargeAt = startAt
	}
	fmt.Println("endAt---",endAt,"nextChargeAt---",nextChargeAt)
	// Return both end_at and next_charge_at
	return endAt, nextChargeAt
}

func (as *AuthSubscriptionUsecase)VerifySubscriptionPayment(verifySubscriptionPaymentReq requestmodels.VerifySubscriptionPaymentRequest)(responsemodels.VerifySubscriptionPaymentResponse,error){
	var subscriptionRes responsemodels.VerifySubscriptionPaymentResponse
	//razorpayClient:=utils.NewRazorpayClient(as.RazorpayCredentials.KeyId,as.RazorpayCredentials.KeySecret)
	subscription, err := as.RazorpayClient.Subscription.Fetch(verifySubscriptionPaymentReq.RazorpaySubscriptionId, nil,nil)
	fmt.Println("------------------------")
	//fmt.Println("subscription",subscription)
	if err != nil {
    return responsemodels.VerifySubscriptionPaymentResponse{},err
	}
	startAt,ok := subscription["start_at"].(float64)
	fmt.Println("print value start at",startAt)
	if !ok{
		fmt.Println("what if its coming here *******")
		planId,err:=as.AuthSubscriptionRepository.FetchRazorpayPlanIdFromRazrorpaySubscriptionId(verifySubscriptionPaymentReq.RazorpaySubscriptionId)
		if err!=nil{
			return responsemodels.VerifySubscriptionPaymentResponse{},fmt.Errorf("database error :%w",err)
		}
		period,interval,err:=as.AuthSubscriptionRepository.FetchIntervalPeriodFromSubscriptionPlan(planId)
		if err!=nil{
			return responsemodels.VerifySubscriptionPaymentResponse{},fmt.Errorf("database error: %w",err)
		}
		totalCount,err:=as.AuthSubscriptionRepository.FetchTotalCountFromUserSubscription(verifySubscriptionPaymentReq.RazorpaySubscriptionId)
		if err!=nil{
			return responsemodels.VerifySubscriptionPaymentResponse{},fmt.Errorf("database error: %w",err)
		}
		startAt:=time.Now()
		// Calculate the end_at and NextChargeAt times
		endAt, nextChargeAt := calculateEndAndNextChargeTime(startAt, period,interval, totalCount)
		fmt.Println("print inside",startAt,endAt,nextChargeAt)
		subscriptionRes,err=as.AuthSubscriptionRepository.UpdateTimeUserSubscription(startAt,endAt,nextChargeAt,verifySubscriptionPaymentReq.RazorpaySubscriptionId)
		if err!=nil{
			return responsemodels.VerifySubscriptionPaymentResponse{},fmt.Errorf("database error: %w",err)
		}
		go as.pollRazorpayAndSync(verifySubscriptionPaymentReq.RazorpaySubscriptionId)
	} else{
		fmt.Println("hi i hope its here------")
		subscriptionRes,err=as.AuthSubscriptionRepository.UpdateUserSubscripion(verifySubscriptionPaymentReq.RazorpaySubscriptionId,subscription)
		if err!=nil{
			return responsemodels.VerifySubscriptionPaymentResponse{},err
		}
	}
	
	
	userid,err:=as.AuthSubscriptionRepository.FetchUserIdFromSubscriptionId(verifySubscriptionPaymentReq.RazorpaySubscriptionId)
	if err!=nil{
		return responsemodels.VerifySubscriptionPaymentResponse{},err
	}
	err=as.AuthSubscriptionRepository.TurnBlueTickTrueForUserId(userid)
	if err!=nil{
		return responsemodels.VerifySubscriptionPaymentResponse{},err
	}
	payment,err:=as.RazorpayClient.Payment.Fetch(verifySubscriptionPaymentReq.RazorpayPaymentId,nil,nil)
	if err!=nil{
		return responsemodels.VerifySubscriptionPaymentResponse{},err
	}
	//fmt.Println("payment : ",payment)
	_,err=as.AuthSubscriptionRepository.PopulatePayment(payment,verifySubscriptionPaymentReq)
	if err!=nil{
		return responsemodels.VerifySubscriptionPaymentResponse{},nil
	}
	//fmt.Println("payment table",paymentRes)
	fmt.Println("just before into service ",subscriptionRes.StartAt,subscriptionRes.NextChargeAt)
	return subscriptionRes,nil
}

func (as *AuthSubscriptionUsecase) Unsubscribe(unsubscribeReq requestmodels.UnsubscribeRequest)(responsemodels.UnsubscribeResponse,error){
	data:=map[string]interface{}{
		"cancel_at_cycle_end":false,
	}
	razorpaySubscritpionId,err:=as.AuthSubscriptionRepository.FetchRazorpaySubscriptionIdFromSubcriptionId(unsubscribeReq.SubId)
	if err!=nil{
		return  responsemodels.UnsubscribeResponse{},fmt.Errorf("database error: %w",err)
	}
	resp,err:=as.RazorpayClient.Subscription.Cancel(razorpaySubscritpionId,data,nil)
	if err!=nil{
		fmt.Println("print the error on cancellation razorpay api call",err)
		return responsemodels.UnsubscribeResponse{},err
	}
	fmt.Println("is it actually nil,???",unsubscribeReq.SubId)
	unsubscibeRes,err:=as.AuthSubscriptionRepository.ChangeUserSubscriptionStatusToCancelled(unsubscribeReq.SubId,resp)
	if err!=nil{
		return responsemodels.UnsubscribeResponse{},err
	}
	userid,err:=as.AuthSubscriptionRepository.FetchUserIdFromSubscriptionId(razorpaySubscritpionId)
	if err!=nil{
		return responsemodels.UnsubscribeResponse{},err
	}
	nextChargeAt,err:=as.AuthSubscriptionRepository.FetchNextChargeAtFromUserSubcription(razorpaySubscritpionId)
	if err!=nil{
		return responsemodels.UnsubscribeResponse{},err
	}
	delay:=time.Until(nextChargeAt)
	go func ()  {
		<-time.After(delay)
		err:=as.AuthSubscriptionRepository.TurnOffBlueTickForUserId(userid)
		if err!=nil{
			fmt.Println("error while turning off blue tick",err)
		}
	}()
	return unsubscibeRes,nil
}
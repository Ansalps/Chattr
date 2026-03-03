package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/infrastructure/jwt"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/infrastructure/razorpaygateway"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/infrastructure/smtp"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/responsemodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/usecase/interfacesUsecase"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AuthSubscriptionUsecase struct {
	SmtpProvider               *smtp.SmtpCredentials
	AuthSubscriptionRepository interfacesRepository.AuthSubscriptionRepository
	Config                     *config.Config
	JwtProvider                *jwt.JwtProvider
	//RazorpayCredentials	*config.Razorpay
	RazorpayGateway *razorpaygateway.RazorpayGateway
	AwsS3Client     *s3.Client
	AwsBucket       string
}

func NewAuthSubscriptionUsecase(repository interfacesRepository.AuthSubscriptionRepository,
	smtpProvider *smtp.SmtpCredentials, config *config.Config, jwtProvider *jwt.JwtProvider /*razorpayCredentials *config.Razorpay,*/, razorpayGateway *razorpaygateway.RazorpayGateway, awsS3Client *s3.Client, awsBucket string) interfacesUsecase.AuthSubscriptionUsecase {
	return &AuthSubscriptionUsecase{
		AuthSubscriptionRepository: repository,
		SmtpProvider:               smtpProvider,
		Config:                     config,
		JwtProvider:                jwtProvider,
		//RazorpayCredentials: razorpayCredentials,
		RazorpayGateway: razorpayGateway,
		AwsS3Client:     awsS3Client,
		AwsBucket:       awsBucket,
	}
}

var (

// ErrRazropayApi=errors.New("error calling razorpay api")
)

func (as *AuthSubscriptionUsecase) DoesUserExists(userid uint64) (bool, error) {
	resp, err := as.AuthSubscriptionRepository.DoesUserExists(userid)
	if err != nil {
		return false, err
	}
	return resp, nil
}
func (as *AuthSubscriptionUsecase) AdminLogin(ctx context.Context, admin requestmodels.AdminLoginRequest) (responsemodels.AdminLoginResponse, error) {
	admins, err := as.AuthSubscriptionRepository.CheckAdminExistsByEmail(ctx, admin.Email)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDatabaseConnectionTimeOut):
			return responsemodels.AdminLoginResponse{}, err
		case errors.Is(err, gorm.ErrRecordNotFound):
			return responsemodels.AdminLoginResponse{}, domain.ErrUserNotFound
		default:
			return responsemodels.AdminLoginResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
		}
	}
	if admin.Password != admins.Password {
		return responsemodels.AdminLoginResponse{}, domain.ErrInvalidCredentials
	}
	adminAccessTokenString, err := as.JwtProvider.GenerateToken(as.Config.Token.AdminSecurityKey, uint64(admins.ID), admins.Email, "admin", "access", 24*time.Hour)
	if err != nil {
		return responsemodels.AdminLoginResponse{}, fmt.Errorf("%w: %v:", domain.ErrAdminAccessTokenFail, err)
	}
	adminRefreshTokenString, err := as.JwtProvider.GenerateToken(as.Config.Token.AdminRefreshKey, uint64(admins.ID), admins.Email, "admin", "refresh", 24*7*time.Hour)
	if err != nil {
		return responsemodels.AdminLoginResponse{}, fmt.Errorf("%w: %v:", domain.ErrAdminRefreshTokenFail, err)
	}
	return responsemodels.AdminLoginResponse{
		Admin: responsemodels.AdminDetails{
			ID:    admins.ID,
			Email: admins.Email,
		},
		AccessToken:  adminAccessTokenString,
		RefreshToken: adminRefreshTokenString,
	}, nil
}

func (as *AuthSubscriptionUsecase) BlockUser(blockUserReq requestmodels.BlockUserRequest) (responsemodels.BlockUserResponse, error) {
	if blockUserReq.UserId == 0 {
		return responsemodels.BlockUserResponse{}, domain.ErrUserNotFound
	}
	status, err := as.AuthSubscriptionRepository.CheckUserStatus(blockUserReq.UserId)
	if err != nil {
		return responsemodels.BlockUserResponse{}, fmt.Errorf("database error: %w", err)
	}
	if status != "active" {
		return responsemodels.BlockUserResponse{}, domain.ErrUserNotActive
	}
	err = as.AuthSubscriptionRepository.ChangeUserStatusToBlockedByUserId(blockUserReq)
	if err != nil {
		return responsemodels.BlockUserResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.BlockUserResponse{
		UserId: blockUserReq.UserId,
	}, nil
}

func (as *AuthSubscriptionUsecase) UnblockUser(unblockUserReq requestmodels.UnblockUserRequest) (responsemodels.UnblockUserResponse, error) {
	if unblockUserReq.UserId == 0 {
		return responsemodels.UnblockUserResponse{}, domain.ErrUserNotFound
	}
	status, err := as.AuthSubscriptionRepository.CheckUserStatus(unblockUserReq.UserId)
	if err != nil {
		return responsemodels.UnblockUserResponse{}, fmt.Errorf("database error: %w", err)
	}
	if status != "blocked" {
		return responsemodels.UnblockUserResponse{}, domain.ErrUserNotBlocked
	}
	err = as.AuthSubscriptionRepository.ChangeUserStatusToActiveByUserId(unblockUserReq)
	if err != nil {
		return responsemodels.UnblockUserResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.UnblockUserResponse{
		UserId: unblockUserReq.UserId,
	}, nil
}

func (as *AuthSubscriptionUsecase) GetAllUsers(getAllUsersReq requestmodels.GetAllUsersRequest) (responsemodels.GetAllUsersResponse, error) {
	users, err := as.AuthSubscriptionRepository.GetAllUsers(getAllUsersReq)
	if err != nil {
		return responsemodels.GetAllUsersResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.GetAllUsersResponse{
		Users: users.Users,
	}, nil
}

func (as *AuthSubscriptionUsecase) UserSignUp(ctx context.Context, userReq requestmodels.UserSignUpRequest) (responsemodels.UserSignupResponse, error) {
	err := as.AuthSubscriptionRepository.DeletePendingUser(ctx, userReq.Email)
	if err != nil {
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.UserSignupResponse{}, err
		}
		return responsemodels.UserSignupResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
	}
	user, err := as.AuthSubscriptionRepository.CheckUserExistsByEmail(ctx, userReq.Email)
	if err != nil {
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.UserSignupResponse{}, err
		}
		if err != gorm.ErrRecordNotFound {
			return responsemodels.UserSignupResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
		}
	}
	if user != nil {
		return responsemodels.UserSignupResponse{
			ID:       user.ID,
			UserName: user.UserName,
			Name:     user.Name,
			Email:    user.Email,
		}, domain.ErrUserAlreadyExistsByEmail
	}
	usernameAlredayExists, err := as.AuthSubscriptionRepository.CheckUserExistsByUseraname(ctx, userReq.UserName)
	if err != nil {
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.UserSignupResponse{}, err
		}
		if err != gorm.ErrRecordNotFound {
			return responsemodels.UserSignupResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
		}
	}
	if usernameAlredayExists != nil {
		return responsemodels.UserSignupResponse{
			ID:       usernameAlredayExists.ID,
			UserName: usernameAlredayExists.UserName,
			Name:     usernameAlredayExists.Name,
			Email:    usernameAlredayExists.Email,
		}, domain.ErrUserAlreadyExistsByUsername
	}
	err = as.AuthSubscriptionRepository.DeleteOtpByEmail(ctx, userReq.Email)
	if err != nil {
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.UserSignupResponse{}, err
		}
		if err != gorm.ErrRecordNotFound {
			return responsemodels.UserSignupResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
		}
	}
	otp := utils.RandomNumber()
	fmt.Println("Otp is ----", otp)
	expiration := time.Now().Add(5 * time.Minute)
	err = as.AuthSubscriptionRepository.TemporarySavingUserOtp(ctx, otp, userReq.Email, expiration)
	if err != nil {
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.UserSignupResponse{}, err
		}
		return responsemodels.UserSignupResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
	}
	go func() {
		err = as.SmtpProvider.SendVerifcationEmailWithOtp(otp, userReq.Email, userReq.Name)
		if err != nil {
			//return responsemodels.UserSignupResponse{}, fmt.Errorf("%w: %v:", domain.ErrSendVerifyOtpToEmail, err)
			log.Printf("Failed to send email to %s: %v", userReq.Email, err)
		}
	}()
	hashedPassword := utils.HashPassword(userReq.ConfirmPassword)
	userReq.Password = hashedPassword

	userRes, err := as.AuthSubscriptionRepository.CreateUser(ctx, &userReq)
	if err != nil {
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.UserSignupResponse{}, err
		}
		return responsemodels.UserSignupResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
	}
	//fmt.Println("userRes.ID is ", userRes.ID)
	otpVerificationToken, err := as.JwtProvider.GenerateToken(as.Config.Token.OtpVerificationSecurityKey, uint64(userRes.ID), userRes.Email, "otpverification", "access", 5*time.Minute)
	if err != nil {
		return responsemodels.UserSignupResponse{}, fmt.Errorf("%w: %v:", domain.ErrVerifyOtpTokenFail, err)
	}
	return responsemodels.UserSignupResponse{
		ID:                   userRes.ID,
		UserName:             userRes.UserName,
		Name:                 userRes.Name,
		Email:                userRes.Email,
		OtpVerificationToken: otpVerificationToken,
	}, nil
}

func (as *AuthSubscriptionUsecase) VerifyOtp(ctx context.Context,otpReq requestmodels.OtpRequest) (responsemodels.OtpVerificationResponse, error) {
	otp, err := as.AuthSubscriptionRepository.CheckOtpExistsByEmail(otpReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.OtpVerificationResponse{}, domain.ErrUserNotFound
		}
		return responsemodels.OtpVerificationResponse{}, fmt.Errorf("%w: %v:",domain.ErrDatabase, err)
	}
	if otp.OTP != otpReq.OtpCode {
		return responsemodels.OtpVerificationResponse{}, domain.ErrInvalidCredentials
	}
	if time.Now().After(otp.Expiration) {
		return responsemodels.OtpVerificationResponse{}, domain.ErrOtpExpired
	}
	err = as.AuthSubscriptionRepository.ChangeOtpStatus(otpReq.Email)
	if err != nil {
		return responsemodels.OtpVerificationResponse{}, fmt.Errorf("%w: %v:",domain.ErrDatabase, err)
	}
	err = as.AuthSubscriptionRepository.ChangeUserStatusByEmail(otpReq.Email)
	if err != nil {
		return responsemodels.OtpVerificationResponse{}, fmt.Errorf("%w: %v:",domain.ErrDatabase, err)
	}
	if otpReq.Purpose == "user-forgot-password" {
		resetPasswordToken, err := as.JwtProvider.GenerateToken(as.Config.Token.ResetPasswordSecurityKey, uint64(otpReq.UserId), otp.Email, "resetpassword", "access", 5*time.Minute)
		if err != nil {
			return responsemodels.OtpVerificationResponse{}, fmt.Errorf("%w: %v:",domain.ErrVerifyOtpTokenFail, err)
		}
		return responsemodels.OtpVerificationResponse{
			Email:     otp.Email,
			Status:    "verified",
			TempToken: resetPasswordToken,
		}, nil
	}
	userAccessTokenString, err := as.JwtProvider.GenerateToken(as.Config.Token.UserSecurityKey, uint64(otpReq.UserId), otp.Email, "user", "access", 24*time.Hour)
	if err != nil {
		return responsemodels.OtpVerificationResponse{}, fmt.Errorf("%w: %v:",domain.ErrUserAccessTokenFail, err)
	}
	userRefreshTokenString, err := as.JwtProvider.GenerateToken(as.Config.Token.UserRefreshKey, uint64(otpReq.UserId), otp.Email, "user", "refresh", 24*7*time.Hour)
	if err != nil {
		return responsemodels.OtpVerificationResponse{}, fmt.Errorf("%w: %v:",domain.ErrUserRefreshTokenFail, err)
	}
	//fmt.Println("userAccessTokenSting",userAccessTokenString)
	//fmt.Println("userRefreshTokenString",userRefreshTokenString)
	return responsemodels.OtpVerificationResponse{
		Email:        otp.Email,
		Status:       "verified",
		AccessToken:  userAccessTokenString,
		RefreshToken: userRefreshTokenString,
	}, nil
}

func (as *AuthSubscriptionUsecase) ResendOtp(ctx context.Context,resendOtpReq requestmodels.ResendOtpRequest) (responsemodels.ResendOtpResponse, error) {
	err := as.AuthSubscriptionRepository.DeleteOtpByEmail(ctx,resendOtpReq.Email)
	if err != nil {
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.ResendOtpResponse{}, err
		}
		return responsemodels.ResendOtpResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
	}
	otp := utils.RandomNumber()
	expiration := time.Now().Add(5 * time.Minute)
	err = as.AuthSubscriptionRepository.TemporarySavingUserOtp(ctx,otp, resendOtpReq.Email, expiration)
	if err != nil {
		//return responsemodels.ResendOtpResponse{}, fmt.Errorf("database error: %w", err)
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.ResendOtpResponse{}, err
		}
		return responsemodels.ResendOtpResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
	}
	go func(){
		err = as.SmtpProvider.SendVerifcationEmailWithOtp(otp, resendOtpReq.Email, resendOtpReq.Name)
		if err != nil {
			//return responsemodels.ResendOtpResponse{}, fmt.Errorf("Error in sending otp to email address: %w", err)
			//return responsemodels.UserSignupResponse{}, fmt.Errorf("%w: %v:", domain.ErrSendVerifyOtpToEmail, err)
			log.Printf("Failed to send email to %s: %v", resendOtpReq.Email, err)
		}
	}()
	return responsemodels.ResendOtpResponse{
		Email: resendOtpReq.Email,
	}, nil
}

func (as *AuthSubscriptionUsecase) AccessRegenerator(accessRegeneratorReq requestmodels.AccessRegeneratorRequest) (responsemodels.AccessRegeneratorResponse, error) {
	var accessTokenString string
	switch accessRegeneratorReq.Role {
	case "admin":
		adminAccessTokenString, err := as.JwtProvider.GenerateToken(as.Config.Token.AdminSecurityKey, uint64(accessRegeneratorReq.ID), accessRegeneratorReq.Email, "admin", "access", 24*time.Hour)
		if err != nil {
			return responsemodels.AccessRegeneratorResponse{}, fmt.Errorf("%w: %v:", domain.ErrAdminAccessTokenFail, err)
		}
		accessTokenString = adminAccessTokenString
	case "user":
		userAccessTokenString, err := as.JwtProvider.GenerateToken(as.Config.Token.UserSecurityKey, uint64(accessRegeneratorReq.ID), accessRegeneratorReq.Email, "user", "access", 24*time.Hour)
		if err != nil {
			return responsemodels.AccessRegeneratorResponse{}, fmt.Errorf("%w: %v:", domain.ErrUserAccessTokenFail, err)
		}
		accessTokenString = userAccessTokenString
	}

	return responsemodels.AccessRegeneratorResponse{
		Id:             accessRegeneratorReq.ID,
		Email:          accessRegeneratorReq.Email,
		Role:           accessRegeneratorReq.Role,
		NewAccessToken: accessTokenString,
	}, nil
}
func (as *AuthSubscriptionUsecase) ForgotPassword(ctx context.Context,forgotPasswordReq requestmodels.ForgotPasswordRequest) (responsemodels.ForgotPassordResponse, error) {
	user, err := as.AuthSubscriptionRepository.CheckUserExistsByEmail(ctx,forgotPasswordReq.Email)
	if err != nil {
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.ForgotPassordResponse{}, err
		}
		if err == gorm.ErrRecordNotFound {
			return  responsemodels.ForgotPassordResponse{}, domain.ErrUserNotFound
		}
		return responsemodels.ForgotPassordResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
	}
	otp := utils.RandomNumber()
	err = as.AuthSubscriptionRepository.DeleteOtpByEmail(ctx,user.Email)
	if err != nil {
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.ForgotPassordResponse{}, err
		}
		return responsemodels.ForgotPassordResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
	}

	expiration := time.Now().Add(5 * time.Minute)
	err = as.AuthSubscriptionRepository.TemporarySavingUserOtp(ctx,otp, user.Email, expiration)
	if err != nil {
		//log.Println("cannont save otp in db")
		//return responsemodels.ForgotPassordResponse{}, fmt.Errorf("database error: %w", err)
		if errors.Is(err, domain.ErrDatabaseConnectionTimeOut) {
			return responsemodels.ForgotPassordResponse{}, err
		}
		return responsemodels.ForgotPassordResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
	}
	go func(){
	err = as.SmtpProvider.SendResetPasswordEmailOtp(otp, user.Email)
	if err != nil {
		//return responsemodels.ForgotPassordResponse{}, fmt.Errorf("Error in sending otp to email address: %w", err)
		log.Printf("Failed to send email to %s: %v", forgotPasswordReq.Email, err)
	}
	}()
	otpVerificationToken, err := as.JwtProvider.GenerateToken(as.Config.Token.OtpVerificationSecurityKey, uint64(user.ID), user.Email, "otpverification", "access", 5*time.Minute)
	if err != nil {
		return responsemodels.ForgotPassordResponse{}, fmt.Errorf("%w: %v:",domain.ErrVerifyOtpTokenFail, err)
	}
	return responsemodels.ForgotPassordResponse{
		Email:     user.Email,
		TempToken: otpVerificationToken,
	}, nil
}
func (as *AuthSubscriptionUsecase) ResetPassword(resetPasswordReq requestmodels.ResetPasswordRequest) (responsemodels.ResetPasswordResponse, error) {
	if resetPasswordReq.Email == "" {
		return responsemodels.ResetPasswordResponse{}, domain.ErrUserNotFound
	}
	hashedPassword := utils.HashPassword(resetPasswordReq.Password)
	resetPasswordReq.Password = hashedPassword
	err := as.AuthSubscriptionRepository.UpdatePassword(resetPasswordReq)
	if err != nil {
		return responsemodels.ResetPasswordResponse{}, fmt.Errorf("%w: %v:",domain.ErrDatabase, err)
	}
	return responsemodels.ResetPasswordResponse{
		Email: resetPasswordReq.Email,
	}, nil
}

func (as *AuthSubscriptionUsecase) UserLogin(ctx context.Context,userLoginReq requestmodels.UserLoginRequest) (responsemodels.UserLoginResponse, error) {
	user, err := as.AuthSubscriptionRepository.CheckUserExistsByEmail(ctx,userLoginReq.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.UserLoginResponse{}, domain.ErrUserNotFound
		}
		return responsemodels.UserLoginResponse{}, fmt.Errorf("database error: %w", err)
	}
	err = utils.CompareWithHashedPassword(user.Password, userLoginReq.Password)
	if err != nil {
		return responsemodels.UserLoginResponse{}, domain.ErrInvalidCredentials
	}
	if user.Status == "blocked" {
		return responsemodels.UserLoginResponse{}, domain.ErrBlockedLogin
	}
	if user.Status == "pending" {
		return responsemodels.UserLoginResponse{}, domain.ErrPendingLogin
	}
	//fmt.Println("inside user login ", user.ID)
	userAccessTokenString, err := as.JwtProvider.GenerateToken(as.Config.Token.UserSecurityKey, uint64(user.ID), user.Email, "user", "access", 24*time.Hour)
	if err != nil {
		return responsemodels.UserLoginResponse{}, fmt.Errorf("Failed to generarate access token for user: %w", err)
	}
	userRefreshTokenString, err := as.JwtProvider.GenerateToken(as.Config.Token.UserRefreshKey, uint64(user.ID), user.Email, "user", "refresh", 24*7*time.Hour)
	if err != nil {
		return responsemodels.UserLoginResponse{}, fmt.Errorf("Failed to generarate refresh token for user: %w", err)
	}
	return responsemodels.UserLoginResponse{
		User: responsemodels.UserDetailsResponse{
			Id:       uint64(user.ID),
			Name:     user.Name,
			UserName: user.UserName,
			Email:    user.Email,
			Status:   user.Status,
			BlueTick: user.BlueTick,
		},
		AccessToken:  userAccessTokenString,
		RefreshToken: userRefreshTokenString,
	}, nil
}

func (as *AuthSubscriptionUsecase) CreateSubscriptionPlan(createSubscriptionPlanReq requestmodels.CreateSubscriptionPlanRequest) (responsemodels.CreateSubscriptionPlanResponse, error) {

	planData := map[string]interface{}{
		"period":   createSubscriptionPlanReq.Period,
		"interval": createSubscriptionPlanReq.Interval,
		"item": map[string]interface{}{
			"name":        createSubscriptionPlanReq.Name,
			"amount":      createSubscriptionPlanReq.Price * 100,
			"currency":    createSubscriptionPlanReq.Currency,
			"description": createSubscriptionPlanReq.Description,
		},
	}
	plan, err := as.RazorpayGateway.CreatePlan(planData)
	if err != nil {
		//fmt.Println("i think here is the error", err)
		return responsemodels.CreateSubscriptionPlanResponse{}, err
	}
	subscriptionPlanRes, err := as.AuthSubscriptionRepository.CreateSubscriptionPlan(plan)
	if err != nil {
		return responsemodels.CreateSubscriptionPlanResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.CreateSubscriptionPlanResponse{
		ID:          subscriptionPlanRes.ID,
		CreatedAt:   subscriptionPlanRes.CreatedAt,
		UpdatedAt:   subscriptionPlanRes.UpdatedAt,
		Name:        subscriptionPlanRes.Name,
		Price:       subscriptionPlanRes.Price,
		Currency:    subscriptionPlanRes.Currency,
		Period:      subscriptionPlanRes.Period,
		Interval:    subscriptionPlanRes.Interval,
		Description: subscriptionPlanRes.Description,
		IsActive:    subscriptionPlanRes.IsActive,
	}, nil
}

func (as *AuthSubscriptionUsecase) ActivateSubscriptionPlan(activateSubscriptionPlanReq requestmodels.ActivateSubscriptionPlanRequest) (responsemodels.ActivateSubscriptionPlanResponse, error) {
	status, err := as.AuthSubscriptionRepository.FetchStatusFromSubcriptionPlan(activateSubscriptionPlanReq.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.ActivateSubscriptionPlanResponse{}, domain.ErrSubPlanNotFound
		}
		return responsemodels.ActivateSubscriptionPlanResponse{}, fmt.Errorf("database error: %w", err)
	}
	if status {
		return responsemodels.ActivateSubscriptionPlanResponse{}, domain.ErrSubscriptionPlanAlreadyActive
	}
	subscriptionPlan, err := as.AuthSubscriptionRepository.ActivateSubscriptionPlan(activateSubscriptionPlanReq)
	if err != nil {
		return responsemodels.ActivateSubscriptionPlanResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.ActivateSubscriptionPlanResponse{
		ID:             subscriptionPlan.ID,
		CreatedAt:      subscriptionPlan.CreatedAt,
		UpdatedAt:      subscriptionPlan.UpdatedAt,
		RazorpayPlanId: subscriptionPlan.RazorpayPlanId,
		Name:           subscriptionPlan.Name,
		Price:          subscriptionPlan.Price,
		Currency:       subscriptionPlan.Currency,
		Period:         subscriptionPlan.Period,
		Interval:       subscriptionPlan.Interval,
		Description:    subscriptionPlan.Description,
		IsActive:       subscriptionPlan.IsActive,
	}, nil
}

func (as *AuthSubscriptionUsecase) DeactivateSubscriptionPlan(deactivateSubscriptionPlanReq requestmodels.DeactivateSubscriptionPlanRequest) (responsemodels.DeactivateSubscriptionPlanResponse, error) {
	status, err := as.AuthSubscriptionRepository.FetchStatusFromSubcriptionPlan(deactivateSubscriptionPlanReq.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.DeactivateSubscriptionPlanResponse{}, domain.ErrSubPlanNotFound
		}
		return responsemodels.DeactivateSubscriptionPlanResponse{}, fmt.Errorf("database error: %w", err)
	}
	if !status {
		return responsemodels.DeactivateSubscriptionPlanResponse{}, domain.ErrSubscriptionPlanAlreadyDeactive
	}
	subscriptionPlan, err := as.AuthSubscriptionRepository.DeactivateSubscriptionPlan(deactivateSubscriptionPlanReq)
	if err != nil {
		return responsemodels.DeactivateSubscriptionPlanResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.DeactivateSubscriptionPlanResponse{
		ID:             subscriptionPlan.ID,
		CreatedAt:      subscriptionPlan.CreatedAt,
		UpdatedAt:      subscriptionPlan.UpdatedAt,
		RazorpayPlanId: subscriptionPlan.RazorpayPlanId,
		Name:           subscriptionPlan.Name,
		Price:          subscriptionPlan.Price,
		Currency:       subscriptionPlan.Currency,
		Period:         subscriptionPlan.Period,
		Interval:       subscriptionPlan.Interval,
		Description:    subscriptionPlan.Description,
		IsActive:       subscriptionPlan.IsActive,
	}, nil
}

func (as *AuthSubscriptionUsecase) GetAllSubscriptionPlans(getAllSubscripionPlansReq requestmodels.GetAllSubscriptionPlansRequest) (responsemodels.GetAllSubscriptionPlansResponse, error) {
	subscriptionPlans, err := as.AuthSubscriptionRepository.GetAllSubscriptionPlans(getAllSubscripionPlansReq)
	if err != nil {
		return responsemodels.GetAllSubscriptionPlansResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.GetAllSubscriptionPlansResponse{
		SubscriptionPlans: subscriptionPlans.SubscriptionPlans,
	}, nil
}

func (as *AuthSubscriptionUsecase) GetAllActiveSubscriptionPlans(getAllActiveSubscriptionPlansReq requestmodels.GetAllActiveSubscriptionPlansRequest) (responsemodels.GetAllActiveSubscriptionPlansResponse, error) {
	subscriptionPlans, err := as.AuthSubscriptionRepository.GetAllActiveSubscriptionPlans(getAllActiveSubscriptionPlansReq)
	if err != nil {
		return responsemodels.GetAllActiveSubscriptionPlansResponse{}, fmt.Errorf("database error: %w", err)
	}
	return responsemodels.GetAllActiveSubscriptionPlansResponse{
		SubscriptionPlans: subscriptionPlans.SubscriptionPlans,
	}, nil
}

func (as *AuthSubscriptionUsecase) Subscribe(subscribeReq requestmodels.SubscribeRequest) (responsemodels.SubscribeResponse, error) {
	razorpayPlanId, err := as.AuthSubscriptionRepository.FetchRazorpayPlanIdFromId(subscribeReq.PlanId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.SubscribeResponse{}, domain.ErrSubPlanNotFound
		}
		return responsemodels.SubscribeResponse{}, fmt.Errorf("plan lookup failed: %w", err)
	}
	//fmt.Println("RazorpayPlanId", razorpayPlanId)

	eligible, err := as.AuthSubscriptionRepository.IsEligibleForSubsciption(subscribeReq)
	if err != nil {
		log.Println(err)
		return responsemodels.SubscribeResponse{}, err
	}
	if !eligible {
		return responsemodels.SubscribeResponse{}, domain.ErrNotEligible
	}

	userDetail, err := as.AuthSubscriptionRepository.FetchUserPublicData(subscribeReq.UserId)
	if err != nil {
		return responsemodels.SubscribeResponse{}, fmt.Errorf("database error: %w", err)
	}
	//fmt.Println("userDetail", userDetail)
	//return responsemodels.SubscribeResponse{},nil
	if userDetail.RazorpayCustomerID == "" {
		userDetail.RazorpayCustomerID, err = as.createAndSaveCustomer(userDetail, subscribeReq.UserEmail)
		if err != nil {
			return responsemodels.SubscribeResponse{}, err
		}
	}

	//razorpayClient:=utils.NewRazorpayClient(as.RazorpayCredentials.KeyId,as.RazorpayCredentials.KeySecret)
	subscriptionData := map[string]interface{}{
		"plan_id":     razorpayPlanId,
		"total_count": subscribeReq.TotalCount,
		"customer_id": userDetail.RazorpayCustomerID, // Use the ID instead of the "customer" map
		//"quantity":        1,
		//"customer_notify": 1,

		//before adding the below code, first we will have to create the cusomer and store customer_id i user_subsciptions
		// ADD THIS SECTION:
		// "customer": map[string]interface{}{
		// 	"name":  userDetail.UserName,
		// 	"email": subscribeReq.UserEmail,
		// 	"contact": "1234567890", // Provide a dummy or real number
		// },

		"notes": map[string]interface{}{
			"email":     subscribeReq.UserEmail, // Storing the email in Razorpay metadata
			"user_id":   subscribeReq.UserId,    // Good practice to store UserID too
			"user_name": userDetail.UserName,
		},
	}
	subscription, err := as.RazorpayGateway.CreateSubscription(subscriptionData)
	if err != nil {
		log.Println("error on subscribing", err)
		return responsemodels.SubscribeResponse{}, fmt.Errorf("provider error: %w", err)
	}
	//fmt.Println(subscription)
	subcribeRes, err := as.AuthSubscriptionRepository.CreateSubscription(subscribeReq, subscription)
	if err != nil {
		log.Printf("is there any error returning after createSubscripion %v", err)
		return responsemodels.SubscribeResponse{}, err
	}
	//fmt.Println("subscribeRes",subcribeRes)
	return subcribeRes, nil
}

func (as *AuthSubscriptionUsecase) createAndSaveCustomer(userDetail responsemodels.UserPublicDataResponse, email string) (string, error) {
	// 1. Prepare Customer Data for Razorpay
	customerParams := map[string]interface{}{
		"name":          userDetail.UserName,
		"email":         email,
		"contact":       userDetail.Phone, // Ensure this is in your user model
		"fail_existing": 0,                // 0 means if email exists, it returns the existing customer instead of erroring
	}
	// 2. Call Razorpay API
	// Assuming as.RazorpayGateway is your initialized razorpay client
	res, err := as.RazorpayGateway.Client.Customer.Create(customerParams, nil)
	if err != nil {
		log.Printf("Failed to create Razorpay customer: %v", err)
		return "", err
	}
	// 3. Extract the ID from the response
	// res is usually a map[string]interface{} or a specific struct depending on your SDK wrapper
	customerID, ok := res["id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format from Razorpay")
	}

	// 4. Update your local database
	// This links your internal UserID to the Razorpay CustomerID forever
	err = as.AuthSubscriptionRepository.UpdateUserRazorpayCustomerID(userDetail.UserID, customerID)
	if err != nil {
		log.Printf("Failed to save customerID %s to DB for user %d: %v", customerID, userDetail.UserID, err)
		if err == gorm.ErrRecordNotFound {
			return "", domain.ErrUserNotFound
		}
		return "", err
	}

	return customerID, nil
}

// func (as *AuthSubscriptionUsecase) VerifySubscriptionPayment(verifySubscriptionPaymentReq requestmodels.VerifySubscriptionPaymentRequest) (responsemodels.VerifySubscriptionPaymentResponse, error) {
// 	var subscriptionRes responsemodels.VerifySubscriptionPaymentResponse
// 	//razorpayClient:=utils.NewRazorpayClient(as.RazorpayCredentials.KeyId,as.RazorpayCredentials.KeySecret)
// 	subscription, err := as.RazorpayGateway.Client.Subscription.Fetch(verifySubscriptionPaymentReq.RazorpaySubscriptionId, nil, nil)
// 	//fmt.Println("------------------------")
// 	//fmt.Println("subscription",subscription)
// 	if err != nil {
// 		return responsemodels.VerifySubscriptionPaymentResponse{}, err
// 	}
// 	startAt, ok := subscription["start_at"].(float64)
// 	fmt.Println("print value start at", startAt)
// 	if !ok {
// 		//fmt.Println("what if its coming here *******")
// 		planId, err := as.AuthSubscriptionRepository.FetchRazorpayPlanIdFromRazrorpaySubscriptionId(verifySubscriptionPaymentReq.RazorpaySubscriptionId)
// 		if err != nil {
// 			return responsemodels.VerifySubscriptionPaymentResponse{}, fmt.Errorf("database error :%w", err)
// 		}
// 		period, interval, err := as.AuthSubscriptionRepository.FetchIntervalPeriodFromSubscriptionPlan(planId)
// 		if err != nil {
// 			return responsemodels.VerifySubscriptionPaymentResponse{}, fmt.Errorf("database error: %w", err)
// 		}
// 		totalCount, err := as.AuthSubscriptionRepository.FetchTotalCountFromUserSubscription(verifySubscriptionPaymentReq.RazorpaySubscriptionId)
// 		if err != nil {
// 			return responsemodels.VerifySubscriptionPaymentResponse{}, fmt.Errorf("database error: %w", err)
// 		}
// 		startAt := time.Now()
// 		// Calculate the end_at and NextChargeAt times
// 		endAt, nextChargeAt := calculateEndAndNextChargeTime(startAt, period, interval, totalCount)
// 		//fmt.Println("print inside", startAt, endAt, nextChargeAt)
// 		subscriptionRes, err = as.AuthSubscriptionRepository.UpdateTimeUserSubscription(startAt, endAt, nextChargeAt, verifySubscriptionPaymentReq.RazorpaySubscriptionId)
// 		if err != nil {
// 			return responsemodels.VerifySubscriptionPaymentResponse{}, fmt.Errorf("database error: %w", err)
// 		}
// 		go as.pollRazorpayAndSync(verifySubscriptionPaymentReq.RazorpaySubscriptionId)
// 	} else {
// 		//fmt.Println("hi i hope its here------")
// 		subscriptionRes, err = as.AuthSubscriptionRepository.UpdateUserSubscripion(verifySubscriptionPaymentReq.RazorpaySubscriptionId, subscription)
// 		if err != nil {
// 			return responsemodels.VerifySubscriptionPaymentResponse{}, err
// 		}
// 	}

// 	userid, err := as.AuthSubscriptionRepository.FetchUserIdFromSubscriptionId(verifySubscriptionPaymentReq.RazorpaySubscriptionId)
// 	if err != nil {
// 		return responsemodels.VerifySubscriptionPaymentResponse{}, err
// 	}
// 	err = as.AuthSubscriptionRepository.TurnBlueTickTrueForUserId(userid)
// 	if err != nil {
// 		return responsemodels.VerifySubscriptionPaymentResponse{}, err
// 	}
// 	payment, err := as.RazorpayGateway.Client.Payment.Fetch(verifySubscriptionPaymentReq.RazorpayPaymentId, nil, nil)
// 	if err != nil {
// 		return responsemodels.VerifySubscriptionPaymentResponse{}, err
// 	}
// 	//fmt.Println("payment : ",payment)
// 	_, err = as.AuthSubscriptionRepository.PopulatePayment(payment, verifySubscriptionPaymentReq)
// 	if err != nil {
// 		return responsemodels.VerifySubscriptionPaymentResponse{}, nil
// 	}
// 	//fmt.Println("payment table",paymentRes)
// 	fmt.Println("just before into service ", subscriptionRes.StartAt, subscriptionRes.NextChargeAt)
// 	return subscriptionRes, nil
// }

func (as *AuthSubscriptionUsecase) Unsubscribe(unsubscribeReq requestmodels.UnsubscribeRequest) (responsemodels.UnsubscribeResponse, error) {
	razorpaySubscritpionId, err := as.AuthSubscriptionRepository.FetchRazorpaySubscriptionIdFromUserId(unsubscribeReq.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.UnsubscribeResponse{}, domain.ErrNoActiveSubscription
		}
		return responsemodels.UnsubscribeResponse{}, fmt.Errorf("%w: %v:", domain.ErrDatabase, err)
	}
	unsubscribeReq.RazorpaySubId = razorpaySubscritpionId
	//fmt.Println("is it actually nil,???", unsubscribeReq.RazorpaySubId)

	status, err := as.AuthSubscriptionRepository.FetchUserSubscription(unsubscribeReq.RazorpaySubId)
	if err != nil {
		log.Println(err)
		return responsemodels.UnsubscribeResponse{}, fmt.Errorf("%w: %v: %v:", domain.ErrDatabase, err, "fetching subscription from database failed")
	}
	if status == "completed" {
		return responsemodels.UnsubscribeResponse{}, domain.ErrSubCompleted
	}
	if status == "cancelled" {
		return responsemodels.UnsubscribeResponse{}, domain.ErrSubCancelled
	}
	data := map[string]interface{}{
		"cancel_at_cycle_end": unsubscribeReq.CancelAtCycleEnd,
	}

	_, err = as.RazorpayGateway.Client.Subscription.Cancel(razorpaySubscritpionId, data, nil)
	if err != nil {
		log.Println("print the error on cancellation razorpay api call", err)
		return responsemodels.UnsubscribeResponse{},
			fmt.Errorf("%w: %v", domain.ErrRazorpayCancel, err)
	}

	unsubscibeRes, err := as.AuthSubscriptionRepository.SetCancelReason(unsubscribeReq)
	if err != nil {
		return responsemodels.UnsubscribeResponse{}, fmt.Errorf("%w: %v: %v:", domain.ErrDatabase, err, "cancel reason failed to update in db")
	}
	// userid, err := as.AuthSubscriptionRepository.FetchUserIdFromSubscriptionId(razorpaySubscritpionId)
	// if err != nil {
	// 	return responsemodels.UnsubscribeResponse{}, err
	// }
	// nextChargeAt, err := as.AuthSubscriptionRepository.FetchNextChargeAtFromUserSubcription(razorpaySubscritpionId)
	// if err != nil {
	// 	return responsemodels.UnsubscribeResponse{}, err
	// }
	// delay := time.Until(nextChargeAt)
	// go func() {
	// 	<-time.After(delay)
	// 	err := as.AuthSubscriptionRepository.TurnOffBlueTickForUserId(userid)
	// 	if err != nil {
	// 		fmt.Println("error while turning off blue tick", err)
	// 	}
	// }()
	return unsubscibeRes, nil
}

func (as *AuthSubscriptionUsecase) SetProfileImage(setProfileImageReq requestmodels.SetProfileImageRequest) (responsemodels.SetProfileImageResponse, error) {
	ct := setProfileImageReq.ContentType

	ct = strings.TrimPrefix(ct, "image/")

	filename := fmt.Sprintf("%d_%d.%s", setProfileImageReq.UserId, time.Now().Unix(), ct)
	//fmt.Println("file name", filename)
	key := "profiles/" + filename
	//fmt.Println("inside usecase type", setProfileImageReq.ContentType)
	if setProfileImageReq.ContentType == "" {
		return responsemodels.SetProfileImageResponse{}, fmt.Errorf("content type is nil")
	}
	//fmt.Println("hi hello",aws.String(as.AwsBucket),aws.String(key),as.AwsBucket,key)
	uploader := manager.NewUploader(as.AwsS3Client)
	_, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(as.AwsBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(setProfileImageReq.Image),
		ContentType: aws.String(setProfileImageReq.ContentType),
	})
	if err != nil {
		//fmt.Println("is it here")
		return responsemodels.SetProfileImageResponse{}, status.Errorf(codes.Internal, "upload failed: %v", err)
	}
	// Construct URL
	imageURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", as.AwsBucket, as.Config.Aws.AwsRegion, key)
	// Save to DB
	err = as.AuthSubscriptionRepository.UpdateProfileImage(setProfileImageReq.UserId, imageURL)
	if err != nil {
		return responsemodels.SetProfileImageResponse{}, status.Errorf(codes.Internal, "db update failed: %v", err)
	}

	return responsemodels.SetProfileImageResponse{
		ImageUrl: imageURL,
	}, nil
}

func (as *AuthSubscriptionUsecase) CheckUserExists(userId uint64) (bool, error) {
	status, err := as.AuthSubscriptionRepository.CheckUserExistsById(userId)
	if err != nil {
		return false, err
	}
	if !status {
		return false, domain.ErrUserNotFound
	}
	return status, err
}

func (as *AuthSubscriptionUsecase) CheckAllUsersExists(users []uint64) ([]uint64, error) {
	resp, err := as.AuthSubscriptionRepository.CheckAllUsersExists(users)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (as *AuthSubscriptionUsecase) GetProfileInformation(req requestmodels.GetProfileInformationRequest) (responsemodels.GetProfileInformationResponse, error) {
	resp, err := as.AuthSubscriptionRepository.GetProfileInformation(req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.GetProfileInformationResponse{}, domain.ErrUserNotFound
		}
		return responsemodels.GetProfileInformationResponse{}, err
	}
	//fmt.Println("resp in usecase", resp)
	return resp, nil
}
func (as *AuthSubscriptionUsecase) EditProfileInformation(userId uint64, updateData map[string]interface{}) (responsemodels.EditProfile, error) {
	resp, err := as.AuthSubscriptionRepository.EditProfileInformation(userId, updateData)
	if err != nil {
		return responsemodels.EditProfile{}, err
	}
	return resp, nil
}
func (as *AuthSubscriptionUsecase) ChangePassword(req requestmodels.ChangePassword) (responsemodels.ChangePasswordResponse, error) {
	password, err := as.AuthSubscriptionRepository.FetchHashedPassword(req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.ChangePasswordResponse{}, domain.ErrUserNotFound
		}
		return responsemodels.ChangePasswordResponse{}, err
	}
	err = utils.CompareWithHashedPassword(password, req.OldPassword)
	if err != nil {
		return responsemodels.ChangePasswordResponse{}, err
	}
	passwordHash := utils.HashPassword(req.ConfirmNewPassword)
	resp, err := as.AuthSubscriptionRepository.ChangePassword(req, passwordHash)
	if err != nil {
		return responsemodels.ChangePasswordResponse{}, err
	}
	return resp, nil
}
func (as *AuthSubscriptionUsecase) SearchUser(req requestmodels.SearchUser) (responsemodels.SearchUserResponse, error) {
	resp, err := as.AuthSubscriptionRepository.SearchUser(req)
	if err != nil {
		return responsemodels.SearchUserResponse{}, err
	}
	return resp, nil
}

func (as *AuthSubscriptionUsecase) FetchUserPublicData(userid uint64) (responsemodels.UserPublicDataResponse, error) {
	resp, err := as.AuthSubscriptionRepository.FetchUserPublicData(userid)
	if err != nil {
		return responsemodels.UserPublicDataResponse{}, err
	}
	return resp, nil
}
func (as *AuthSubscriptionUsecase) FetchUserMetaData(userids []uint64) (map[uint64]responsemodels.UserMetaData, error) {
	resp, err := as.AuthSubscriptionRepository.FetchUserMetaData(userids)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return resp, nil
}

func (as *AuthSubscriptionUsecase) CheckUserListExists(userids []uint64) ([]uint64, error) {
	resp, err := as.AuthSubscriptionRepository.CheckUserListExists(userids)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNoUsersFound
		}
		return nil, err
	}
	return resp, nil
}

func (as *AuthSubscriptionUsecase) GetSubscriptionDetails(req requestmodels.GetSubscriptionDetails) (responsemodels.GetSubscriptionDetails, error) {
	resp, err := as.AuthSubscriptionRepository.GetSubscriptionDetails(req)
	if err != nil {
		log.Println(err)
		if err == gorm.ErrRecordNotFound {
			return responsemodels.GetSubscriptionDetails{}, domain.ErrNoSubscription
		}
		return responsemodels.GetSubscriptionDetails{}, err
	}
	return resp, nil
}

func (as *AuthSubscriptionUsecase) WebhookSubscriptionActivated(req requestmodels.WebhookSubscriptionActivatedRequest) (responsemodels.WebhookSubscriptionActivatedResponse, error) {
	status, err := as.AuthSubscriptionRepository.FetchSubStatus(req.RazorpaySubscriptionId)
	if err != nil {
		log.Println(err)
		return responsemodels.WebhookSubscriptionActivatedResponse{}, err
	}
	if status == "completed" || status == "cancelled" {
		return responsemodels.WebhookSubscriptionActivatedResponse{}, errors.New("subscription already in cancelled or completed state")
	}
	if req.Status == "active" {
		err := as.AuthSubscriptionRepository.TurnBlueTickTrueForUserId(req.UserID)
		if err != nil {
			log.Printf("failed to trun on blue tick for user id %d\n", req.UserID)
		}
	}
	resp, err := as.AuthSubscriptionRepository.UpddateActivatedSubscription(req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.WebhookSubscriptionActivatedResponse{}, domain.RazorpaySubscriptionIdNotFound
		}
		return responsemodels.WebhookSubscriptionActivatedResponse{}, err
	}
	return resp, nil
}

func (as *AuthSubscriptionUsecase) WebhookSubscriptionCharged(req requestmodels.WebhookSubscriptionChargedRequest) (responsemodels.WebhookSubscriptionChargedResponse, error) {
	status, err := as.AuthSubscriptionRepository.FetchSubStatus(req.RazorpaySubscriptionId)
	if err != nil {
		log.Println(err)
		return responsemodels.WebhookSubscriptionChargedResponse{}, err
	}
	//fmt.Println("is it coming here",req.RazorpaySubscriptionId)
	err = as.AuthSubscriptionRepository.UpdateNextChargeAt(req.NextChargeAt, req.RazorpaySubscriptionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.WebhookSubscriptionChargedResponse{}, domain.RazorpaySubscriptionIdNotFound
		}
	}
	err = as.AuthSubscriptionRepository.UpdateCount(req)

	resp, err := as.AuthSubscriptionRepository.UpdatePayment(req)
	if err != nil {
		return responsemodels.WebhookSubscriptionChargedResponse{}, err
	}

	if status == "completed" || status == "cancelled" {
		return responsemodels.WebhookSubscriptionChargedResponse{}, errors.New("subscription already in cancelled or completed state")
	}

	err = as.AuthSubscriptionRepository.UpdateStatusToActive(req.RazorpaySubscriptionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.WebhookSubscriptionChargedResponse{}, domain.RazorpaySubscriptionIdNotFound
		}
	}

	//fmt.Println("request in charged",req)
	//if req.Status=="active"{
	err = as.AuthSubscriptionRepository.TurnBlueTickTrueForUserId(req.UserID)
	if err != nil {
		log.Printf("failed to trun on blue tick for user id %d\n", req.UserID)
	}
	//}

	return resp, nil
}
func (as *AuthSubscriptionUsecase) WebhookSubscriptionHalted(req requestmodels.WebhookSubscriptionHaltedRequest) (responsemodels.WebhookSubscriptionHaltedResponse, error) {
	status, err := as.AuthSubscriptionRepository.FetchSubStatus(req.RazorpaySubscriptionId)
	if err != nil {
		log.Println(err)
		return responsemodels.WebhookSubscriptionHaltedResponse{}, err
	}
	if status == "completed" || status == "cancelled" {
		return responsemodels.WebhookSubscriptionHaltedResponse{}, errors.New("subscription already in cancelled or completed state")
	}
	err = as.AuthSubscriptionRepository.UpdateStatusHalted(req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.WebhookSubscriptionHaltedResponse{}, domain.RazorpaySubscriptionIdNotFound
		}
		return responsemodels.WebhookSubscriptionHaltedResponse{}, err
	}
	err = as.AuthSubscriptionRepository.TurnOffBlueTickForUserId(req.UserId)
	if err != nil {
		return responsemodels.WebhookSubscriptionHaltedResponse{}, err
	}
	return responsemodels.WebhookSubscriptionHaltedResponse{
		RazorpaySubcriptionId: req.RazorpaySubscriptionId,
	}, nil
}

func (as *AuthSubscriptionUsecase) WebhookSubscriptionCancelled(req requestmodels.WebhookSubscriptionCancelledRequest) (responsemodels.WebhookSubscriptionCancelledResponse, error) {
	status, err := as.AuthSubscriptionRepository.FetchSubStatus(req.RazorpaySubscriptionId)
	if err != nil {
		log.Println(err)
		return responsemodels.WebhookSubscriptionCancelledResponse{}, err
	}
	if status == "completed" || status == "cancelled" {
		return responsemodels.WebhookSubscriptionCancelledResponse{}, errors.New("subscription already in cancelled or completed state")
	}
	resp, err := as.AuthSubscriptionRepository.UpdateSubscriptionCancelled(req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.WebhookSubscriptionCancelledResponse{}, domain.RazorpaySubscriptionIdNotFound
		}
		return responsemodels.WebhookSubscriptionCancelledResponse{}, err
	}
	err = as.AuthSubscriptionRepository.TurnOffBlueTickForUserId(req.UserId)
	if err != nil {
		log.Printf("failed to turn off blue tick for user id:%d\n", req.UserId)
		return responsemodels.WebhookSubscriptionCancelledResponse{}, err
	}
	return resp, nil
}

func (as *AuthSubscriptionUsecase) WebhookSubscriptionCompleted(req requestmodels.WebhookSubscriptionCompletedRequest) (responsemodels.WebhookSubscriptionCompletedResponse, error) {
	//	fmt.Println("why not coming in completed")
	//fmt.Println("complete usecase req",req)
	resp, err := as.AuthSubscriptionRepository.UpdateSubscripionCompleted(req)
	if err != nil {
		log.Println("error while updating to completed", err)
		if err == gorm.ErrRecordNotFound {
			return responsemodels.WebhookSubscriptionCompletedResponse{}, domain.RazorpaySubscriptionIdNotFound
		}
		return responsemodels.WebhookSubscriptionCompletedResponse{}, err
	}
	err = as.AuthSubscriptionRepository.TurnOffBlueTickForUserId(req.UserId)
	if err != nil {
		log.Printf("failed to turn off blue tick for user id:%d\n", req.UserId)
		return responsemodels.WebhookSubscriptionCompletedResponse{}, err
	}
	return resp, nil
}

func (as *AuthSubscriptionUsecase) Webhook(webhookReq requestmodels.RazorpayEvent) (responsemodels.WebhookResponse, error) {
	// data := map[string]interface{}{
	// 	"remaining_count": 12,
	// 	"customer_notify": 1,
	// }
	// updatedSubscriptionData, err := as.RazorpayClient.Subscription.Update(webhookReq.AccountID, data, nil)
	// if err != nil {
	// 	fmt.Println("calling update subscripton failed", err)
	// 	return responsemodels.WebhookResponse{}, err
	// }
	//fmt.Println("updated Subscription Data is ", updatedSubscriptionData)

	resp, err := as.SmtpProvider.SendNotificationEmailForResubscribing(webhookReq)
	if err != nil {
		return responsemodels.WebhookResponse{}, err
	}
	return resp, nil
}

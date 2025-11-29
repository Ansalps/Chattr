package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/pb"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/usecase"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/usecase/interfacesUsecase"
	//"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/repository/interfaces"
)

type AuthSubscriptionServer struct {
	AuthSubscriptionUsecase interfacesUsecase.AuthSubscriptionUsecase
	pb.AuthSubscriptionServiceServer
}

func NewAuthSubscriptionServer(useCase interfacesUsecase.AuthSubscriptionUsecase) *AuthSubscriptionServer {
	return &AuthSubscriptionServer{
		AuthSubscriptionUsecase: useCase,
	}
}

func (as *AuthSubscriptionServer) AdminLogin(ctx context.Context, req *pb.AdminLoginRequest) (*pb.AdminLoginResponse, error) {
	adminLogin := requestmodels.AdminLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	admin, err := as.AuthSubscriptionUsecase.AdminLogin(adminLogin)
	if err != nil {
		log.Printf("AdminLogin failed for email=%s: %v", req.Email, err)
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, usecase.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		default:
			return nil, status.Error(codes.Internal, "interanal server error")
		}
	}
	adminDetails := &pb.AdminDetails{
		Id:    uint64(admin.Admin.ID),
		Email: admin.Admin.Email,
	}
	fmt.Println("adminDetails", adminDetails)
	fmt.Println("adminToken", admin.AccessToken)
	return &pb.AdminLoginResponse{
		AdminDetails: adminDetails,
		AccessToken:  admin.AccessToken,
		RefreshToken: admin.RefreshToken,
	}, nil
}

func (as *AuthSubscriptionServer) BlockUser(ctx context.Context, req *pb.BlockUserRequest) (*pb.BlockUserResponse, error) {
	blockUserReq := requestmodels.BlockUserRequest{
		UserId: req.UserId,
	}
	blockUserResponse, err := as.AuthSubscriptionUsecase.BlockUser(blockUserReq)
	if err != nil {
		log.Printf("Block user failed for user id %d : %v", blockUserReq.UserId, err)
		switch {
		case errors.Is(err, usecase.ErrUserNotActive):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		case errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Errorf(codes.NotFound, "user not found for user id: %v", blockUserReq.UserId)
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &pb.BlockUserResponse{
		UserId: blockUserResponse.UserId,
	}, nil
}

func (as *AuthSubscriptionServer) UnblockUser(ctx context.Context, req *pb.UnblockUserRequest) (*pb.UnblockUserResponse, error) {
	unblockUserReq := requestmodels.UnblockUserRequest{
		UserId: req.UserId,
	}
	unblockUserResponse, err := as.AuthSubscriptionUsecase.UnblockUser(unblockUserReq)
	if err != nil {
		log.Printf("Unblock user failed for user id %d : %v", unblockUserReq.UserId, err)
		switch {
		case errors.Is(err, usecase.ErrUserNotBlocked):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		case errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Errorf(codes.NotFound, "user not found for user id: %v", unblockUserReq.UserId)
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &pb.UnblockUserResponse{
		UserId: unblockUserResponse.UserId,
	}, nil
}

func (as *AuthSubscriptionServer) GetAllUsers(ctx context.Context, req *pb.GetAllUsersRequest) (*pb.GetAllUsersResponse, error) {
	getAllUsersReq := requestmodels.GetAllUsersRequest{
		Limit:  req.Limit,
		Offset: req.Offset,
	}
	users, err := as.AuthSubscriptionUsecase.GetAllUsers(getAllUsersReq)
	if err != nil {
		log.Printf("Get All Users failed : %v", err)
		switch {
		default:
			return nil, status.Error(codes.Internal, "interanal server error")
		}
	}
	pbUsers := make([]*pb.User, len(users.Users))
	for i, user := range users.Users {
		pbUsers[i] = &pb.User{
			Id:            user.ID,
			Name:          user.Name,
			UserName:      user.UserName,
			Email:         user.Email,
			Bio:           user.Bio,
			ProfileImgUrl: user.ProfileImgUrl,
			Links:         user.Links,
			Status:        user.Status,
		}
	}
	return &pb.GetAllUsersResponse{
		Users: pbUsers,
	}, nil
}

func (as *AuthSubscriptionServer) UserSignUp(ctx context.Context, req *pb.UserSignUpRequest) (*pb.UserSignUpResponse, error) {
	userSignup := requestmodels.UserSignUpRequest{
		UserName:        req.UserName,
		Name:            req.Name,
		Email:           req.Email,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	}
	userResponse, err := as.AuthSubscriptionUsecase.UserSignUp(userSignup)
	if err != nil {
		log.Printf("UsersignUp failed for email=%s and username=%s: %v", req.Email, req.UserName, err)
		switch {
		case errors.Is(err, usecase.ErrUserAlreadyExistsByEmail):
			return nil, status.Errorf(codes.AlreadyExists, "user with email=%s already exist", req.Email)
		case errors.Is(err, usecase.ErrUserAlreadyExistsByUsername):
			return nil, status.Errorf(codes.AlreadyExists, "username %s is already taken", req.UserName)
		default:
			return nil, status.Error(codes.Internal, "interanal server error")
		}

	}
	return &pb.UserSignUpResponse{
		Id:                   uint64(userResponse.ID),
		UserName:             userResponse.UserName,
		Name:                 userResponse.Name,
		Email:                userResponse.Email,
		OtpVerificationToken: userResponse.OtpVerificationToken,
	}, nil

}

func (as *AuthSubscriptionServer) VerifyOtp(ctx context.Context, req *pb.OtpRequest) (*pb.OtpVerificationResponse, error) {
	otpReq := requestmodels.OtpRequest{
		UserId: req.UserId,
		OtpCode: req.OtpCode,
		Email:   req.Email,
		Purpose: req.Purpose,
	}
	otpResponse, err := as.AuthSubscriptionUsecase.VerifyOtp(otpReq)
	if err != nil {
		log.Printf("OTP verification failed for email %s (reason: mismatch): %v", otpReq.Email, err)
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Errorf(codes.NotFound, "user not found")
		case errors.Is(err, usecase.ErrInvalidCredentials):
			return nil, status.Errorf(codes.InvalidArgument, "invalid otp")
		case errors.Is(err, usecase.ErrOtpExpired):
			return nil, status.Error(codes.FailedPrecondition, "otp expired")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	fmt.Println("in service Access Token",otpResponse.AccessToken)
	fmt.Println("in service Refresh Token",otpResponse.RefreshToken)
	fmt.Println("in service Temp Token",otpResponse.TempToken)
	return &pb.OtpVerificationResponse{
		Email:  otpResponse.Email,
		Status: otpResponse.Status,
		AccessToken: otpResponse.AccessToken,
		RefreshToken: otpResponse.RefreshToken,
		TempToken: otpResponse.TempToken,
	}, nil
}

func (as *AuthSubscriptionServer) ResendOtp(ctx context.Context, req *pb.ResendOtpRequest) (*pb.ResendOtpResponse, error) {
	resendOtpReq := requestmodels.ResendOtpRequest{
		Email: req.Email,
	}
	resendOtpResponse, err := as.AuthSubscriptionUsecase.ResendOtp(resendOtpReq)
	if err != nil {
		log.Printf("Resend otp failed for email %s : %v", resendOtpReq.Email, err)
		switch {
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &pb.ResendOtpResponse{
		Email: resendOtpResponse.Email,
	}, nil
}

func (as *AuthSubscriptionServer) AccessRegenerator(ctx context.Context, req *pb.AccessRegeneratorRequest) (*pb.AccessRegeneratorResponse, error) {
	accessRegeneratorReq := requestmodels.AccessRegeneratorRequest{
		ID:    req.Id,
		Email: req.Email,
		Role:  req.Role,
	}
	accessRegeneratorResponse, err := as.AuthSubscriptionUsecase.AccessRegenerator(accessRegeneratorReq)
	if err != nil {
		log.Printf("new access regeneration failed for email %s : %v", accessRegeneratorReq.Email, err)
		switch {
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &pb.AccessRegeneratorResponse{
		Id:             accessRegeneratorResponse.Id,
		Email:          accessRegeneratorResponse.Email,
		Role:           accessRegeneratorResponse.Role,
		NewAccessToken: accessRegeneratorResponse.NewAccessToken,
	}, nil
}
func (as *AuthSubscriptionServer)ForgetPassword(ctx context.Context,req *pb.ForgotPasswordReqeust)(*pb.ForgotPasswordResponse,error){
	forgotPasswordReq:=requestmodels.ForgotPasswordRequest{
		Email: req.Email,
	}
	forgotPasswordRes,err:=as.AuthSubscriptionUsecase.ForgotPassword(forgotPasswordReq)
	if err!=nil{
		log.Printf("OTP Forgot Password failed for email %s (reason: mismatch): %v", forgotPasswordReq.Email, err)
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Errorf(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &pb.ForgotPasswordResponse{
		Email: forgotPasswordRes.Email,
		TempToken: forgotPasswordRes.TempToken,
	},nil
}
func (as *AuthSubscriptionServer) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	resetPasswordReq := requestmodels.ResetPasswordRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	resetPasswordResponse, err := as.AuthSubscriptionUsecase.ResetPassword(resetPasswordReq)
	if err != nil {
		log.Printf("Reset password failed for email %s : %v", resetPasswordReq.Email, err)
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Errorf(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &pb.ResetPasswordResponse{
		Email: resetPasswordResponse.Email,
	}, nil
}

func (as *AuthSubscriptionServer) UserLogin(ctx context.Context, req *pb.UserLoginRequest) (*pb.UserLoginResponse, error) {
	userLoginReq := requestmodels.UserLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	user, err := as.AuthSubscriptionUsecase.UserLogin(userLoginReq)
	if err != nil {
		log.Printf("User Login failed for email=%s: %v", req.Email, err)
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, usecase.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case errors.Is(err, usecase.ErrBlockedLogin):
			return nil, status.Error(codes.PermissionDenied, "user account is blocked by admin, cannot login")
		case errors.Is(err, usecase.ErrPendingLogin):
			return nil, status.Error(codes.FailedPrecondition, "email verification pending")
		default:
			return nil, status.Error(codes.Internal, "interanal server error")
		}
	}
	userDetails := &pb.UserDetails{
		Id:       user.User.Id,
		Name:     user.User.Name,
		UserName: user.User.UserName,
		Email:    user.User.Email,
		Status:   user.User.Status,
		BlueTick: user.User.BlueTick,
	}
	return &pb.UserLoginResponse{
		UserDetails:  userDetails,
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
	}, nil
}

func (as *AuthSubscriptionServer) CreateSubscriptionPlan(ctx context.Context, req *pb.CreateSubscriptionPlanRequest) (*pb.CreateSubscriptionPlanResponse, error) {
	createSubscriptionPlanReq := requestmodels.CreateSubscriptionPlanRequest{
		Name:         req.Name,
		Price:        req.Price,
		Currency: req.Currency,
		Period: req.Period,
		Interval: req.Interval,
		Description:  req.Description,
	}
	createSubscriptionPlanResponse, err := as.AuthSubscriptionUsecase.CreateSubscriptionPlan(createSubscriptionPlanReq)
	if err != nil {
		log.Printf("Create Subscription paln failed for subscription paln =%s: %v", req.Name, err)
		switch {
		default:
			return nil, status.Error(codes.Internal, "interanal server error")
		}
	}
	return &pb.CreateSubscriptionPlanResponse{
		Id:           createSubscriptionPlanResponse.ID,
		CreatedAt:    utils.ToProtoTimestamp(createSubscriptionPlanResponse.CreatedAt),
		UpdatedAt:    utils.ToProtoTimestamp(createSubscriptionPlanResponse.UpdatedAt),
		Name:         createSubscriptionPlanResponse.Name,
		Price:        createSubscriptionPlanResponse.Price,
		Currency: createSubscriptionPlanReq.Currency,
		Period: createSubscriptionPlanReq.Period,
		Interval: createSubscriptionPlanResponse.Interval,
		Description:  createSubscriptionPlanResponse.Description,
		IsActive:     createSubscriptionPlanResponse.IsActive,
	}, nil
}



func (as *AuthSubscriptionServer) ActivateSubscriptionPlan(ctx context.Context, req *pb.ActivateSubscriptionPlanRequest) (*pb.ActivateSubscriptionPlanResponse, error) {
	activateSubscriptionPlanReq := requestmodels.ActivateSubscriptionPlanRequest{
		ID: req.Id,
	}
	activateSubscriptionPlanResponse, err := as.AuthSubscriptionUsecase.ActivateSubscriptionPlan(activateSubscriptionPlanReq)
	if err != nil {
		log.Printf("Actvate Subscription paln failed for subscription paln with id =%d: %v", req.Id, err)
		switch {
		case errors.Is(err,usecase.ErrSubscriptionPlanAlreadyActive):
			return nil,status.Error(codes.FailedPrecondition,"subscription plan is already active")
		default:
			return nil, status.Error(codes.Internal, "interanal server error")
		}
	}
	return &pb.ActivateSubscriptionPlanResponse{
		Id:           activateSubscriptionPlanResponse.ID,
		CreatedAt:    utils.ToProtoTimestamp(activateSubscriptionPlanResponse.CreatedAt),
		UpdatedAt:    utils.ToProtoTimestamp(activateSubscriptionPlanResponse.UpdatedAt),
		RazorpayPlanId: activateSubscriptionPlanResponse.RazorpayPlanId,
		Name:         activateSubscriptionPlanResponse.Name,
		Price:        activateSubscriptionPlanResponse.Price,
		Currency: activateSubscriptionPlanResponse.Currency,
		Period: activateSubscriptionPlanResponse.Period,
		Interval: activateSubscriptionPlanResponse.Interval,
		Description:  activateSubscriptionPlanResponse.Description,
		IsActive:     activateSubscriptionPlanResponse.IsActive,
	}, nil
}

func (as *AuthSubscriptionServer) DeactivateSubscriptionPlan(ctx context.Context, req *pb.DeactivateSubscriptionPlanRequest) (*pb.DeactivateSubscriptionPlanResponse, error) {
	deactivateSubscriptionPlanReq := requestmodels.DeactivateSubscriptionPlanRequest{
		ID: req.Id,
	}
	deactivateSubscriptionPlanResponse, err := as.AuthSubscriptionUsecase.DeactivateSubscriptionPlan(deactivateSubscriptionPlanReq)
	if err != nil {
		log.Printf("Actvate Subscription paln failed for subscription paln with id =%d: %v", req.Id, err)
		switch {
		case errors.Is(err,usecase.ErrSubscriptionPlanAlreadyDeactive):
			return nil,status.Error(codes.FailedPrecondition,"subscription plan is already deactive")
		default:
			return nil, status.Error(codes.Internal, "interanal server error")
		}
	}
	return &pb.DeactivateSubscriptionPlanResponse{
		Id:           deactivateSubscriptionPlanResponse.ID,
		CreatedAt:    utils.ToProtoTimestamp(deactivateSubscriptionPlanResponse.CreatedAt),
		UpdatedAt:    utils.ToProtoTimestamp(deactivateSubscriptionPlanResponse.UpdatedAt),
		RazorpayPlanId: deactivateSubscriptionPlanResponse.RazorpayPlanId,	
		Name:         deactivateSubscriptionPlanResponse.Name,
		Price:        deactivateSubscriptionPlanResponse.Price,
		Currency: deactivateSubscriptionPlanResponse.Currency,
		Period: deactivateSubscriptionPlanResponse.Period,
		Interval: deactivateSubscriptionPlanResponse.Interval,
		Description:  deactivateSubscriptionPlanResponse.Description,
		IsActive:     deactivateSubscriptionPlanResponse.IsActive,
	}, nil
}

func (as *AuthSubscriptionServer) GetAllSubscriptionPlans(ctx context.Context, req *pb.GetAllSubscriptionPlansRequest) (*pb.GetAllSubscriptionPlansResponse, error) {
	getAllSubscriptionPlanReq := requestmodels.GetAllSubscriptionPlansRequest{
		Limit:  req.Limit,
		Offset: req.Offset,
	}
	subscriptionPlans, err := as.AuthSubscriptionUsecase.GetAllSubscriptionPlans(getAllSubscriptionPlanReq)
	if err != nil {
		log.Printf("Get All Subscription Plans failed : %v", err)
		switch {
		default:
			return nil, status.Error(codes.Internal, "interanal server error")
		}
	}
	pbSubscriptionPlans := make([]*pb.SubscriptioPlan, len(subscriptionPlans.SubscriptionPlans))
	for i, subscriptionPlan := range subscriptionPlans.SubscriptionPlans {
		pbSubscriptionPlans[i] = &pb.SubscriptioPlan{
			Id:           subscriptionPlan.ID,
			CreatedAt:    utils.ToProtoTimestamp(subscriptionPlan.CreatedAt),
			UpdatedAt:    utils.ToProtoTimestamp(subscriptionPlan.UpdatedAt),
			RazorpayPlanId: subscriptionPlan.RazorpayPlanId,
			Name:         subscriptionPlan.Name,
			Price:        subscriptionPlan.Price,
			Currency: subscriptionPlan.Currency,
			Period: subscriptionPlan.Period,
			Interval: subscriptionPlan.Interval,
			Description:  subscriptionPlan.Description,
			IsActive:     subscriptionPlan.IsActive,
		}
	}
	return &pb.GetAllSubscriptionPlansResponse{
		SubscriptioPlans: pbSubscriptionPlans,
	}, nil
}

func (as *AuthSubscriptionServer) GetAllActiveSubscriptionPlans(ctx context.Context, req *pb.GetAllActiveSubscriptionPlansRequest) (*pb.GetAllActiveSubscriptionPlansResponse, error) {
	getAllActiveSubscriptionPlansReq := requestmodels.GetAllActiveSubscriptionPlansRequest{
		Limit:  req.Limit,
		Offset: req.Offset,
	}
	subscriptionPlans, err := as.AuthSubscriptionUsecase.GetAllActiveSubscriptionPlans(getAllActiveSubscriptionPlansReq)
	if err != nil {
		log.Printf("Get All Active Subscription Plans failed : %v", err)
		switch {
		default:
			return nil, status.Error(codes.Internal, "interanal server error")
		}
	}
	pbSubscriptionPlans := make([]*pb.SubscriptioPlan, len(subscriptionPlans.SubscriptionPlans))
	for i, subscriptionPlan := range subscriptionPlans.SubscriptionPlans {
		pbSubscriptionPlans[i] = &pb.SubscriptioPlan{
			Id:           subscriptionPlan.ID,
			CreatedAt:    utils.ToProtoTimestamp(subscriptionPlan.CreatedAt),
			UpdatedAt:    utils.ToProtoTimestamp(subscriptionPlan.UpdatedAt),
			RazorpayPlanId: subscriptionPlan.RazorpayPlanId,
			Name:         subscriptionPlan.Name,
			Price:        subscriptionPlan.Price,
			Currency: subscriptionPlan.Currency,
			Period: subscriptionPlan.Period,
			Interval: subscriptionPlan.Interval,
			Description:  subscriptionPlan.Description,
			IsActive:     subscriptionPlan.IsActive,
		}
	}
	return &pb.GetAllActiveSubscriptionPlansResponse{
		SubscriptioPlans: pbSubscriptionPlans,
	}, nil
}

func (as *AuthSubscriptionServer)Subscribe(ctx context.Context,req *pb.SubscribeReqeust)(*pb.SubscribeResponse,error){
	subscribeReq:=requestmodels.SubscribeRequest{
		UserId: req.UserId,
		PlanId: req.PlanId,
	}
	subscribeRes,err:=as.AuthSubscriptionUsecase.Subscribe(subscribeReq)
	if err!=nil{

	}
	return &pb.SubscribeResponse{
		Id: subscribeRes.ID,
		CreatedAt: utils.ToProtoTimestamp(subscribeRes.CreatedAt),
		UpdatedAt: utils.ToProtoTimestamp(subscribeRes.UpdatedAt),
		UserId: subscribeRes.ID,
		RazorpaySubcriptionId: subscribeRes.RazorpaySubscriptionId,
		Status: subscribeRes.Status,
		TotalCount: int64(subscribeRes.TotalCount),
		RemainingCount: int64(subscribeRes.RemainingCount),
		PaidCount: int64(subscribeRes.PaidCount),
	},nil
}

func (as *AuthSubscriptionServer)VerifySubscriptionPayment(ctx context.Context,req *pb.VerifySubscriptionPaymentRequest)(*pb.VerifySubscriptionPaymentResponse,error){
	verifySubscriptionPaymentReq:=requestmodels.VerifySubscriptionPaymentRequest{
		RazorpaySubscriptionId: req.RazorpaySubscriptionId,
		RazorpayPaymentId: req.RazorpayPaymentId,
	}
	verifySubscriptionPaymentRes,err:=as.AuthSubscriptionUsecase.VerifySubscriptionPayment(verifySubscriptionPaymentReq)
	if err!=nil{
		log.Printf("Get All Subscription Plans failed : %v", err)
		switch {
		default:
			return nil, status.Error(codes.Internal, "interanal server error")
		}
	}
	fmt.Println("print start at",verifySubscriptionPaymentRes.StartAt,utils.ToProtoTimestamp(verifySubscriptionPaymentRes.StartAt))
	return &pb.VerifySubscriptionPaymentResponse{
		Id: verifySubscriptionPaymentRes.ID,
		CreatedAt: utils.ToProtoTimestamp(verifySubscriptionPaymentRes.CreatedAt),
		UpdatedAt: utils.ToProtoTimestamp(verifySubscriptionPaymentRes.UpdatedAt),
		UserId: verifySubscriptionPaymentRes.UserID,
		RazorpaySubcriptionId: verifySubscriptionPaymentReq.RazorpaySubscriptionId,
		Status: verifySubscriptionPaymentRes.Status,
		StartAt: utils.ToProtoTimestamp(verifySubscriptionPaymentRes.StartAt),
		EndAt: utils.ToProtoTimestamp(verifySubscriptionPaymentRes.EndAt),
		NextChargeAt: utils.ToProtoTimestamp(verifySubscriptionPaymentRes.NextChargeAt),
		TotalCount: int64(verifySubscriptionPaymentRes.TotalCount),
		RemainingCount: int64(verifySubscriptionPaymentRes.RemainingCount),
		PaidCount: int64(verifySubscriptionPaymentRes.PaidCount),
	},nil
}

func (as *AuthSubscriptionServer) Unsubscribe(ctx context.Context,req *pb.UnsubscribeRequest)(*pb.UnsubscribeResponse,error){
	unsubscribeReq:=requestmodels.UnsubscribeRequest{
		SubId: req.SubId,
		CancelReason: req.CancelReason,
	}
	unsubscribeRes,err:=as.AuthSubscriptionUsecase.Unsubscribe(unsubscribeReq)
	if err!=nil{

	}
	return &pb.UnsubscribeResponse{
		Id: unsubscribeRes.ID,
		CreatedAt: utils.ToProtoTimestamp(unsubscribeRes.CreatedAt),
		UpdatedAt: utils.ToProtoTimestamp(unsubscribeRes.UpdatedAt),
		UserId: unsubscribeRes.UserID,
		RazorpaySubcriptionId: unsubscribeRes.RazorpaySubscriptionId,
		Status:unsubscribeRes.Status,
		StartAt: utils.ToProtoTimestamp(unsubscribeRes.StartAt),
		EndAt: utils.ToProtoTimestamp(unsubscribeRes.EndAt),
		NextChargeAt: utils.ToProtoTimestamp(unsubscribeRes.NextChargeAt),
		TotalCount: int64(unsubscribeRes.TotalCount),
		RemainingCount: int64(unsubscribeRes.RemainingCount),
		PaidCount: int64(unsubscribeRes.PaidCount),
		CancelledAt: utils.ToProtoTimestamp(unsubscribeRes.CancelledAt),
		CancelReason: unsubscribeReq.CancelReason,
	},nil
}
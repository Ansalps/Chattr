package client

import (
	"context"
	"fmt"
	"log"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client/interfaces"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/pb/auth_subscription"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthSubscriptionClient struct {
	Client auth_subscription.AuthSubscriptionServiceClient
}

func NewAuthSubscriptionClient(cfg *config.Config) interfaces.AuthSubscriptionClient /*auth_subscription.AuthSubscriptionServiceClient*/ {
	grpcConnection, err := grpc.NewClient(cfg.AuthSubscriptionSvcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	fmt.Println("grpc connection ---", grpcConnection)
	grpcClient := auth_subscription.NewAuthSubscriptionServiceClient(grpcConnection)
	return &AuthSubscriptionClient{
		Client: grpcClient,
	}
	//return grpcClient

}

func (as *AuthSubscriptionClient) AdminLogin(adminDetails requestmodels.AdminLoginRequest) (responsemodels.AdminLoginResponse, error) {
	resp, err := as.Client.AdminLogin(context.Background(), &auth_subscription.AdminLoginRequest{
		Email:    adminDetails.Email,
		Password: adminDetails.Password,
	})

	if err != nil {
		log.Printf("grpc admin login call failed :%v", err)
		return responsemodels.AdminLoginResponse{}, err
	}
	return responsemodels.AdminLoginResponse{
		Admin: responsemodels.AdminDetailsResponse{
			ID:    uint(resp.AdminDetails.Id),
			Email: resp.AdminDetails.Email,
		},
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (as *AuthSubscriptionClient) UserSignUp(user requestmodels.UserSignUpRequest) (responsemodels.UserSignupResponse, error) {
	resp, err := as.Client.UserSignUp(context.Background(), &auth_subscription.UserSignUpRequest{
		UserName:        user.UserName,
		Name:            user.Name,
		Email:           user.Email,
		Password:        user.Password,
		ConfirmPassword: user.ConfirmPassword,
	})
	if err != nil {
		log.Printf("grpc user signup call failed :%v", err)
		return responsemodels.UserSignupResponse{}, err
	}
	return responsemodels.UserSignupResponse{
		ID:                   uint(resp.Id),
		UserName:             resp.UserName,
		Name:                 resp.Name,
		Email:                resp.Email,
		OtpVerificationToken: resp.OtpVerificationToken,
	}, nil
}

func (as *AuthSubscriptionClient) VerifyOtp(otpReq requestmodels.OtpRequest) (responsemodels.OtpVerificationResponse, error) {
	resp, err := as.Client.VerifyOtp(context.Background(), &auth_subscription.OtpRequest{
		UserId:  otpReq.UserId,
		OtpCode: otpReq.OtpCode,
		Email:   otpReq.Email,
		Purpose: otpReq.Purpose,
	})
	if err != nil {
		log.Printf("grpc verify otp call failed :%v", err)
		return responsemodels.OtpVerificationResponse{}, err
	}

	return responsemodels.OtpVerificationResponse{
		Email:  resp.Email,
		Status: resp.Status,
	}, nil
}

func (as *AuthSubscriptionClient) ResendOtp(resendOtpReq requestmodels.ResendOtpRequest) (responsemodels.ResendOtpResponse, error) {
	resp, err := as.Client.ResendOtp(context.Background(), &auth_subscription.ResendOtpRequest{
		Name:  resendOtpReq.Email,
		Email: resendOtpReq.Email,
	})
	if err != nil {
		log.Printf("grpc resend otp call failed :%v", err)
		return responsemodels.ResendOtpResponse{}, err
	}
	return responsemodels.ResendOtpResponse{
		Email: resp.Email,
	}, nil
}

func (as *AuthSubscriptionClient) AccessRegenerator(accessRegeneratorReq requestmodels.AccessRegeneratorRequest) (responsemodels.AccessRegeneratorResponse, error) {
	resp, err := as.Client.AccessRegenerator(context.Background(), &auth_subscription.AccessRegeneratorRequest{
		Id:    accessRegeneratorReq.ID,
		Email: accessRegeneratorReq.Email,
		Role:  accessRegeneratorReq.Role,
	})
	if err != nil {
		log.Printf("grpc access regenerator call failed :%v", err)
		return responsemodels.AccessRegeneratorResponse{}, err
	}
	return responsemodels.AccessRegeneratorResponse{
		Id:             resp.Id,
		Email:          resp.Email,
		Role:           resp.Role,
		NewAccessToken: resp.NewAccessToken,
	}, nil
}

func (as *AuthSubscriptionClient) ResetPassword(resetPasswordReq requestmodels.ResetPasswordRequest) (responsemodels.ResetPasswordResponse, error) {
	resp, err := as.Client.ResetPassword(context.Background(), &auth_subscription.ResetPasswordRequest{
		Email:    resetPasswordReq.Email,
		Password: resetPasswordReq.Password,
	})
	if err != nil {
		log.Printf("grpc reset password call failed :%v", err)
		return responsemodels.ResetPasswordResponse{}, err
	}
	return responsemodels.ResetPasswordResponse{
		Email: resp.Email,
	}, nil
}

func (as *AuthSubscriptionClient) BlockUser(blockUserReq requestmodels.BlockUserRequest) (responsemodels.BlockUserResponse, error) {
	resp, err := as.Client.BlockUser(context.Background(), &auth_subscription.BlockUserRequest{
		UserId: blockUserReq.UserId,
	})
	if err != nil {
		log.Printf("grpc block user call failed :%v", err)
		return responsemodels.BlockUserResponse{}, err
	}
	return responsemodels.BlockUserResponse{
		UserId: resp.UserId,
	}, nil
}

func (as *AuthSubscriptionClient) UnblockUser(unblockUserReq requestmodels.UnblockUserRequest) (responsemodels.UnblockUserResponse, error) {
	resp, err := as.Client.UnblockUser(context.Background(), &auth_subscription.UnblockUserRequest{
		UserId: unblockUserReq.UserId,
	})
	if err != nil {
		log.Printf("grpc unblock user call failed :%v", err)
		return responsemodels.UnblockUserResponse{}, err
	}
	return responsemodels.UnblockUserResponse{
		UserId: resp.UserId,
	}, nil
}

func (as *AuthSubscriptionClient) UserLogin(userLoginReq requestmodels.UserLoginRequest) (responsemodels.UserLoginResponse, error) {
	resp, err := as.Client.UserLogin(context.Background(), &auth_subscription.UserLoginRequest{
		Email:    userLoginReq.Email,
		Password: userLoginReq.Password,
	})
	if err != nil {
		log.Printf("grpc user login call failed :%v", err)
		return responsemodels.UserLoginResponse{}, err
	}
	return responsemodels.UserLoginResponse{
		User: responsemodels.UserDetailsResponse{
			Id:       resp.UserDetails.Id,
			Name:     resp.UserDetails.Name,
			UserName: resp.UserDetails.UserName,
			Email:    resp.UserDetails.Email,
			Status:   resp.UserDetails.Status,
			BlueTick: resp.UserDetails.BlueTick,
		},
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (as *AuthSubscriptionClient) GetAllUsers(getAllUsersReq requestmodels.GetAllUsersRequest) (responsemodels.GetAllUsersResponse, error) {
	resp, err := as.Client.GetAllUsers(context.Background(), &auth_subscription.GetAllUsersRequest{
		Limit:  getAllUsersReq.Limit,
		Offset: getAllUsersReq.Offset,
	})
	if err != nil {
		log.Printf("grpc Get All Users call failed: %v", err)
		return responsemodels.GetAllUsersResponse{}, err
	}
	users := make([]responsemodels.User, len(resp.Users))
	for i, user := range resp.Users {
		users[i] = responsemodels.User{
			ID:            user.Id,
			Name:          user.Name,
			UserName:      user.UserName,
			Email:         user.Email,
			Bio:           user.Bio,
			ProfileImgUrl: user.ProfileImgUrl,
			Links:         user.Links,
			Status:        user.Status,
		}
	}
	return responsemodels.GetAllUsersResponse{
		Users: users,
	}, nil
}

func (as *AuthSubscriptionClient) CreateSubscriptionPlan(createSubscriptionPlanReq requestmodels.CreateSubscriptionPlanRequest) (responsemodels.CreateSubscriptionPlanResponse, error) {
	resp, err := as.Client.CreateSubscriptionPlan(context.Background(), &auth_subscription.CreateSubscriptionPlanRequest{
		Name:         createSubscriptionPlanReq.Name,
		Price:        createSubscriptionPlanReq.Price,
		Currency: createSubscriptionPlanReq.Currency,
		Period: createSubscriptionPlanReq.Period,
		Interval: createSubscriptionPlanReq.Interval,
		Description:  createSubscriptionPlanReq.Description,
	})
	if err != nil {
		log.Printf("grpc create subscription plan call failed :%v", err)
		return responsemodels.CreateSubscriptionPlanResponse{}, err
	}
	return responsemodels.CreateSubscriptionPlanResponse{
		ID:           resp.Id,
		CreatedAt:    resp.CreatedAt.AsTime(),
		UpdatedAt:    resp.UpdatedAt.AsTime(),
		Name:         resp.Name,
		Price:        resp.Price,
		Currency: resp.Currency,
		Period: resp.Period,
		Interval: resp.Interval,
		Description:  resp.Description,
		IsActive:     resp.IsActive,
	}, nil
}

func (as *AuthSubscriptionClient) UpdateSubscriptionPlan(updateSubscriptionPlanReq requestmodels.UpdateSubscriptionPlanRequest) (responsemodels.UpdateSubscriptionPlanResponse, error) {
	resp, err := as.Client.UpdateSubscriptionPlan(context.Background(), &auth_subscription.UpdateSubscriptionPlanRequest{
		Id: updateSubscriptionPlanReq.ID,
		Name:         updateSubscriptionPlanReq.Name,
		Price:        float32(updateSubscriptionPlanReq.Price),
		DurationDays: updateSubscriptionPlanReq.DurationDays,
		Description:  updateSubscriptionPlanReq.Description,
	})
	if err != nil {
		log.Printf("grpc update subscription plan call failed: %v", err)
		return responsemodels.UpdateSubscriptionPlanResponse{}, err
	}
	return responsemodels.UpdateSubscriptionPlanResponse{
		ID:           resp.Id,
		CreatedAt:    resp.CreatedAt.AsTime(),
		UpdatedAt:    resp.UpdatedAt.AsTime(),
		Name:         resp.Name,
		Price:        float64(resp.Price),
		DurationDays: resp.DurationDays,
		Description:  resp.Description,
		IsActive:     resp.IsActive,
	}, nil
}

func (as *AuthSubscriptionClient)ActivateSubscriptionPlan(activateSubscriptionPlanReq requestmodels.ActivateSubscriptionPlanRequest)(responsemodels.ActivateSubscriptionPlanResponse,error){
	resp,err:=as.Client.ActivateSubscriptionPlan(context.Background(),&auth_subscription.ActivateSubscriptionPlanRequest{
		Id: activateSubscriptionPlanReq.ID,
	})
	if err!=nil{
		log.Printf("grpc activate subscription plan call failed: %v", err)
		return responsemodels.ActivateSubscriptionPlanResponse{},err
	}
	return responsemodels.ActivateSubscriptionPlanResponse{
		ID:           resp.Id,
		CreatedAt:    resp.CreatedAt.AsTime(),
		UpdatedAt:    resp.UpdatedAt.AsTime(),
		Name:         resp.Name,
		Price:        float64(resp.Price),
		DurationDays: resp.DurationDays,
		Description:  resp.Description,
		IsActive:     resp.IsActive,
	},nil
}

func (as *AuthSubscriptionClient) DeactivateSubscriptionPlan(deactivateSubscriptionPlanReq requestmodels.DeactivateSubscriptionPlanRequest)(responsemodels.DeactivateSubscriptionPlanResponse,error){
	resp,err:=as.Client.DeactivateSubscriptionPlan(context.Background(),&auth_subscription.DeactivateSubscriptionPlanRequest{
		Id: deactivateSubscriptionPlanReq.ID,
	})
	if err!=nil{
		log.Printf("grpc deactivate subscription plan call failed: %v",err)
		return responsemodels.DeactivateSubscriptionPlanResponse{},err
	}
	return responsemodels.DeactivateSubscriptionPlanResponse{
		ID:           resp.Id,
		CreatedAt:    resp.CreatedAt.AsTime(),
		UpdatedAt:    resp.UpdatedAt.AsTime(),
		Name:         resp.Name,
		Price:        float64(resp.Price),
		DurationDays: resp.DurationDays,
		Description:  resp.Description,
		IsActive:     resp.IsActive,
	},nil
}

func (as *AuthSubscriptionClient)GetAllSubscriptionPlans(getAllSubscritpionPlansReq requestmodels.GetAllSubscriptionPlansRequest)(responsemodels.GetAllSubscriptionPlansResponse,error){
	resp,err:=as.Client.GetAllSubscriptionPlans(context.Background(),&auth_subscription.GetAllSubscriptionPlansRequest{
		Limit: getAllSubscritpionPlansReq.Limit,
		Offset: getAllSubscritpionPlansReq.Offset,
	})
	if err!=nil{
		log.Printf("grpc get all subscription plans call failed: %v",err)
		return responsemodels.GetAllSubscriptionPlansResponse{},err
	}
	subscriptionPlans:=make([]responsemodels.SubscriptionPlan,len(resp.SubscriptioPlans))
	for i,subscriptionPlan:=range resp.SubscriptioPlans{
		subscriptionPlans[i]=responsemodels.SubscriptionPlan{
			ID: subscriptionPlan.Id,
			CreatedAt: subscriptionPlan.CreatedAt.AsTime(),
			UpdatedAt: subscriptionPlan.UpdatedAt.AsTime().UTC(),
			Name: subscriptionPlan.Name,
			Price: float64(subscriptionPlan.Price),
			DurationDays: subscriptionPlan.DurationDays,
			Description: subscriptionPlan.Description,
			IsActive: subscriptionPlan.IsActive,
		}
	}
	return responsemodels.GetAllSubscriptionPlansResponse{
		SubscriptionPlans: subscriptionPlans,
	},nil
}

func (as *AuthSubscriptionClient)GetAllActiveSubscriptionPlans(getAllActiveSubscriptionPlansReq requestmodels.GetAllActiveSubscriptionPlansRequest)(responsemodels.GetAllActiveSubscriptionPlansResponse,error){
	resp,err:=as.Client.GetAllActiveSubscriptionPlans(context.Background(),&auth_subscription.GetAllActiveSubscriptionPlansRequest{
		Limit: getAllActiveSubscriptionPlansReq.Limit,
		Offset: getAllActiveSubscriptionPlansReq.Offset,
	})
	if err!=nil{
		log.Printf("grpc get all active subscription plans call failed: %v",err)
		return responsemodels.GetAllActiveSubscriptionPlansResponse{},err
	}
	subscriptionPlans:=make([]responsemodels.SubscriptionPlan,len(resp.SubscriptioPlans))
	for i,subscriptionPlan:=range resp.SubscriptioPlans{
		subscriptionPlans[i]=responsemodels.SubscriptionPlan{
			ID: subscriptionPlan.Id,
			CreatedAt: subscriptionPlan.CreatedAt.AsTime(),
			UpdatedAt: subscriptionPlan.UpdatedAt.AsTime().UTC(),
			Name: subscriptionPlan.Name,
			Price: float64(subscriptionPlan.Price),
			DurationDays: subscriptionPlan.DurationDays,
			Description: subscriptionPlan.Description,
			IsActive: subscriptionPlan.IsActive,
		}
	}
	return responsemodels.GetAllActiveSubscriptionPlansResponse{
		SubscriptionPlans: subscriptionPlans,
	},nil
}
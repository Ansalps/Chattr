package client

import (
	"context"
	"fmt"
	"log"
	"time"

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

func NewAuthSubscriptionClient(cfg *config.Config) interfaces.AuthSubscriptionClient  {
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
	fmt.Print("in client calling server function",otpReq.UserId)
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
		Email:        resp.Email,
		Status:       resp.Status,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TempToken:    resp.TempToken,
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

func (as *AuthSubscriptionClient) ForgotPassword(forgetPasswordReq requestmodels.ForgotPasswordRequest) (responsemodels.ForgetPassordResponse, error) {
	resp, err := as.Client.ForgetPassword(context.Background(), &auth_subscription.ForgotPasswordRequest{
		Email: forgetPasswordReq.Email,
	})
	if err != nil {
		return responsemodels.ForgetPassordResponse{}, err
	}
	return responsemodels.ForgetPassordResponse{
		Email:     resp.Email,
		TempToken: resp.TempToken,
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
	fmt.Println("see if we get back user id in unblock user in client function",resp.UserId)
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
		Name:        createSubscriptionPlanReq.Name,
		Price:       createSubscriptionPlanReq.Price,
		Currency:    createSubscriptionPlanReq.Currency,
		Period:      createSubscriptionPlanReq.Period,
		Interval:    createSubscriptionPlanReq.Interval,
		Description: createSubscriptionPlanReq.Description,
	})
	if err != nil {
		log.Printf("grpc create subscription plan call failed :%v", err)
		return responsemodels.CreateSubscriptionPlanResponse{}, err
	}
	return responsemodels.CreateSubscriptionPlanResponse{
		ID:          resp.Id,
		CreatedAt:   resp.CreatedAt.AsTime(),
		UpdatedAt:   resp.UpdatedAt.AsTime(),
		Name:        resp.Name,
		Price:       resp.Price,
		Currency:    resp.Currency,
		Period:      resp.Period,
		Interval:    resp.Interval,
		Description: resp.Description,
		IsActive:    resp.IsActive,
	}, nil
}

func (as *AuthSubscriptionClient) ActivateSubscriptionPlan(activateSubscriptionPlanReq requestmodels.ActivateSubscriptionPlanRequest) (responsemodels.ActivateSubscriptionPlanResponse, error) {
	resp, err := as.Client.ActivateSubscriptionPlan(context.Background(), &auth_subscription.ActivateSubscriptionPlanRequest{
		Id: activateSubscriptionPlanReq.ID,
	})
	if err != nil {
		log.Printf("grpc activate subscription plan call failed: %v", err)
		return responsemodels.ActivateSubscriptionPlanResponse{}, err
	}
	return responsemodels.ActivateSubscriptionPlanResponse{
		ID:          resp.Id,
		CreatedAt:   resp.CreatedAt.AsTime(),
		UpdatedAt:   resp.UpdatedAt.AsTime(),
		Name:        resp.Name,
		Price:       resp.Price,
		Currency:    resp.Currency,
		Period:      resp.Period,
		Interval:    resp.Interval,
		Description: resp.Description,
		IsActive:    resp.IsActive,
	}, nil
}

func (as *AuthSubscriptionClient) DeactivateSubscriptionPlan(deactivateSubscriptionPlanReq requestmodels.DeactivateSubscriptionPlanRequest) (responsemodels.DeactivateSubscriptionPlanResponse, error) {
	resp, err := as.Client.DeactivateSubscriptionPlan(context.Background(), &auth_subscription.DeactivateSubscriptionPlanRequest{
		Id: deactivateSubscriptionPlanReq.ID,
	})
	if err != nil {
		log.Printf("grpc deactivate subscription plan call failed: %v", err)
		return responsemodels.DeactivateSubscriptionPlanResponse{}, err
	}
	return responsemodels.DeactivateSubscriptionPlanResponse{
		ID:          resp.Id,
		CreatedAt:   resp.CreatedAt.AsTime(),
		UpdatedAt:   resp.UpdatedAt.AsTime(),
		Name:        resp.Name,
		Price:       resp.Price,
		Currency:    resp.Currency,
		Period:      resp.Period,
		Interval:    resp.Interval,
		Description: resp.Description,
		IsActive:    resp.IsActive,
	}, nil
}

func (as *AuthSubscriptionClient) GetAllSubscriptionPlans(getAllSubscritpionPlansReq requestmodels.GetAllSubscriptionPlansRequest) (responsemodels.GetAllSubscriptionPlansResponse, error) {
	resp, err := as.Client.GetAllSubscriptionPlans(context.Background(), &auth_subscription.GetAllSubscriptionPlansRequest{
		Limit:  getAllSubscritpionPlansReq.Limit,
		Offset: getAllSubscritpionPlansReq.Offset,
	})
	if err != nil {
		log.Printf("grpc get all subscription plans call failed: %v", err)
		return responsemodels.GetAllSubscriptionPlansResponse{}, err
	}
	subscriptionPlans := make([]responsemodels.SubscriptionPlan, len(resp.SubscriptioPlans))
	for i, subscriptionPlan := range resp.SubscriptioPlans {
		subscriptionPlans[i] = responsemodels.SubscriptionPlan{
			ID:             subscriptionPlan.Id,
			CreatedAt:      subscriptionPlan.CreatedAt.AsTime(),
			UpdatedAt:      subscriptionPlan.UpdatedAt.AsTime().UTC(),
			RazorpayPlanId: subscriptionPlan.RazorpayPlanId,
			Name:           subscriptionPlan.Name,
			Price:          subscriptionPlan.Price,
			Currency:       subscriptionPlan.Currency,
			Period:         subscriptionPlan.Period,
			Interval:       subscriptionPlan.Interval,
			Description:    subscriptionPlan.Description,
			IsActive:       subscriptionPlan.IsActive,
		}
	}
	return responsemodels.GetAllSubscriptionPlansResponse{
		SubscriptionPlans: subscriptionPlans,
	}, nil
}

func (as *AuthSubscriptionClient) GetAllActiveSubscriptionPlans(getAllActiveSubscriptionPlansReq requestmodels.GetAllActiveSubscriptionPlansRequest) (responsemodels.GetAllActiveSubscriptionPlansResponse, error) {
	resp, err := as.Client.GetAllActiveSubscriptionPlans(context.Background(), &auth_subscription.GetAllActiveSubscriptionPlansRequest{
		Limit:  getAllActiveSubscriptionPlansReq.Limit,
		Offset: getAllActiveSubscriptionPlansReq.Offset,
	})
	if err != nil {
		log.Printf("grpc get all active subscription plans call failed: %v", err)
		return responsemodels.GetAllActiveSubscriptionPlansResponse{}, err
	}
	subscriptionPlans := make([]responsemodels.SubscriptionPlan, len(resp.SubscriptioPlans))
	for i, subscriptionPlan := range resp.SubscriptioPlans {
		subscriptionPlans[i] = responsemodels.SubscriptionPlan{
			ID:             subscriptionPlan.Id,
			CreatedAt:      subscriptionPlan.CreatedAt.AsTime(),
			UpdatedAt:      subscriptionPlan.UpdatedAt.AsTime().UTC(),
			RazorpayPlanId: subscriptionPlan.RazorpayPlanId,
			Name:           subscriptionPlan.Name,
			Price:          subscriptionPlan.Price,
			Currency:       subscriptionPlan.Currency,
			Period:         subscriptionPlan.Period,
			Interval:       subscriptionPlan.Interval,
			Description:    subscriptionPlan.Description,
			IsActive:       subscriptionPlan.IsActive,
		}
	}
	return responsemodels.GetAllActiveSubscriptionPlansResponse{
		SubscriptionPlans: subscriptionPlans,
	}, nil
}

func (as *AuthSubscriptionClient) Subscribe(subscribeReq requestmodels.SubscribeRequest) (responsemodels.SubscribeResponse, error) {
	resp, err := as.Client.Subscribe(context.Background(), &auth_subscription.SubscribeReqeust{
		UserId: subscribeReq.UserId,
		PlanId: subscribeReq.PlanId,
	})
	if err != nil {
		log.Printf("grpc subscribe call failed: %v", err)
		return responsemodels.SubscribeResponse{}, err
	}
	return responsemodels.SubscribeResponse{
		ID:                     resp.Id,
		CreatedAt:              resp.CreatedAt.AsTime(),
		UpdatedAt:              resp.UpdatedAt.AsTime(),
		UserID:                 resp.UserId,
		RazorpaySubscriptionId: resp.RazorpaySubcriptionId,
		Status:                 resp.Status,
		TotalCount:             int(resp.TotalCount),
		RemainingCount:         int(resp.RemainingCount),
		PaidCount:              int(resp.PaidCount),
	}, nil
}

func (as *AuthSubscriptionClient) VerifySubscriptionPayment(verifySubscriptionPaymentReq requestmodels.VerifySubscriptionPaymentRequest) (responsemodels.VerifySubscriptionPaymentResponse, error) {
	resp, err := as.Client.VerifySubscriptionPayment(context.Background(), &auth_subscription.VerifySubscriptionPaymentRequest{
		RazorpaySubscriptionId: verifySubscriptionPaymentReq.RazorpaySubscriptionId,
		RazorpayPaymentId:      verifySubscriptionPaymentReq.RazorpayPaymentId,
		RazorpaySignature:      verifySubscriptionPaymentReq.RazorpaySignature,
	})
	if err != nil {
		return responsemodels.VerifySubscriptionPaymentResponse{}, err
	}
	fmt.Println("just printing in apin gateway --", resp.StartAt, resp.StartAt.AsTime())
	loc, _ := time.LoadLocation("Asia/Kolkata")
	return responsemodels.VerifySubscriptionPaymentResponse{
		ID:                     resp.Id,
		CreatedAt:              resp.CreatedAt.AsTime().In(loc),
		UpdatedAt:              resp.UpdatedAt.AsTime().In(loc),
		UserID:                 resp.UserId,
		RazorpaySubscriptionId: resp.RazorpaySubcriptionId,
		Status:                 resp.Status,
		StartAt:                resp.StartAt.AsTime().In(loc),
		EndAt:                  resp.EndAt.AsTime().In(loc),
		NextChargeAt:           resp.NextChargeAt.AsTime().In(loc),
		TotalCount:             int(resp.TotalCount),
		RemainingCount:         int(resp.RemainingCount),
		PaidCount:              int(resp.PaidCount),
	}, nil
}

func (as *AuthSubscriptionClient) Unsubscribe(unsubscribeReq requestmodels.UnsubscribeRequest) (responsemodels.UnsubscribeResponse, error) {
	resp, err := as.Client.Unsubscribe(context.Background(), &auth_subscription.UnsubscribeRequest{
		SubId:        unsubscribeReq.SubId,
		CancelReason: unsubscribeReq.CancelReason,
	})
	if err != nil {
		return responsemodels.UnsubscribeResponse{}, err
	}
	loc, _ := time.LoadLocation("Asia/Kolkata")
	return responsemodels.UnsubscribeResponse{
		ID:                     resp.Id,
		CreatedAt:              resp.CreatedAt.AsTime().In(loc),
		UpdatedAt:              resp.UpdatedAt.AsTime().In(loc),
		UserID:                 resp.UserId,
		RazorpaySubscriptionId: resp.RazorpaySubcriptionId,
		Status:                 resp.Status,
		StartAt:                resp.StartAt.AsTime().In(loc),
		EndAt:                  resp.EndAt.AsTime().In(loc),
		NextChargeAt:           resp.NextChargeAt.AsTime().In(loc),
		TotalCount:             int(resp.TotalCount),
		RemainingCount:         int(resp.RemainingCount),
		PaidCount:              int(resp.PaidCount),
		CancelledAt:            resp.CancelledAt.AsTime().In(loc),
		CancelReason:           resp.CancelReason,
	}, nil
}

func (as *AuthSubscriptionClient)SetProfileImage(setProfileImgReq requestmodels.SetProfileImageRequest)(responsemodels.SetProfileImageResponse,error){
	var resp *auth_subscription.SetProfileImageResponse
	resp,err:=as.Client.SetProfileImage(context.Background(),&auth_subscription.SetProfileImageRequest{
		UserId: setProfileImgReq.UserId,
		ContentType: setProfileImgReq.ContentType,
		Image: setProfileImgReq.Image,
	})
	if err!=nil{
		log.Println("print in client error",err)
		return responsemodels.SetProfileImageResponse{},err
	}
	fmt.Println("if no error pint  image url",resp.ImageUrl)
	return responsemodels.SetProfileImageResponse{
		ImageUrl: resp.ImageUrl,
	},nil
}
// func (as *AuthSubscriptionClient)Webhook(webhookReq requestmodels.WebhookRequest)(responsemodels.WebhookResponse,error){
// 	resp,err:=as.Client.Webhook(context.Background(),&auth_subscription.WebhookRequest{
// 		Event: webhookReq.Event,
// 		Payload: &auth_subscription.Payload{
// 			Subscription: &auth_subscription.Subscription{
// 				Id: webhookReq.Payload.Subscription.ID,
// 			},
// 		},
// 	})
// 	if err!=nil{
// 		return responsemodels.WebhookResponse{},nil
// 	}
// 	return responsemodels.WebhookResponse{
// 		RazropaySubscriptinId:resp.RazorpaySubscriptionId,
// 		Event: resp.Event,
// 	},nil
// }

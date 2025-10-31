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
		return responsemodels.AdminLoginResponse{}, err
	}
	return responsemodels.AdminLoginResponse{
		Admin: responsemodels.AdminDetailsResponse{
			ID:    uint(resp.AdminDetails.Id),
			Email: resp.AdminDetails.Email,
		},
		Token: resp.Token,
	}, nil
}

func (as *AuthSubscriptionClient) UserSignUp(user requestmodels.UserSignUpRequest) (responsemodels.UserSignupResponse, error) {
	resp,err:=as.Client.UserSignUp(context.Background(),&auth_subscription.UserSignUpRequest{
		UserName: user.UserName,
		Name: user.Name,
		Email: user.Email,
		Password: user.Password,
		ConfirmPassword: user.ConfirmPassword,
	})
	if err != nil {
		return responsemodels.UserSignupResponse{}, err
	}
	return responsemodels.UserSignupResponse{
		
	}, nil
}

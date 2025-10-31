package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/pb"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/usecase"
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
	fmt.Println("adminToken",admin.Token)
	return &pb.AdminLoginResponse{
		AdminDetails: adminDetails,
		Token:        admin.Token,
	}, nil
}

func (as *AuthSubscriptionServer) UserSignUp(ctx context.Context,req *pb.UserSignUpRequest) (*pb.UserSignUpResponse,error){
	userSignup:=requestmodels.UserSignUpRequest{
		UserName: req.UserName,
		Name: req.Name,
		Email: req.Email,
		Password: req.Password,
		ConfirmPassword: req.ConfirmPassword,
	}
	userResponse,err:=as.AuthSubscriptionUsecase.UserSignUp(userSignup)
}
package usecase

import (
	"errors"
	"fmt"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/helper"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/responsemodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/usecase/interfacesUsecase"
	"gorm.io/gorm"
)

type AuthSubscriptionUsecase struct {
	AuthSubscriptionRepository interfacesRepository.AuthSubscriptionRepository
}

func NewAuthSubscriptionUsecase(repository interfacesRepository.AuthSubscriptionRepository) interfacesUsecase.AuthSubscriptionUsecase {
	return &AuthSubscriptionUsecase{
		AuthSubscriptionRepository: repository,
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
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
	tokenString, err := helper.GenerateToken(uint64(admins.ID), admins.Email, "admin")
	if err != nil {
		return responsemodels.AdminLoginResponse{}, fmt.Errorf("Failed to generarate token: %w", err)
	}
	return responsemodels.AdminLoginResponse{
		Admin: responsemodels.AdminDetails{
			ID:    admins.ID,
			Email: admins.Email,
		},
		Token: tokenString,
	}, nil
}

func(as *AuthSubscriptionUsecase) UserSignUp(userReq requestmodels.UserSignUpRequest)(responsemodels.UserSignupResponse,error){
	
}
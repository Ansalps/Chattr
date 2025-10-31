package interfacesUsecase

import (
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/responsemodels"
)

type AuthSubscriptionUsecase interface{
	AdminLogin(admin requestmodels.AdminLoginRequest)(responsemodels.AdminLoginResponse,error)
	UserSignUp(requestmodels.UserSignUpRequest)(responsemodels.UserSignupResponse,error)
}

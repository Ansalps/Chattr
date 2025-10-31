package interfaces

import (
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
)

type AuthSubscriptionClient interface{
	AdminLogin(requestmodels.AdminLoginRequest) (responsemodels.AdminLoginResponse,error)
	UserSignUp(requestmodels.UserSignUpRequest)(responsemodels.UserSignupResponse,error)
}
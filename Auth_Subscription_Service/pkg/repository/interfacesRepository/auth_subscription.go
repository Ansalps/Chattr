package interfacesRepository

import (
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
	//"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
)

type AuthSubscriptionRepository interface{
	//AdminLogin(admin requestmodels.AdminLoginRequest)(domain.Admin,error)
	CheckAdminExistsByEmail(email string) (domain.Admin,error)
}
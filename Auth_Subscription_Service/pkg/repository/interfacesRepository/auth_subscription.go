package interfacesRepository

import (
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/responsemodels"
	//"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
)

type AuthSubscriptionRepository interface {
	//AdminLogin(admin requestmodels.AdminLoginRequest)(domain.Admin,error)
	CheckAdminExistsByEmail(email string) (*domain.Admin, error)
	DeletePendingUser(email string)(error)
	CheckUserExistsByEmail(email string) (*domain.User, error)
	CheckUserExistsByUseraname(username string) (*domain.User, error)
	DeleteOtpByEmail(email string)(error)
	TemporarySavingUserOtp(otp int, userEmail string, expiration time.Time) error
	CreateUser(userData *requestmodels.UserSignUpRequest) (*responsemodels.UserSignupResponse,error)
	CheckOtpExistsByEmail(requestmodels.OtpRequest)(*domain.Otp,error)
	ChangeOtpStatus(email string)error
	ChangeUserStatusByEmail(email string)error
	UpdatePassword(requestmodels.ResetPasswordRequest)error
	CheckUserStatus(uint64)(string,error)
	ChangeUserStatusToBlockedByUserId(requestmodels.BlockUserRequest)error
	ChangeUserStatusToActiveByUserId(requestmodels.UnblockUserRequest)error
	GetAllUsers(requestmodels.GetAllUsersRequest)(responsemodels.GetAllUsersResponse,error)
	CreateSubscriptionPlan(map[string]interface{})(responsemodels.CreateSubscriptionPlanResponse,error)
	CreateSubscription(requestmodels.SubscribeRequest,map[string]interface{})(responsemodels.SubscribeResponse,error)
	
	ActivateSubscriptionPlan(requestmodels.ActivateSubscriptionPlanRequest)(responsemodels.ActivateSubscriptionPlanResponse,error)
	DeactivateSubscriptionPlan(requestmodels.DeactivateSubscriptionPlanRequest)(responsemodels.DeactivateSubscriptionPlanResponse,error)
	FetchStatusFromSubcriptionPlan(uint64)(bool,error)
	GetAllSubscriptionPlans(requestmodels.GetAllSubscriptionPlansRequest)(responsemodels.GetAllSubscriptionPlansResponse,error)
	GetAllActiveSubscriptionPlans(requestmodels.GetAllActiveSubscriptionPlansRequest)(responsemodels.GetAllActiveSubscriptionPlansResponse,error)
	FetchRazorpayPlanIdFromId(uint64)(string,error)
	UpdateUserSubscripion(string,map[string]interface{})(responsemodels.VerifySubscriptionPaymentResponse,error)
	FetchAmountCurrencyFromSubscriptionPlan(id uint64)(int64,string,error)
	FetchRazorpaySubscriptionIdFromSubcriptionId(subid uint64)(string,error)
	ChangeUserSubscriptionStatusToCancelled(uint64,map[string]interface{})(responsemodels.UnsubscribeResponse,error)
	FetchUserIdFromSubscriptionId(string)(uint64,error)
	TurnBlueTickTrueForUserId(uint64)error
	PopulatePayment(map[string]interface{},requestmodels.VerifySubscriptionPaymentRequest)(domain.Payment,error)
	FetchRazorpayPlanIdFromRazrorpaySubscriptionId(string)(string,error)
	FetchIntervalPeriodFromSubscriptionPlan(planid string)(string,uint64,error)
	FetchTotalCountFromUserSubscription(subId string)(int,error)
	UpdateTimeUserSubscription(startAt,nextAt,nextChatgeAT time.Time,subid string)(responsemodels.VerifySubscriptionPaymentResponse,error)
	FetchNextChargeAtFromUserSubcription(string)(time.Time,error)
	TurnOffBlueTickForUserId(userid uint64)error
	UpdateProfileImage(userid uint64,imageUrl string)(error)

	CheckUserExistsById(userid uint64)(bool,error)

	SearchUser(requestmodels.SearchUser)(responsemodels.SearchUserResponse,error)

	GetProfileInformation(requestmodels.GetProfileInformationRequest)(responsemodels.GetProfileInformationResponse,error)
	EditProfileInformation(uint64,map[string]interface{})(responsemodels.EditProfile,error)

	FetchHashedPassword(requestmodels.ChangePassword)(string,error)
	ChangePassword(requestmodels.ChangePassword,string)(responsemodels.ChangePasswordResponse,error)

	FetchUserPublicData(uint64)(responsemodels.UserPublicDataResponse,error)
}

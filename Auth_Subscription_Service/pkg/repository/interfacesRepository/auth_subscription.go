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
	CreateSubscriptionPlan(*domain.CreatedPlanDTO)(responsemodels.CreateSubscriptionPlanResponse,error)
	CreateSubscription(requestmodels.SubscribeRequest,*domain.CreatedSubscriptionDTO)(responsemodels.SubscribeResponse,error)
	
	ActivateSubscriptionPlan(requestmodels.ActivateSubscriptionPlanRequest)(responsemodels.ActivateSubscriptionPlanResponse,error)
	DeactivateSubscriptionPlan(requestmodels.DeactivateSubscriptionPlanRequest)(responsemodels.DeactivateSubscriptionPlanResponse,error)
	FetchStatusFromSubcriptionPlan(uint64)(bool,error)
	GetAllSubscriptionPlans(requestmodels.GetAllSubscriptionPlansRequest)(responsemodels.GetAllSubscriptionPlansResponse,error)
	GetAllActiveSubscriptionPlans(requestmodels.GetAllActiveSubscriptionPlansRequest)(responsemodels.GetAllActiveSubscriptionPlansResponse,error)
	FetchRazorpayPlanIdFromId(uint64)(string,error)
	UpdateUserSubscripion(string,map[string]interface{})(responsemodels.VerifySubscriptionPaymentResponse,error)
	FetchAmountCurrencyFromSubscriptionPlan(id uint64)(int64,string,error)
	FetchRazorpaySubscriptionIdFromSubcriptionId(subid uint64)(string,error)
	SetCancelReason(requestmodels.UnsubscribeRequest)(responsemodels.UnsubscribeResponse,error)
	FetchUserIdFromSubscriptionId(string)(uint64,error)
	TurnBlueTickTrueForUserId(uint64)error
	//PopulatePayment(map[string]interface{},requestmodels.VerifySubscriptionPaymentRequest)(domain.Payment,error)
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

	FetchUserMetaData([]uint64)(map[uint64]responsemodels.UserMetaData,error)
	CheckUserListExists(userids []uint64)([]uint64,error)

	GetSubscriptionDetails(requestmodels.GetSubscriptionDetails)(responsemodels.GetSubscriptionDetails,error)

	UpddateActivatedSubscription(requestmodels.WebhookSubscriptionActivatedRequest)(responsemodels.WebhookSubscriptionActivatedResponse,error)

	UpdateNextChargeAt(time.Time,string)error

	UpdatePayment(requestmodels.WebhookSubscriptionChargedRequest)(responsemodels.WebhookSubscriptionChargedResponse,error)

	UpdateStatusHalted(requestmodels.WebhookSubscriptionHaltedRequest)error

	UpdateSubscriptionCancelled(requestmodels.WebhookSubscriptionCancelledRequest)(responsemodels.WebhookSubscriptionCancelledResponse,error)

	UpdateSubscripionCompleted(requestmodels.WebhookSubscriptionCompletedRequest)(responsemodels.WebhookSubscriptionCompletedResponse,error)

	IsEligibleForSubsciption(requestmodels.SubscribeRequest)(bool,error)

	UpdateStatusToActive(status string, razorpaySubId string) error

	UpdateCount(requestmodels.WebhookSubscriptionChargedRequest)error

	FetchUserSubscription(uint64)(string,error)

	DoesUserExists(userid uint64)(bool,error)

	UpdateUserRazorpayCustomerID(uint64,string)error

	CheckAllUsersExists([]uint64)([]uint64,error)
}

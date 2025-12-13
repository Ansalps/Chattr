package responsemodels

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Type  string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type AdminLoginResponse struct {
	Admin AdminDetailsResponse
	AccessToken string
	RefreshToken string
}
type AdminDetailsResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

type UserSignupResponse struct {
	ID                   uint
	UserName             string
	Name                 string
	Email                string
	OtpVerificationToken string
}

type OtpVerificationResponse struct {
	Email    string
	Status   string
	AccessToken string
	RefreshToken string
	TempToken string
}

type ResendOtpResponse struct{
	Email string
}

type AccessRegeneratorResponse struct{
	Id uint64
	Email string
	Role string
	NewAccessToken string
}
type ForgetPassordResponse struct{
	Email string
	TempToken string
}
type ResetPasswordResponse struct{
	Email string
}

type BlockUserResponse struct{
	UserId uint64 
}

type UnblockUserResponse struct{
	UserId uint64
}

type UserDetailsResponse struct{
	Id uint64
	Name string
	UserName string
	Email string
	Status string
	BlueTick bool
}
type UserLoginResponse struct{
	User UserDetailsResponse
	AccessToken string
	RefreshToken string
}

type User struct{
	ID uint64
	Name string
	UserName string
	Email string
	Bio string
	ProfileImgUrl string
	Links string
	Status string
}

type GetAllUsersResponse struct{
	Users []User
}

type CreateSubscriptionPlanResponse struct{
	ID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	Name string
	Price int64
	Currency string
	Period string
	Interval uint64
	Description string
	IsActive bool
}

type UpdateSubscriptionPlanResponse struct{
	ID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	Name string
	Price int64
	Currency string
	Period string
	Interval uint64
	Description string
	IsActive bool
}

type ActivateSubscriptionPlanResponse struct{
	ID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	RazorpayPlanId string
	Name string
	Price int64
	Currency string
	Period string
	Interval uint64
	Description string
	IsActive bool
}

type DeactivateSubscriptionPlanResponse struct{
	ID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	RazorpayPlanId string
	Name string
	Price int64
	Currency string
	Period string
	Interval uint64
	Description string
	IsActive bool
}

type SubscriptionPlan struct{
	ID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	RazorpayPlanId string
	Name string
	Price int64
	Currency string
	Period string
	Interval uint64
	Description string
	IsActive bool
}

type GetAllSubscriptionPlansResponse struct{
	SubscriptionPlans []SubscriptionPlan
}

type GetAllActiveSubscriptionPlansResponse struct{
	SubscriptionPlans []SubscriptionPlan
}

// type SubscribeResponse struct{
// 	RazorpaySubcriptionId string
// }
type SubscribeResponse struct{
	ID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID uint64
	RazorpaySubscriptionId string
	Status string
	TotalCount int
	RemainingCount int
	PaidCount int
}

type VerifySubscriptionPaymentResponse struct{
	ID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID uint64
	RazorpaySubscriptionId string
	Status string
	StartAt time.Time
	EndAt	time.Time
	NextChargeAt time.Time
	TotalCount int
	RemainingCount int
	PaidCount int
}

type UnsubscribeResponse struct{
	ID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID uint64
	RazorpaySubscriptionId string
	Status string
	StartAt time.Time
	EndAt	time.Time
	NextChargeAt time.Time
	TotalCount int
	RemainingCount int
	PaidCount int
	CancelledAt time.Time
	CancelReason string
}

type WebhookResponse struct{
	Event string
	RazropaySubscriptinId string
}

type SetProfileImageResponse struct{
	ImageUrl string
}
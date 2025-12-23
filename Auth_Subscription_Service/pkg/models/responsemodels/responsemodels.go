package responsemodels

import "time"

type AdminLoginResponse struct {
	Admin        AdminDetails
	AccessToken  string
	RefreshToken string
}
type AdminDetails struct {
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
	Email        string
	Status       string
	AccessToken  string
	RefreshToken string
	TempToken    string
}

type ResendOtpResponse struct {
	Email string
}

type AccessRegeneratorResponse struct {
	Id             uint64
	Email          string
	Role           string
	NewAccessToken string
}
type ForgotPassordResponse struct{
	Email string
	TempToken string
}
type ResetPasswordResponse struct {
	Email string
}

type BlockUserResponse struct {
	UserId uint64
}

type UnblockUserResponse struct {
	UserId uint64
}

type UserDetailsResponse struct {
	Id       uint64
	Name     string
	UserName string
	Email    string
	Status   string
	BlueTick bool
}
type UserLoginResponse struct {
	User         UserDetailsResponse
	AccessToken  string
	RefreshToken string
}

type User struct {
	ID            uint64
	Name          string
	UserName      string
	Email         string
	Bio           string
	ProfileImgUrl string
	Links         string
	Status        string
}

type GetAllUsersResponse struct {
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
	Amount int64
	Currency string
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

type SetProfileImageResponse struct{
	ImageUrl string
}

type GetProfileInformationResponse struct{
	UserID uint64
	Name string
	UserName      string
	Email         string
	Bio           string
	ProfileImgUrl string
	Links         string
	BlueTick	bool	
}

type EditProfile struct{
	UserID uint64	`json:"user_id"`
	Name *string	`json:"name,omitempty"`
	Bio *string	`json:"bio,omitempty"`
	Links *string `json:"links,omitempty"`
}

type ChangePasswordResponse struct{
	UserID uint64
}
type UserMetaData struct {
	UserID        uint64
	UserName      string
	Name          string
	ProfileImgUrl string
	BlueTick      bool
}
type SearchUserResponse struct{
	Usermetadata []UserMetaData
}
type WebhookResponse struct{
	Event string
	RazropaySubscriptinId string
}
type UserPublicDataResponse struct{
	UserID uint64
	UserName string
	Name string
	ProfileImgUrl string
	Bio string
	Links string
	BlueTick bool
}


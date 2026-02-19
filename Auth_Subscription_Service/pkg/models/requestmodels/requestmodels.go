package requestmodels

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required" validat:"required"`
	Password string `json:"password" binding:"required" validate:"min=6 max=20"`
}

type JwtClaims struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Type  string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type UserSignUpRequest struct {
    Name            string `json:"Name" binding:"required,min=3,max=30"`
    UserName        string `json:"UserName" binding:"required,min=3,max=30"`
    Email           string `json:"Email" binding:"required,email"`
    Password        string `json:"Password" binding:"required,min=3,max=30"`
    ConfirmPassword string `json:"ConfirmPassword" binding:"required,eqfield=Password"`
	Phone string	`json:"phone" binding:"required,regexp=^[0-9]{10}$"`
}

type OtpRequest struct{
	UserId uint64
    OtpCode uint64 `json:"otp_code" bindiing:"required"`
	Email string
	Purpose string
}

type AccessRegeneratorRequest struct{
    ID uint64 `json:"id" binding:"required"`
    Email string `json:"email" binding:"required"`
    Role string `json:"role" binding:"required"`
}

type ResendOtpRequest struct{
	Name string `json:"name" bindig:"required"`
    Email string `json:"email" binding:"required"`
}
type ForgotPasswordRequest struct{
    Email string `json:"email" binding:"required"`
}
type ResetPasswordRequest struct{
    Email string
    Password        string `json:"Password" binding:"required,min=3,max=30"`
    //ConfirmPassword string `json:"ConfirmPassword" binding:"required,eqfield=Password"`
}

type BlockUserRequest struct{
    UserId uint64 `json:"user_id" binding:"required"`
}

type UnblockUserRequest struct{
    UserId uint64 `json:"user_id" binding:"required"`
}

type UserLoginRequest struct{
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6,max=20"`
}

type GetAllUsersRequest struct{
    Limit uint64
    Offset uint64
}

type CreateSubscriptionPlanRequest struct{
    Name string    `json:"name" binding:"required"`
    Price int64   `json:"price" binding:"required"`
    Currency string `json:"currency" binding:"required"`
    Period string `json:"period" binding:"required"`
    Interval uint64 `json:"interval" binding:"required"`
    Description string `json:"description" binding:"required"`
}

type UpdateSubscriptionPlanRequest struct {
	ID           uint64
	Name        string `json:"name" binding:"required"`
	Price       int64  `json:"price" binding:"required"`
	Currency    string `json:"currency" binding:"required"`
	Period      string `json:"period" binding:"required"`
	Interval    uint64 `json:"interval" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type ActivateSubscriptionPlanRequest struct{
    ID uint64
}

type DeactivateSubscriptionPlanRequest struct{
    ID uint64
}

type GetAllSubscriptionPlansRequest struct{
    Limit uint64
    Offset uint64
}

type GetAllActiveSubscriptionPlansRequest struct{
    Limit uint64
    Offset uint64
}

type SubscribeRequest struct{
    UserId uint64
    PlanId uint64
	UserEmail string
	TotalCount uint64
}

type VerifySubscriptionPaymentRequest struct{
    RazorpaySubscriptionId string
    RazorpayPaymentId string
    RazorpaySignature string
}

type UnsubscribeRequest struct{
	RazorpaySubId string
	UserID uint64
	CancelReason string	`json:"cancel_reason" binding:"required"`
	CancelAtCycleEnd bool
}
type SetProfileImageRequest struct{
	UserId uint64
    ContentType string
	Image []byte
}

type GetProfileInformationRequest struct{
	UserId uint64
}

type ChangePassword struct{
	UserID uint64
	OldPassword	string	`json:"old_password" validate:"required"`
	NewPassword        string `json:"new_password" binding:"required,min=3,max=30"`
	ConfirmNewPassword string `json:"confirm_new_password" binding:"required,eqfield=Password"`
}
type SearchUser struct{
	SearchText string `json:"search_text"`
    Limit int64
	Offset int64
}


type RazorpayEvent struct {
	Entity    string   `json:"entity"`
	AccountID string   `json:"account_id"`
	Event     string   `json:"event"` // e.g., "subscription.completed"
	Contains  []string `json:"contains"`
	Payload   Payload  `json:"payload"`
	CreatedAt int64    `json:"created_at"`
}
type Payload struct {
	Subscription SubscriptionWrapper `json:"subscription"`
}

type SubscriptionWrapper struct {
	Entity SubscriptionEntity `json:"entity"`
}

type SubscriptionEntity struct {
    ID             string            `json:"id"`
    Status         string            `json:"status"`
    PlanID         string            `json:"plan_id"`
    Notes          map[string]string `json:"notes"`
    // Added fields
    StartAt        int64             `json:"start_at"`         // Use int64 for Unix timestamps
    EndAt          int64             `json:"end_at"`
    //NextChargeAt   int64             `json:"next_charge_at"`
	CurrentEnd   int64             `json:"current_end"`
	EndedAt          int64 `json:"ended_at"`
    RemainingCount int               `json:"remaining_count"`
    PaidCount      int               `json:"paid_count"`
    TotalCount     int               `json:"total_count"`
}

type WebhookSubscriptionActivatedRequest struct{
	RazorpaySubscriptionId string
	Status string
	PaidCount int
	RemainingCount int
	StartAt        time.Time             `json:"start_at"`         // Use int64 for Unix timestamps
    EndAt          time.Time             `json:"end_at"`
	UserID uint64
}
type WebhookSubscriptionChargedRequest struct{
	RazorpaySubscriptionId string
	RazorpayPlanId string
	NextChargeAt time.Time `json:"current_end"`
	InvoiceID      string `json:"invoice_id"`
	Amount         int64  `json:"amount"` // Amount in paise (e.g., 50000 for ₹500)
    Currency       string `json:"currency"`
	Method         string `json:"method"` // card, upi, etc.
    Status         string `json:"status"`
	PaidCount int
	RemainingCount int
    TransactionDate time.Time    `json:"transaction_date"`
	PaymentID  string `json:"id"`
	UserID uint64
}
type WebhookSubscriptionHaltedRequest struct{
	RazorpaySubscriptionId string
	Status string
	UserId uint64
}
type WebhookSubscriptionCancelledRequest struct{
	RazorpaySubscriptionId string
	Status string
	CancelledAt time.Time
	UserId uint64
}
type WebhookSubscriptionCompletedRequest struct{
	RazorpaySubscriptionId string
	Status string
	UserId uint64
}
type GetSubscriptionDetails struct{
	UserID uint64
	//SubID uint64
}


package requestmodels

type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

type UserSignUpRequest struct {
	Name            string `json:"Name" binding:"required,min=3,max=30"`
	UserName        string `json:"UserName" binding:"required,min=3,max=30"`
	Email           string `json:"Email" binding:"required,email"`
	Password        string `json:"Password" binding:"required,min=3,max=30"`
	ConfirmPassword string `json:"ConfirmPassword" binding:"required,eqfield=Password"`
}

type OtpRequest struct {
	UserId  uint64
	OtpCode uint64 `json:"otp_code" binding:"required"`
	Email   string
	Purpose string `json:"purpose" binding:"required,oneof=user-forgot-password user-signup"`
}

type ResendOtpRequest struct {
	Name  string `json:"name" bindig:"required"`
	Email string `json:"email" binding:"required"`
}

type AccessRegeneratorRequest struct {
	ID    uint64 `json:"id" binding:"required"`
	Email string `json:"email" binding:"required"`
	Role  string `json:"role" binding:"required"`
}

type ForgotPasswordRequest struct{
    Email string `json:"email" binding:"required"`
}
type ResetPasswordRequest struct {
	Email           string
	Password        string `json:"Password" binding:"required,min=3,max=30"`
	ConfirmPassword string `json:"ConfirmPassword" binding:"required,eqfield=Password"`
}

type BlockUserRequest struct {
	UserId uint64 `json:"user_id" binding:"required"`
}

type UnblockUserRequest struct {
	UserId uint64 `json:"user_id" binding:"required"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

type GetAllUsersRequest struct {
	Limit  uint64
	Offset uint64
}

type CreateSubscriptionPlanRequest struct {
	Name        string `json:"name" binding:"required"`
	Price       int64  `json:"price" binding:"required"`
	Currency    string `json:"currency" binding:"required"`
	Period      string `json:"period" binding:"required"`
	Interval    uint64 `json:"interval" binding:"required"`
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

type ActivateSubscriptionPlanRequest struct {
	ID uint64
}

type DeactivateSubscriptionPlanRequest struct {
	ID uint64
}

type GetAllSubscriptionPlansRequest struct {
	Limit  uint64
	Offset uint64
}

type GetAllActiveSubscriptionPlansRequest struct {
	Limit  uint64
	Offset uint64
}

type SubscribeRequest struct{
    UserId uint64
    PlanId uint64
	UserEmail string
}

type VerifySubscriptionPaymentRequest struct{
    RazorpaySubscriptionId string `json:"razorpay_subscription_id" binding:"required"`
    RazorpayPaymentId string    `json:"razorpay_payment_id" binding:"required"`
    RazorpaySignature string    `json:"razorpay_signature" binding:"required"`
}

type UnsubscribeRequest struct{
	SubId uint64
	CancelReason string	`json:"cancel_reason" binding:"required"`
	CancelAtCycleEnd bool `json:"cancel_at_cycle_end" validate:"required"`
}

type SetProfileImageRequest struct{
	UserId uint64
	ContentType string
	Image []byte
}
type GetProfileInformationRequest struct{
	UserId uint64
}

type EditProfile struct{
	Name *string	`json:"name"`
	Bio *string	`json:"bio"`
	Links *string `json:"links"`
}
type ChangePassword struct{
	UserID uint64
	OldPassword	string	`json:"old_password" validate:"required"`
	NewPassword        string `json:"new_password" binding:"required,min=3,max=30"`
	ConfirmNewPassword string `json:"confirm_new_password" binding:"required,eqfield=NewPassword"`
}
type SearchUser struct{
	SearchText string `json:"search_text"`
	Limit int
	Offset int
}
type GetPublicProfile struct{
	UserID uint64
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
	Payment      PaymentWrapper      `json:"payment"` // Add this!
}

type SubscriptionWrapper struct {
	Entity SubscriptionEntity `json:"entity"`
	
}
type PaymentWrapper struct {
    Entity PaymentEntity `json:"entity"`
}
// type SubscriptionEntity struct {
// 	ID           string            `json:"id"`
// 	Entity       string            `json:"entity"`
// 	PlanID       string            `json:"plan_id"`
// 	CustomerID   string            `json:"customer_id"`
// 	Status       string            `json:"status"` // "completed", "active", etc.
// 	CurrentStart int64             `json:"current_start"`
// 	CurrentEnd   int64             `json:"current_end"`
// 	EndedAt      int64             `json:"ended_at"`
// 	Notes        map[string]string `json:"notes"` // Useful for passing Email/UserID
// }
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
type PaymentEntity struct {
    ID             string `json:"id"`
    Amount         int64  `json:"amount"` // Amount in paise (e.g., 50000 for ₹500)
    Currency       string `json:"currency"`
    Status         string `json:"status"`
    Method         string `json:"method"` // card, upi, etc.
    InvoiceID      string `json:"invoice_id"`
    ExternalID     string `json:"external_id"`
}

type WebhookSubscriptionActivatedRequest struct{
	RazorpaySubscriptionId string
	Status string
	PaidCount int
	RemainingCount int
	StartAt        int64             `json:"start_at"`         // Use int64 for Unix timestamps
    EndAt          int64             `json:"end_at"`
	UserID uint64
}
type WebhookSubscriptionChargedRequest struct{
	RazorpaySubscriptionId string
	RazorpayPlanId string
	CurrentEnd int64 `json:"current_end"`
	InvoiceID      string `json:"invoice_id"`
	Amount         int64  `json:"amount"` // Amount in paise (e.g., 50000 for ₹500)
    Currency       string `json:"currency"`
	Method         string `json:"method"` // card, upi, etc.
    Status         string `json:"status"`
    CreatedAt int64    `json:"created_at"`
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
	CancelledAt int64
	UserId uint64
}
type WebhookSubscriptionCompletedRequest struct{
	RazorpaySubscriptionId string
	Status string
	UserId uint64
}
type GetSubscriptionDetails struct{
	UserID uint64
	//SubId uint64
}
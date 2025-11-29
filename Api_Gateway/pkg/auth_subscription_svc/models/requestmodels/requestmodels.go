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
}

type VerifySubscriptionPaymentRequest struct{
    RazorpaySubscriptionId string `json:"razorpay_subscription_id" binding:"required"`
    RazorpayPaymentId string    `json:"razorpay_payment_id" binding:"required"`
    RazorpaySignature string    `json:"razorpay_signature" binding:"required"`
}

type UnsubscribeRequest struct{
	SubId uint64
	CancelReason string	`json:"cancel_reason" binding:"required"`
}

type WebhookRequest struct {
	Event   string `json:"event"`
	Payload struct {
		Subscription struct {
			ID        string `json:"id"`
			PlanID    string `json:"plan_id"`
			Status    string `json:"status"`
			Amount    int    `json:"amount"`
			Currency  string `json:"currency"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
			Customer  struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"customer"`
		} `json:"subscription"`
	} `json:"payload"`
}

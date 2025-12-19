package requestmodels

import "github.com/golang-jwt/jwt/v5"

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
}

type VerifySubscriptionPaymentRequest struct{
    RazorpaySubscriptionId string
    RazorpayPaymentId string
    RazorpaySignature string
}

type UnsubscribeRequest struct{
	SubId uint64
	CancelReason string	`json:"cancel_reason" binding:"required"`
}
type SetProfileImageRequest struct{
	UserId uint64
    ContentType string
	Image []byte
}

type GetProfileInformationRequest struct{
	UserId uint64
}
// type EditProfile struct{
// 	Name *string	`json:"name"`
// 	Bio *string	`json:"bio"`
// 	Links *string `json:"links"`
// }
// type CheckUserExistsRequest struct{
//     UserId uint64
// }
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
type WebhookRequest struct {
    Event string `json:"event"`
    Payload struct {
        Subscription struct {
            ID string `json:"id"`
        } `json:"subscription"`
    } `json:"payload"`
}

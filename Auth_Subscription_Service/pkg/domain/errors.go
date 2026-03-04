package domain

import (
	"errors"
)

var (
	ErrNotEligible                 = errors.New("user already have a subscription or user should update the payment the payment to continue with the halted subscription or finish the payment of created subsciption")
	ErrSubPlanNotFound             = errors.New("subscription plan not found")
	ErrNoActiveSubscription        = errors.New("user has no active subscription")
	ErrRazorpayCancel              = errors.New("razorpay cancel failed")
	ErrDatabase=errors.New("database error")
	ErrSubNotFound=errors.New("Subscription not found")
	ErrSubCompleted=errors.New("subscription already completed")
	ErrSubCancelled=errors.New("subscription already cancelled")
	ErrNoSubscription=errors.New("no subscription to show")
	ErrAdminAccessTokenFail=errors.New("Failed to generarate access token for admin")
	ErrAdminRefreshTokenFail=errors.New("Failed to generarate refresh token for admin")
	ErrUserAccessTokenFail=errors.New("Failed to generarate access token for user")
	ErrUserRefreshTokenFail=errors.New("Failed to generarate refresh token for user")

	ErrInvalidCredentials              = errors.New("invalid credentials")
	ErrUserNotFound                    = errors.New("user not found")
	ErrUserAlreadyExistsByEmail        = errors.New("user already exists, try again with another email")
	ErrUserAlreadyExistsByUsername     = errors.New("username already taken, try with another username")
	ErrOtpExpired                      = errors.New("otp expired")
	ErrUserNotActive                   = errors.New("Cannot block user, email not verified or user alreday blocked")
	ErrUserNotBlocked                  = errors.New("Cannnot unblock user, unblock allowed for users who are alreday in blocked state")
	ErrBlockedLogin                    = errors.New("User account blocked, cannot login")
	ErrPendingLogin                    = errors.New("Otp Verfication Pending, verfiy otp to login")
	ErrSubscriptionPlanAlreadyActive   = errors.New("Cannot the activate the subscription plan, subscription plan is already active")
	ErrSubscriptionPlanAlreadyDeactive = errors.New("Cannot the deactivate the subscription plan, subscription plan is already deactive")
	ErrNoUsersFound                    = errors.New("No such user ids in the list exist")

	ErrDatabaseConnectionTimeOut=errors.New("Database connection time out")
	ErrSendVerifyOtpToEmail=errors.New("Failed to send verification otp to email address")
	ErrVerifyOtpTokenFail=errors.New("Failed to generarate token for otp verfication")

	// Payment/Plan Errors (for your Razorpay logic)
    ErrInvalidRequest     = errors.New("the request sent to the payment provider was invalid")
    ErrExternalService    = errors.New("payment provider encountered an internal error")
    ErrServiceUnavailable = errors.New("payment provider is temporarily unavailable")
	ErrUnknown=errors.New("unexpected error from razorpay")
	ErrInvalidResponseRazorpay=errors.New("Invalide response format from razorpay")

	ErrContentTypeNil=errors.New("Content type is nil")
	ErrS3UploadFail=errors.New("profile image updload failed")

	ErrPasswordMismatch=errors.New("passwords does not match")
)

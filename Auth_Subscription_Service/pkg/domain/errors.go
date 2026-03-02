package domain

import (
	"errors"
)

var (
	RazorpaySubscriptionIdNotFound = errors.New("Razorpay subscription id does not exist")
	ErrNotEligible                 = errors.New("user already have a subscription or user should update the payment the payment to continue with the halted subscription or finish the payment of created subsciption")
	ErrSubPlanNotFound             = errors.New("subscription plan id not found")
	ErrNoActiveSubscription        = errors.New("user has no active subscription")
	ErrRazorpayCancel              = errors.New("razorpay cancel failed")
	ErrDatabase=errors.New("database error")
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
)

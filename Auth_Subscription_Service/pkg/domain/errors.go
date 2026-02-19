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
)

package domain

import "errors"

var (
	RazorpaySubscriptionIdNotFound=errors.New("Razorpay subscription id does not exist")
	ErrNotEligible=errors.New("user already have a subscription or user should update the payment the payment to continue with the halted subscription or finish the payment of created subsciption")
)
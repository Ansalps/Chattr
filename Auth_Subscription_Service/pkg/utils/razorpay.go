package utils

import (
	"fmt"

	"github.com/razorpay/razorpay-go"
)

// type RazorpayUtil struct{
// 	RazorpayClient *razorpay.Client
// }

func NewRazorpayClient(KeyId string, KeySecret string) *razorpay.Client {
	RazorpayClient := razorpay.NewClient(KeyId, KeySecret)
	return RazorpayClient
}

func RazorpayCreatePlan(razorpayClient *razorpay.Client, planData map[string]interface{}) (map[string]interface{}, error) /*(*domain.SubscriptionPlan,error)*/ {

	plan, err := razorpayClient.Plan.Create(planData, map[string]string{})
	if err != nil {
		return nil, err
	}
	fmt.Println(plan)
	return plan, err

}

func RazorpayCreateSubscription(razorpayClient *razorpay.Client, subscriptionData map[string]interface{}) (map[string]interface{}, error) {
	subscription, err := razorpayClient.Subscription.Create(subscriptionData, map[string]string{})
	if err != nil {
		return nil, err
	}
	fmt.Println(subscription)
	return subscription, nil
}


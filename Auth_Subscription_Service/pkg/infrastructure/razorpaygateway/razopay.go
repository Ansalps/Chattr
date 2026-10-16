package razorpaygateway

import (
	"fmt"
	"log"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils"
	"github.com/razorpay/razorpay-go"
)

type RazorpayGateway struct {
	Client *razorpay.Client
}

func NewRazorpayGateway(keyID, secret string) *RazorpayGateway {
	return &RazorpayGateway{
		Client: razorpay.NewClient(keyID, secret),
	}
}

func (r *RazorpayGateway) CreatePlan(planData map[string]interface{}) (*domain.CreatedPlanDTO, error) {
	plan, err := r.Client.Plan.Create(planData, nil)
	if err != nil {
		return nil, utils.MapRazorpayError(err) // Clean and readable!
	}

	// Helper to safely extract nested "item" map
	item, ok := plan["item"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: %v",domain.ErrInvalidResponseRazorpay,"item block plan missing or different type")
	}

	// Full Mapping Logic
	dto := &domain.CreatedPlanDTO{
		ID:          getString(plan, "id"),
		Period:      getString(plan, "period"),
		Interval:    getFloat(plan, "interval"),
		Name:        getString(item, "name"),
		Amount:      int64(getFloat(item, "amount")),
		Currency:    getString(item, "currency"),
		Description: getString(item, "description"),
		Active:      getBool(item, "active"),
	}

	return dto, nil
}

// --- Internal Safety Helpers ---

func getString(m map[string]interface{}, key string) string {
	val, ok := m[key].(string)
	if !ok {
		log.Printf("not a string in getString")
	}
	return val
}

func getFloat(m map[string]interface{}, key string) float64 {
	val, ok := m[key].(float64)
	if !ok {
		log.Println("not a float64 in getFloat")
	}
	return val
}

func getBool(m map[string]interface{}, key string) bool {
	val, ok := m[key].(bool)
	if !ok {
		log.Println("not a bool in getFlat")
	}
	return val
}
func (r *RazorpayGateway) CreateSubscription(subData map[string]interface{}) (*domain.CreatedSubscriptionDTO, error) {
	sub, err := r.Client.Subscription.Create(subData, nil)
	if err != nil {
		return nil, utils.MapRazorpayError(err)
	}
	//fmt.Println(sub)
	// Standard helper to handle the map[string]interface{} types
	return &domain.CreatedSubscriptionDTO{
		ID:             getString(sub, "id"),
		PlanID:         getString(sub, "plan_id"),
		Status:         getString(sub, "status"),
		TotalCount:     int(getFloat(sub, "total_count")),
		RemainingCount: int(getFloat(sub, "remaining_count")),
		PaidCount:      int(getFloat(sub, "paid_count")),
		ShortURL:       getString(sub, "short_url"), // VERY IMPORTANT: Give this to the user!
	}, nil
}

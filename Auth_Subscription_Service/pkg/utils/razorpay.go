package utils

import (
	"encoding/json"
	"fmt"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
)

func MapRazorpayError(err error) error {
	if err == nil {
		return nil
	}

	// Define the Razorpay error structure
	type razorpayErrorResponse struct {
		Error struct {
			Code        string            `json:"code"`
			Description string            `json:"description"`
			Metadata    map[string]string `json:"metadata"`
		} `json:"error"`
	}

	var rzErr razorpayErrorResponse

	// Try to unmarshal. If it fails, it's likely a network/string error.
	if parseErr := json.Unmarshal([]byte(err.Error()), &rzErr); parseErr != nil {
		return fmt.Errorf("%w: %v", domain.ErrExternalService, err)
	}

	// Now map the codes
	switch rzErr.Error.Code {
	case "BAD_REQUEST_ERROR":
		return fmt.Errorf("%w: %s", domain.ErrInvalidRequest, rzErr.Error.Description)
	case "GATEWAY_ERROR":
		return fmt.Errorf("%w: %s", domain.ErrServiceUnavailable, rzErr.Error.Description)
	case "SERVER_ERROR":
		return fmt.Errorf("%w: %s", domain.ErrExternalService, rzErr.Error.Description)
	default:
		return fmt.Errorf("%w: %s", domain.ErrUnknown, rzErr.Error.Description)
	}
}

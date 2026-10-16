package interfacesSmtp

import (
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/responsemodels"
)

type Smtp interface {
	SendVerifcationEmailWithOtp(otp int, recieverEmail string, recieverName string) error
	SendResetPasswordEmailOtp(otp int, recieverEmail string) error
	SendNotificationEmailForResubscribing(webhookReq requestmodels.RazorpayEvent) (responsemodels.WebhookResponse, error)
}

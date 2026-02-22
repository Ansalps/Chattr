package smtp

import (
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/responsemodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils/smtp/interfacesSmtp"
)

type SmtpCredentials struct {
	SmtpConfig *config.Smtp
}

func NewSmtpUtil(smtpConfigs *config.Smtp) interfacesSmtp.Smtp {
	return &SmtpCredentials{
		SmtpConfig: smtpConfigs,
	}
}

func (sc *SmtpCredentials) SendVerifcationEmailWithOtp(otp int, recieverEmail string, recieverName string) error {
	from := sc.SmtpConfig.SmtpSender
	password := sc.SmtpConfig.SmtpPassword
	to := []string{recieverEmail}
	smtpHost := sc.SmtpConfig.SmtpHost
	smtpPort := sc.SmtpConfig.SmtpPort

	subject := "Verify Your Email Address for Chattr"
	body := fmt.Sprintf("Hello,%s\n\nThank you for signing up for Chattr. To complete your registration and ensure the security of your account, please verify your email address by entering the One-Time Password (OTP) provided below:\n\nOTP: %d\n\nPlease use the OTP to verify your email address on our platform within the next 10 minutes. After this time, the OTP will expire, and you will need to request a new one.\n\nIf you did not request this verification, please disregard this email.\n\nIf you need any assistance or have questions, feel free to reach out to our support team at support@example.com.\n\nThank you for choosing Chattr.\n\nBest regards,\nThe Chattr Team", recieverName, otp)
	message := []byte("Subject: " + subject + "\r\n" +
		"\r\n" +
		body)

	// Create authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send actual message
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		//fmt.Println("-----", err)
		return err
	}
	return nil
}

func (sc *SmtpCredentials) SendResetPasswordEmailOtp(otp int, recieverEmail string) error {
	from := sc.SmtpConfig.SmtpSender
	password := sc.SmtpConfig.SmtpPassword
	to := []string{recieverEmail}
	smtpHost := sc.SmtpConfig.SmtpHost
	smtpPort := sc.SmtpConfig.SmtpPort

	subject := "Reset Your Password"
	body := fmt.Sprintf("Dear %s,\n\nYou recently requested to reset your password for your Chattr account. To complete the process, please use the following One-Time Password (OTP):\n\nOTP: %d\n\nThis OTP is valid for 10 minutes. Please do not share this OTP with anyone for security reasons. If you did not request a password reset, please ignore this email.\n\nThank you,\nThe Chattr Team", recieverEmail, otp)

	message := []byte("Subject: " + subject + "\r\n" +
		"\r\n" +
		body)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send actual message
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		//fmt.Println("-----", err)
		return err
	}
	return nil
}

func (sc *SmtpCredentials) SendNotificationEmailForResubscribing(webhookReq requestmodels.RazorpayEvent) (responsemodels.WebhookResponse, error) {
	from := sc.SmtpConfig.SmtpSender
	password := sc.SmtpConfig.SmtpPassword
	userEmail := webhookReq.Payload.Subscription.Entity.Notes["email"]
	to := []string{userEmail}
	smtpHost := sc.SmtpConfig.SmtpHost
	smtpPort := sc.SmtpConfig.SmtpPort
	userName := webhookReq.Payload.Subscription.Entity.Notes["user_name"]

	// 1. Convert Unix timestamps to readable dates
	// Razorpay usually sends EndedAt as 0 if the webhook is 'subscription.completed'
	// because the period just finished. You might want to use CurrentEnd if EndedAt is 0.
	expiryTime := webhookReq.Payload.Subscription.Entity.EndedAt
	if expiryTime == 0 {
		expiryTime = webhookReq.Payload.Subscription.Entity.CurrentEnd
	}

	expiryDate := time.Unix(expiryTime, 0).Format("January 02, 2026")
	deadlineDate := time.Unix(expiryTime+604800, 0).Format("January 02, 2026")

	subject := "Action Required: Renew your Blue Tick Verification"

	// 2. Build a cleaner message with proper RFC 822 headers
	header := make(map[string]string)
	header["From"] = from
	header["To"] = userEmail
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""

	var msgBody string
	for k, v := range header {
		msgBody += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	body := fmt.Sprintf(
		"Hello %s,\n\n"+
			"Your blue tick verification subscription ended on %s.\n\n"+
			"To maintain your verified status and keep your blue tick active, please re-subscribe by %s.\n\n"+
			"If you have any questions, reach out to us at support@chattr.com.\n\n"+
			"Best regards,\n"+
			"The Chattr Team",
		userName, expiryDate, deadlineDate,
	)

	fullMessage := []byte(msgBody + "\r\n" + body)

	// Authenticate and Send
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, fullMessage)
	if err != nil {
		log.Printf("SMTP Error: %v", err)
		return responsemodels.WebhookResponse{}, err
	}

	return responsemodels.WebhookResponse{
		Event:                  webhookReq.Event,
		RazorpaySubscriptionId: webhookReq.Payload.Subscription.Entity.ID,
	}, nil
}

package smtp

import (
	"fmt"
	"net/smtp"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
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
		fmt.Println("-----", err)
		return err
	}
	return nil
}

func (sc *SmtpCredentials)SendResetPasswordEmailOtp(otp int,recieverEmail string)error{
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
		fmt.Println("-----", err)
		return err
	}
	return nil
}
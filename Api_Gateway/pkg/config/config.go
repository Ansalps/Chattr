package config

import "github.com/spf13/viper"

type Config struct {
	Port                   string `mapstructure:"PORT"`
	AuthSubscriptionSvcUrl string `mapstructure:"AUTH_SUBSCRIPTION_SVC_URL"`
	PostAndRelSvcUrl       string `mapstructure:"POST_AND_REL_SVC_URL"`
	ChatSvcUrl             string `mapstructure:"CHAT_SVC_URL"`
	NotificationSvcUrl     string `mapstructure:"NOTIFICATION_SVC_URL"`
	Token Token
	Razorpay Razorpay
}
type Token struct{
	UserSecurityKey	string `mapstructure:"USER_SECURITY_KEY"`
	AdminSecurityKey	string `mapstructure:"ADMIN_SECURITY_KEY"`
	OtpVerificationSecurityKey string `mapstructure:"OTPVERIFICATION_SECURITY_KEY"`
	ResetPasswordSecurityKey string `mapstructure:"RESET_PASSWORD_SECURITY_KEY"`
	AdminRefreshKey string	`mapstructure:"ADMIN_REFRESH_KEY"`
	UserRefreshKey string	`mapstructure:"USER_REFRESH_KEY"`
}

type Razorpay struct{
	KeyId string	`mapstructure:"KEY_ID"`
	KeySecret string	`mapstructure:"KEY_SECRET"`
	//WebhookSecret string `mapstructure:"WEBHOOK_SECRET"`
}

func LoadConfig() (*Config, error) {
	var c Config
	var token Token
	var razorpay Razorpay
	viper.AddConfigPath("./pkg/config")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&token)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&razorpay)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}
	c.Token=token
	c.Razorpay=razorpay
	return &c, nil
}

package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port                   string `mapstructure:"PORT"`
	AuthSubscriptionSvcUrl string `mapstructure:"AUTH_SUBSCRIPTION_SVC_URL"`
	PostRelationSvcUrl     string `mapstructure:"POST_RELATION_SVC_URL"`
	ChatSvcUrl             string `mapstructure:"CHAT_SVC_URL"`
	NotificationSvcUrl     string `mapstructure:"NOTIFICATION_SVC_URL"`
	MaxFileNumber          string `mapstructure:"MAX_FILE_NUMBER"`
	ProfileImgSize         string `mapstructure:"PROFILE_IMG_SIZE"`
	PostSize               string `mapstructure:"POST_SIZE"`
	AuthSource             string `mapstructure:"AUTH_SOURCE"`
	Token                  Token
	Razorpay               Razorpay
	Cloudinary             Cloudinary
	Redis                  Redis
}
type Token struct {
	UserSecurityKey            string `mapstructure:"USER_SECURITY_KEY"`
	AdminSecurityKey           string `mapstructure:"ADMIN_SECURITY_KEY"`
	OtpVerificationSecurityKey string `mapstructure:"OTPVERIFICATION_SECURITY_KEY"`
	ResetPasswordSecurityKey   string `mapstructure:"RESET_PASSWORD_SECURITY_KEY"`
	AdminRefreshKey            string `mapstructure:"ADMIN_REFRESH_KEY"`
	UserRefreshKey             string `mapstructure:"USER_REFRESH_KEY"`
}

type Razorpay struct {
	KeyId         string `mapstructure:"RAZORPAY_KEY_ID"`
	KeySecret     string `mapstructure:"RAZORPAY_KEY_SECRET"`
	WebhookSecret string `mapstructure:"RAZORPAY_WEBHOOK_SECRET"`
}
type Cloudinary struct {
	CloundName string `mapstructure:"CLOUDINARY_CLOUD_NAME"`
	ApiKey     string `mapstructure:"CLOUDINARY_API_KEY"`
	ApiSecret  string `mapstructure:"CLOUDINARY_API_SECRET"`
}
type Redis struct {
	Address string `mapstructure:"REDIS_ADDRESS"`
}

func LoadConfig() (*Config, error) {
	var config Config
	// var token Token
	// var razorpay Razorpay
	// var cloudinary Cloudinary
	// var redis Redis
	viper.AddConfigPath("./pkg/config")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		//return nil, err
		// Check if the error is specifically that the file is missing
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// This is OK! In Kubernetes, we provide variables via the OS environment.
			fmt.Println("Config file not found; falling back to environment variables.")
		} else {
			// This is a REAL error (like a syntax error in the config)
			return nil, fmt.Errorf("fatal error config file: %w", err)
		}
	}

	// Bind all known keys from your structs so Viper knows to look for them in Env
	allKeys := []string{
		"PORT", "AUTH_SUBSCRIPTION_SVC_URL", "POST_RELATION_SVC_URL", "CHAT_SVC_URL", "NOTIFICATION_SVC_URL",
		"MAX_FILE_NUMBER", "PROFILE_IMG_SIZE", "POST_SIZE", "AUTH_SOURCE",
		"USER_SECURITY_KEY", "ADMIN_SECURITY_KEY",
		"OTPVERIFICATION_SECURITY_KEY", "RESET_PASSWORD_SECURITY_KEY",
		"ADMIN_REFRESH_KEY", "USER_REFRESH_KEY", "RAZORPAY_KEY_ID",
		"RAZORPAY_KEY_SECRET", "RAZORPAY_WEBHOOK_SECRET","CLOUDINARY_CLOUD_NAME", "CLOUDINARY_API_KEY",
		"CLOUDINARY_API_SECRET", "REDIS_ADDRESS",
	}

	for _, key := range allKeys {
		viper.BindEnv(key)
	}

	err = viper.Unmarshal(&config.Token)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config.Razorpay)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config.Cloudinary)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config.Redis)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	// c.Token=token
	// c.Razorpay=razorpay
	// c.Cloudinary=cloudinary
	// c.Redis=redis
	return &config, nil
}

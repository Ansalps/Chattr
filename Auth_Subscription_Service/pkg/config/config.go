package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type PortManager struct {
	RunnerPort         string `mapstructure:"PORT"`
	PostRelationSvcUrl string `mapstructure:"POST_RELATION_SVC_URL"`
}
type Database struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBPort     string `mapstructure:"DB_PORT"`
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
	KeyId     string `mapstructure:"RAZORPAY_KEY_ID"`
	KeySecret string `mapstructure:"RAZORPAY_KEY_SECRET"`
}
type Smtp struct {
	SmtpSender   string `mapstructure:"SMTP_SENDER"`
	SmtpPassword string `mapstructure:"SMTP_APPKEY"`
	SmtpHost     string `mapstructure:"SMTP_HOST"`
	SmtpPort     string `mapstructure:"SMTP_PORT"`
}
type Aws struct {
	AwsRegion          string `mapstructure:"AWS_REGION"`
	AwsAccessKey       string `mapstructure:"AWS_ACCESS_KEY"`
	AwsSecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AwsBucket          string `mapstructure:"AWS_BUCKET"`
}
type Config struct {
	PortMngr PortManager
	DB       Database
	Token    Token
	Smtp     Smtp
	Razorpay Razorpay
	Aws      Aws
}

func LoadConfig() (*Config, error) {
	var config Config
	// var portmngr PortManager
	// var db Database
	// var token Token
	// var smtp Smtp
	// var razorpay Razorpay
	// var aws Aws

	viper.AddConfigPath("./pkg/config")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		//return nil,err
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
		"PORT", "POST_RELATION_SVC_URL", "DB_HOST", "DB_USER", "DB_PASSWORD",
		"DB_NAME", "DB_PORT", "USER_SECURITY_KEY", "ADMIN_SECURITY_KEY",
		"OTPVERIFICATION_SECURITY_KEY", "RESET_PASSWORD_SECURITY_KEY",
		"ADMIN_REFRESH_KEY", "USER_REFRESH_KEY", "RAZORPAY_KEY_ID",
		"RAZORPAY_KEY_SECRET", "SMTP_SENDER", "SMTP_APPKEY", "SMTP_HOST",
		"SMTP_PORT", "AWS_REGION", "AWS_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY", "AWS_BUCKET",
	}

	for _, key := range allKeys {
		viper.BindEnv(key)
	}
	err := viper.Unmarshal(&config.PortMngr)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config.DB)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config.Token)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config.Smtp)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config.Razorpay)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config.Aws)
	if err != nil {
		return nil, err
	}
	//config := Config{PortMngr: portmngr, DB: db, Token: token, Smtp: smtp, Razorpay: razorpay, Aws: aws}
	return &config, nil
}

package config

import "github.com/spf13/viper"


type PortManager struct{
	RunnerPort string `mapstructure:"PORT"`
}
type Database struct{
	DBHost     string `mapstructure:"DB_HOST"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBPort     string `mapstructure:"DB_PORT"`
}
type Token struct{
	UserSecurityKey	string `mapstructure:"USER_SECURITY_KEY"`
	AdminSecurityKey	string `mapstructure:"ADMIN_SECURITY_KEY"`
	OtpVerificationSecurityKey string `mapstructure:"OTPVERIFICATION_SECURITY_KEY"`
	ResetPasswordSecurityKey string `mapstructure:"RESET_PASSWORD_SECURITY_KEY"`
	AdminRefreshKey string `mapstructure:"ADMIN_REFRESH_KEY"`
	UserRefreshKey string `mapstructure:"USER_REFRESH_KEY"`
}
type Razorpay struct{
	KeyId string	`mapstructure:"KEY_ID"`
	KeySecret string	`mapstructure:"KEY_SECRET"`
}
type Smtp struct {
	SmtpSender   string `mapstructure:"SMTP_SENDER"`
	SmtpPassword string `mapstructure:"SMTP_APPKEY"`
	SmtpHost     string `mapstructure:"SMTP_HOST"`
	SmtpPort     string `mapstructure:"SMTP_PORT"`
}
type Config struct{
	PortMngr PortManager
	DB Database
	Token Token
	Smtp Smtp
	Razorpay Razorpay
}

func LoadConfig() (*Config,error){
	var portmngr PortManager
	var db Database
	var token Token
	var smtp Smtp
	var razorpay Razorpay
	
	viper.AddConfigPath("./pkg/config")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err:=viper.ReadInConfig();err!=nil{
		return nil,err
	}
	err := viper.Unmarshal(&portmngr)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&db)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&token)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&smtp)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&razorpay)
	if err != nil {
		return nil, err
	}
	config:=Config{PortMngr: portmngr,DB: db,Token: token,Smtp: smtp,Razorpay: razorpay}
	return &config,nil
}
package config

import "github.com/spf13/viper"

type Config struct {
	Port                   string `mapstructure:"PORT"`
	AuthSubscriptionSvcUrl string `mapstructure:"AUTH_SUBSCRIPTION_SVC_URL"`
	PostAndRelSvcUrl       string `mapstructure:"POST_AND_REL_SVC_URL"`
	ChatSvcUrl             string `mapstructure:"CHAT_SVC_URL"`
	NotificationSvcUrl     string `mapstructure:"NOTIFICATION_SVC_URL"`
}

func LoadConfig() (*Config, error) {
	var c Config
	viper.AddConfigPath("./pkg/config")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

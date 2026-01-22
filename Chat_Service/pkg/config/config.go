package config

import "github.com/spf13/viper"

type PortManager struct {
	RunnerPort string `mapstructure:"PORT"`
	AuthSvcUrl string `mapstructure:"AUTH_SUBSCRIPTION_SVC_URL"`
}
type Database struct {
	MongoDbURI string	`mapstructure:"MONGODB_URI"`
}

type Config struct {
	PortMngr PortManager
	DB       Database
}

func LoadConfig() (*Config, error) {
	var portmngr PortManager
	var db Database

	viper.AddConfigPath("./pkg/config")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	err := viper.Unmarshal(&portmngr)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&db)
	if err != nil {
		return nil, err
	}

	
	config := Config{PortMngr: portmngr, DB: db}
	return &config, nil
}

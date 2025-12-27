package config

import "github.com/spf13/viper"

type PortManager struct {
	RunnerPort string `mapstructure:"PORT"`
	AuthSvcUrl string `mapstructure:"AUTH_SUBSCRIPTION_SVC_URL"`
}
type Database struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBPort     string `mapstructure:"DB_PORT"`
}
type Redis struct{
	Address string	`mapstructure:"REDIS_ADDRESS"`
}
type Config struct {
	PortMngr PortManager
	DB       Database
	Redis Redis
}

func LoadConfig() (*Config, error) {
	var portmngr PortManager
	var db Database
	var redis Redis

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
	err = viper.Unmarshal(&redis)
	if err != nil {
		return nil, err
	}
	
	config := Config{PortMngr: portmngr, DB: db,Redis: redis}
	return &config, nil
}

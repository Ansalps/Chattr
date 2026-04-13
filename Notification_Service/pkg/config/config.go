package config

import (
	"fmt"

	"github.com/spf13/viper"
)

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
type KafkaConfig struct {
	Brokers string `mapstructure:"KAFKA_BROKERS"`
	CACert string	`mapstructure:"KAFKA_CA_CERT"`
	AccessCert string	`mapstructure:"KAFKA_ACCESS_CERT"`
	AccessKey string	`mapstructure:"KAFKA_ACCESS_KEY"`
}
type Config struct {
	PortMngr PortManager
	DB       Database
	Kafka    KafkaConfig
}

func LoadConfig() (*Config, error) {
	var config Config
	// var portmngr PortManager
	// var db Database
	// var kafka KafkaConfig

	viper.AddConfigPath("./pkg/config")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
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
		"PORT", "AUTH_SUBSCRIPTION_SVC_URL","DB_HOST", "DB_USER","DB_PASSWORD",
		"DB_NAME", "DB_PORT", "KAFKA_BROKERS", "KAFKA_CA_CERT", "KAFKA_ACCESS_CERT", "KAFKA_ACCESS_KEY",
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
	err = viper.Unmarshal(&config.Kafka)
	if err != nil {
		return nil, err
	}

	//config := Config{PortMngr: portmngr, DB: db,Kafka: kafka}
	return &config, nil
}

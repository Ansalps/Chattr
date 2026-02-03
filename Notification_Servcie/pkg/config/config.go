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
type KafkaConfig struct {
	Brokers string `mapstructure:"KAFKA_BROKERS"`
}
type Config struct {
	PortMngr PortManager
	DB       Database
	Kafka        KafkaConfig
}

func LoadConfig() (*Config, error) {
	var portmngr PortManager
	var db Database
	var kafka KafkaConfig

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
	err = viper.Unmarshal(&kafka)
	if err != nil {
		return nil, err
	}
	
	
	config := Config{PortMngr: portmngr, DB: db,Kafka: kafka}
	return &config, nil
}

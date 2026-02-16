package config

import "github.com/spf13/viper"

type PortManager struct {
	RunnerPort string `mapstructure:"PORT"`
	AuthSvcUrl string `mapstructure:"AUTH_SUBSCRIPTION_SVC_URL"`
}
type Database struct {
	MongoDbURI string	`mapstructure:"MONGODB_URI"`
}
type KafkaConfig struct {
	Brokers string `mapstructure:"KAFKA_BROKERS"`
}

type Config struct {
	AuthSource 	string `mapstructure:"AUTH_SOURCE"`
	PortMngr PortManager
	DB       Database
	Kafka        KafkaConfig
}

func LoadConfig() (*Config, error) {
	var c Config
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
	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}
	
	c.PortMngr=portmngr
	c.DB=db
	c.Kafka=kafka
	return &c, nil
}

package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type PortManager struct {
	RunnerPort string `mapstructure:"PORT"`
	AuthSvcUrl string `mapstructure:"AUTH_SUBSCRIPTION_SVC_URL"`
}
type Database struct {
	MongoDbURI string `mapstructure:"MONGODB_URI"`
}
type KafkaConfig struct {
	Brokers    string `mapstructure:"KAFKA_BROKERS"`
	CACert     string `mapstructure:"KAFKA_CA_CERT"`
	AccessCert string `mapstructure:"KAFKA_ACCESS_CERT"`
	AccessKey  string `mapstructure:"KAFKA_ACCESS_KEY"`
}
type Aws struct {
	AwsRegion          string `mapstructure:"AWS_REGION"`
	AwsAccessKey       string `mapstructure:"AWS_ACCESS_KEY"`
	AwsSecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AwsBucket          string `mapstructure:"AWS_BUCKET"`
}

type Config struct {
	AuthSource string `mapstructure:"AUTH_SOURCE"`
	MongoDBUri string `mapstructure:"MONGODB_URI"`
	PortMngr   PortManager
	DB         Database
	Kafka      KafkaConfig
	Aws        Aws
}

func LoadConfig() (*Config, error) {
	var config Config
	// var portmngr PortManager
	// var db Database
	// var kafka KafkaConfig
	// var aws Aws

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
		"AUTH_SOURCE", "PORT", "AUTH_SUBSCRIPTION_SVC_URL", "MONGODB_URI", "KAFKA_BROKERS",
		"AWS_REGION", "AWS_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY", "AWS_BUCKET",
		"KAFKA_CA_CERT", "KAFKA_ACCESS_CERT", "KAFKA_ACCESS_KEY",
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
	err = viper.Unmarshal(&config.Aws)
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	caData, err := os.ReadFile(config.Kafka.CACert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert at %s: %w", config.Kafka.CACert, err)
	}
	config.Kafka.CACert = string(caData)

	// Read Access Cert
	accessCertData, err := os.ReadFile(config.Kafka.AccessCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read Access cert at %s: %w", config.Kafka.AccessCert, err)
	}
	config.Kafka.AccessCert = string(accessCertData)

	// Read Access Key
	accessKeyData, err := os.ReadFile(config.Kafka.AccessKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read Access key at %s: %w", config.Kafka.AccessKey, err)
	}
	config.Kafka.AccessKey = string(accessKeyData)

	// c.PortMngr = portmngr
	// c.DB = db
	// c.Kafka = kafka
	// c.Aws = aws
	return &config, nil
}

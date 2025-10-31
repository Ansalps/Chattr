package config

import "github.com/spf13/viper"

type Config struct {
	Port       string `mapstructure:"PORT"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBPort     string `mapstructure:"DB_PORT"`
	JwtKey	string `mapstructure:"JWT_KEY"`
}

func LoadConfig() (*Config,error){
	var config Config
	viper.AddConfigPath("./pkg/config")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err:=viper.ReadInConfig();err!=nil{
		return nil,err
	}
	if err:=viper.Unmarshal(&config);err!=nil{
		return nil,err
	}
	return &config,nil
}
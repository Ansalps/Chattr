package db

import (
	"fmt"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDatabase(cfg *config.Config) (*gorm.DB, error) {
	connectionString := fmt.Sprintf("port=%s host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Port, cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("Error connceting to auth_subscription database: %v", err)
	}
	err = db.AutoMigrate(&domain.Admin{})
	if err != nil {
		return nil, fmt.Errorf("Error in automigrating the table: %v", err)
	}

	return db, nil
}

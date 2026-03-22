package db

import (
	"fmt"
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// func ConnectDatabase(cfg *config.Config) (*gorm.DB, error) {
// 	connectionString := fmt.Sprintf("port=%s host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
// 		cfg.PortMngr.RunnerPort, cfg.DB.DBHost, cfg.DB.DBUser, cfg.DB.DBPassword, cfg.DB.DBName, cfg.DB.DBPort)
// 	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})

// 	if err != nil {
// 		return nil, fmt.Errorf("Error connceting to auth_subscription database: %v", err)
// 	}
// 	err = db.AutoMigrate(&domain.Admin{}, &domain.User{}, &domain.Otp{}, &domain.SubscriptionPlan{}, &domain.UserSubscription{}, &domain.SubscriptionPayment{})
// 	if err != nil {
// 		return nil, fmt.Errorf("Error in automigrating the table: %v", err)
// 	}

//		return db, nil
//	}
func ConnectDatabase(cfg *config.Config) (*gorm.DB, error) {

	connectionString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DB.DBHost,
		cfg.DB.DBUser,
		cfg.DB.DBPassword,
		cfg.DB.DBName,
		cfg.DB.DBPort,
	)

	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error connecting to auth_subscription database: %v", err)
	}

	// ✅ Get underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// ✅ Connection Pool settings
	sqlDB.SetMaxOpenConns(25)                  // max connections to DB
	sqlDB.SetMaxIdleConns(10)                  // idle connections
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // reuse time
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)  // idle timeout

	// ✅ Auto migrate
	err = db.AutoMigrate(
		&domain.Admin{},
		&domain.User{},
		&domain.Otp{},
		&domain.SubscriptionPlan{},
		&domain.UserSubscription{},
		&domain.SubscriptionPayment{},
	)

	if err != nil {
		return nil, fmt.Errorf("error in automigrate: %v", err)
	}

	return db, nil
}

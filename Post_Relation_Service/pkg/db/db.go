package db

import (
	"fmt"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDatabase(cfg *config.Config) (*gorm.DB, error) {
	connectionString := fmt.Sprintf("port=%s host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.PortMngr.RunnerPort, cfg.DB.DBHost, cfg.DB.DBUser, cfg.DB.DBPassword, cfg.DB.DBName, cfg.DB.DBPort)
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("Error connceting to auth_subscription database: %v", err)
	}
	err = db.AutoMigrate(&domain.Post{}, &domain.PostMedia{},&domain.PostLike{},&domain.Comment{},&domain.Relation{})
	if err != nil {
		return nil, fmt.Errorf("Error in automigrating the table: %v", err)
	}

	return db, nil
}

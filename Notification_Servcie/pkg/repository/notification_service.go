package repository

import (
	"github.com/Ansalps/Chattr_Notification_Service/pkg/domain"
	interfacesrepository "github.com/Ansalps/Chattr_Notification_Service/pkg/repository/interfacesRepository"
	"gorm.io/gorm"
)

type NotificationRepository struct{
	DB *gorm.DB
}

func NewNotificationRepository(db *gorm.DB)interfacesrepository.NotificationRepository{
	return &NotificationRepository{
		DB: db,
	}
}
func (ad *NotificationRepository) SaveNotification(notification domain.Notification)error{
	if err:=ad.DB.Create(&notification).Error; err!=nil{
		return err
	}
	return nil
}
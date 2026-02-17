package repository

import (
	"github.com/Ansalps/Chattr_Notification_Service/pkg/domain"
	interfacesrepository "github.com/Ansalps/Chattr_Notification_Service/pkg/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/requestmodels"
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

func (ad *NotificationRepository) GetAllNotifications(req requestmodels.GetNotificationsequest) ([]domain.Notification, error) {
	var resp []domain.Notification
	query := `
		SELECT id, user_id, actor_id, target_id, type, message, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	result:= ad.DB.Raw(query,req.UserID,req.Limit,req.Offset).Scan(&resp)
	if result.Error != nil {
		return nil, result.Error
	}
	// if result.RowsAffected==0{
	// 	return nil,gorm.ErrRecordNotFound
	// }

	return resp, nil
}
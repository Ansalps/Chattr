package interfacesrepository

import (
	"github.com/Ansalps/Chattr_Notification_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/requestmodels"
)

type NotificationRepository interface {
	SaveNotification(domain.Notification) error
	GetAllNotifications(req requestmodels.GetNotificationsequest) ([]domain.Notification, error)
}

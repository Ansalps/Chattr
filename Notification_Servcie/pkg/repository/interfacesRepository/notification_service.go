package interfacesrepository

import "github.com/Ansalps/Chattr_Notification_Service/pkg/domain"

type NotificationRepository interface{
	SaveNotification(domain.Notification)error
}
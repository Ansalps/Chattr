package interfacesUsecase

import (
	"github.com/Ansalps/Chattr_Notification_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/requestmodels"
)

type NotificationUsecase interface{
	ProcessLikeEvent(event requestmodels.PostEvent)error
	ProcessCommentEvent(event requestmodels.PostEvent)error
	ProcessFollowEvent(event requestmodels.UserEvent) error
	ProcessDirectMessageEvent(event requestmodels.DirectMessageEvent)error
	ProcessGroupMessageEvent(event requestmodels.GroupMessageEvent)error
	GetAllNotifications(req requestmodels.GetNotificationsequest)([]domain.Notification,error)
}

// type NotificationEventListener interface {
//     ProcessLikeEvent(event requestmodels.PostEvent)
// 	ProcessCommentEvent(event requestmodels.PostEvent)
// }
package interfacesUsecase

import (
	"github.com/Ansalps/Chattr_Notification_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/requestmodels"
)

type NotificationUsecase interface{
	ProcessLikeEvent(event requestmodels.PostEvent)
	ProcessCommentEvent(event requestmodels.PostEvent)
	ProcessFollowEvent(event requestmodels.UserEvent) 
	ProcessDirectMessageEvent(event requestmodels.DirectMessageEvent)
	ProcessGroupMessageEvent(event requestmodels.GroupMessageEvent)
	GetAllNotifications(req requestmodels.GetNotificationsequest)([]domain.Notification,error)
}

// type NotificationEventListener interface {
//     ProcessLikeEvent(event requestmodels.PostEvent)
// 	ProcessCommentEvent(event requestmodels.PostEvent)
// }
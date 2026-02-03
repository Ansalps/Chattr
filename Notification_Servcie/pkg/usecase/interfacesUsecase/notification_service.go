package interfacesUsecase

import "github.com/Ansalps/Chattr_Notification_Service/pkg/requestmodels"

type NotificationUsecase interface{
	ProcessLikeEvent(event requestmodels.PostEvent)
	ProcessCommentEvent(event requestmodels.PostEvent)
	ProcessFollowEvent(event requestmodels.UserEvent) 
}

// type NotificationEventListener interface {
//     ProcessLikeEvent(event requestmodels.PostEvent)
// 	ProcessCommentEvent(event requestmodels.PostEvent)
// }
package routes

import (
	"github.com/Ansalps/Chattr_Notification_Service/pkg/handler"
	"github.com/gin-gonic/gin"
)

func NotificationServiceRoutes(router *gin.Engine, notificationHandler *handler.NotificationHandler) {
	 router.GET("/user/notification/ws", notificationHandler.WebSocketConnection)
	// router.POST("/user/group", chatHandler.CreateGroup)
	// router.POST("/user/group/add-members", chatHandler.AddMembers)
	// router.DELETE("/user/group/remove-member", chatHandler.RemoveMember)
	// router.GET("/user/get-recent-chat-profiles", chatHandler.GetRecentChatProfiles)
	 router.GET("/user/notifications", notificationHandler.GetAllNotifications)
}

package routes

import (
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/middleware"
	notificationServiceHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/notification_svc/handler"
	"github.com/gin-gonic/gin"
)

func NotificationRoutes(router *gin.Engine, notificationHandler *notificationServiceHandler.NotificationHandler, cfg *config.Config, authMiddleware *middleware.AuthMiddleware) {
	router.GET("/user/notification/ws", authMiddleware.VerifyJwt([]string{"user"}, "access", cfg.Token.UserSecurityKey), notificationHandler.WebSocketConnection)
	// router.POST("/user/group", authMiddleware.VerifyJwt([]string{"user"}, "access", cfg.Token.UserSecurityKey), chatHandler.CreateGroup)
	// router.POST("/user/group/:group_id/add-members", authMiddleware.VerifyJwt([]string{"user"}, "access", cfg.Token.UserSecurityKey), chatHandler.AddMembers)
	// router.DELETE("/user/group/:group_id/remove-member/:member_id", authMiddleware.VerifyJwt([]string{"user"}, "access", cfg.Token.UserSecurityKey), chatHandler.RemoveMember)
	// router.GET("/user/get-recent-chat-profiles", authMiddleware.VerifyJwt([]string{"user"}, "access", cfg.Token.UserSecurityKey), chatHandler.RecentChatProfiles)
	// router.GET("/user/chat/:conv_id", authMiddleware.VerifyJwt([]string{"user"}, "access", cfg.Token.UserSecurityKey), chatHandler.GetChat)
}

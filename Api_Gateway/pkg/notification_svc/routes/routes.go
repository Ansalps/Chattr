package routes

import (
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/middleware"
	notificationServiceHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/notification_svc/handler"
	"github.com/gin-gonic/gin"
)

func NotificationRoutes(router *gin.Engine, notificationHandler *notificationServiceHandler.NotificationHandler, cfg *config.Config, authMiddleware *middleware.AuthMiddleware) {
	router.GET("/user/notification/ws", authMiddleware.VerifyJwt([]string{"user"}, "access", cfg.Token.UserSecurityKey), notificationHandler.WebSocketConnection)
	router.GET("/user/notifications", authMiddleware.VerifyJwt([]string{"user"}, "access", cfg.Token.UserSecurityKey), notificationHandler.GetAllNotifications)
}

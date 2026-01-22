package routes

import (
	chatServiceHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/chat_svc/handler"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func ChatRoutes(router *gin.Engine, chatHandler *chatServiceHandler.ChatHandler, cfg *config.Config, authMiddleware *middleware.AuthMiddleware) {
	router.GET("/user/ws",authMiddleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),chatHandler.WebSocketConnection)
	router.POST("/user/group",authMiddleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),chatHandler.CreateGroup)
	router.POST("/user/group/:group_id/add-members",authMiddleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),chatHandler.AddMembers)
	router.DELETE("/user/group/:group_id/remove-member/:member_id",authMiddleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),chatHandler.RemoveMember)
	router.GET("/user/get-recent-chat-profiles",authMiddleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),chatHandler.RecentChatProfiles)
}

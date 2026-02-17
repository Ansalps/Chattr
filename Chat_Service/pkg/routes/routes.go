package routes

import (
	"github.com/Ansalps/Chattr_Chat_Service/pkg/handler"
	"github.com/gin-gonic/gin"
)

func ChatServiceRoutes(router *gin.Engine, chatHandler *handler.ChatHandler) {
	router.GET("/user/ws", chatHandler.WebSocketConnection)
	router.POST("/user/group",chatHandler.CreateGroup)
	router.POST("/user/group/add-members",chatHandler.AddMembers)
	router.DELETE("/user/group/remove-member",chatHandler.RemoveMember)
	router.GET("/user/get-recent-chat-profiles",chatHandler.GetRecentChatProfiles)
	router.GET("/user/chat/:conv_id",chatHandler.GetChat)
	router.GET("/user/group/:group_id/members",chatHandler.GetGroupMembers)
	router.PATCH("/user/group/:group_id/set-group-profile-image",chatHandler.SetGroupProfileImage)
}

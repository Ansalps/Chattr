package routes

import (
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/middleware"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/handler"
	"github.com/gin-gonic/gin"
)

func PostRelationRoutes(router *gin.Engine,postRelationHandler *handler.PostRelationHandler,cfg *config.Config){
	router.POST("/user/post",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.CreatePost)
}
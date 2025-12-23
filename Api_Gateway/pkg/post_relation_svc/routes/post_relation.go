package routes

import (
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/middleware"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/handler"
	"github.com/gin-gonic/gin"
)

func PostRelationRoutes(router *gin.Engine,postRelationHandler *handler.PostRelationHandler,cfg *config.Config){
	router.POST("/user/post",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.CreatePost)
	router.GET("/user/:user_id/posts",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.FetchAllPosts)
	router.PATCH("/user/post/:post_id",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.EditPost)
	router.DELETE("/user/post/:post_id",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.DeletePost)

	router.POST("/user/post/like/:post_id",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.LikePost)
	router.DELETE("/user/post/like/:post_id",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.UnlikePost)

	router.POST("/user/post/comment/:post_id",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.AddComment)
	router.GET("/user/post/comment/:post_id",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.FetchComments)
	router.PATCH("/user/post/:post_id/comment/:comment_id",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.EditComment)
	router.DELETE("/user/post/:post_id/comment/:comment_id",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.DeleteComment)

	

	router.POST("/user/relation/follow/:following_user_id",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.Follow)
	router.DELETE("/user/relation/unfollow/:unfollowing_user_id",middleware.VerifyJwt([]string{"user"},"access",cfg.Token.UserSecurityKey),postRelationHandler.Unfollow)
}
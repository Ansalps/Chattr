package di

import (
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client"
	authClient "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client"
	authHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/handler"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/repository"
	authRoutes "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/routes"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/middleware"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	postClient "github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/client"
	postRelationHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/handler"
	postRelationRoutes "github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/routes"

	chatServiceHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/chat_svc/handler"
	chatServiceRoutes "github.com/Ansalps/Chattr_Api_Gateway/pkg/chat_svc/routes"

	notificationServiceHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/notification_svc/handler"
	notificationServiceRoutes "github.com/Ansalps/Chattr_Api_Gateway/pkg/notification_svc/routes"
	"github.com/gin-gonic/gin"
)

func DependencyInjection(router *gin.Engine, cfg *config.Config) error {
	redisClient:=client.NewRedisClient(cfg)
	RedisRepository:=repository.NewRedisRepository(redisClient)

	AuthMiddleware:=middleware.NewAuthMiddlware(RedisRepository)

	authSubscriptionClient := authClient.NewAuthSubscriptionClient(cfg)
	authSubClient := authSubscriptionClient.(*client.AuthSubscriptionClient)
	
	postRelationClient := postClient.NewPostRelationClient(cfg)
	postDirectClient:=postRelationClient.(*postClient.PostRelationClient)

	authSubscriptionHandler := authHandler.NewAuthSubscriptionHandler(authSubscriptionClient, cfg, authSubClient,postDirectClient,RedisRepository)
	authRoutes.AuthSubscriptionRoutes(router, authSubscriptionHandler, &cfg.Token,AuthMiddleware)

	postRelationHandler := postRelationHandler.NewPostRelationHandler(postRelationClient, cfg,authSubClient,postDirectClient)
	postRelationRoutes.PostRelationRoutes(router, postRelationHandler, cfg,AuthMiddleware)

	chatHandler:=chatServiceHandler.NewChatHandler(cfg)
	chatServiceRoutes.ChatRoutes(router,chatHandler,cfg,AuthMiddleware)

	notificationHandler:=notificationServiceHandler.NewNotificationHandler(cfg)
	notificationServiceRoutes.NotificationRoutes(router,notificationHandler,cfg,AuthMiddleware)
	return nil
}

package di

import (
	authClient "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client"
	authHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/handler"
	authRoutes "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/routes"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	postRelationClient "github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/client"
	postRelationHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/handler"
	postRelationRoutes "github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/routes"
	"github.com/gin-gonic/gin"
)

func DependencyInjection(router *gin.Engine, cfg *config.Config) error {
	authSubscriptionClient := authClient.NewAuthSubscriptionClient(cfg)
	authSubscriptionHandler := authHandler.NewAuthSubscriptionHandler(authSubscriptionClient,cfg)
	authRoutes.AuthSubscriptionRoutes(router, authSubscriptionHandler, &cfg.Token)

	postRelationClient:=postRelationClient.NewPostRelationClient(cfg)
	postRelationHandler:=postRelationHandler.NewPostRelationHandler(postRelationClient,cfg)
	postRelationRoutes.PostRelationRoutes(router,postRelationHandler,cfg)
	return nil
}

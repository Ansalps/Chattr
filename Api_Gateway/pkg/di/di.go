package di

import (
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client"
	authClient "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client"
	authHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/handler"
	authRoutes "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/routes"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	postClient "github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/client"
	postRelationHandler "github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/handler"
	postRelationRoutes "github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/routes"
	"github.com/gin-gonic/gin"
)

func DependencyInjection(router *gin.Engine, cfg *config.Config) error {
	authSubscriptionClient := authClient.NewAuthSubscriptionClient(cfg)
	authSubClient := authSubscriptionClient.(*client.AuthSubscriptionClient)
	
	postRelationClient := postClient.NewPostRelationClient(cfg)
	postDirectClient:=postRelationClient.(*postClient.PostRelationClient)

	authSubscriptionHandler := authHandler.NewAuthSubscriptionHandler(authSubscriptionClient, cfg, authSubClient,postDirectClient)
	authRoutes.AuthSubscriptionRoutes(router, authSubscriptionHandler, &cfg.Token)

	postRelationHandler := postRelationHandler.NewPostRelationHandler(postRelationClient, cfg,postDirectClient)
	postRelationRoutes.PostRelationRoutes(router, postRelationHandler, cfg)
	return nil
}

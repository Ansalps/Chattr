package di

import (
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/handler"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/routes"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/gin-gonic/gin"
)

func DependencyInjection(router *gin.Engine, cfg *config.Config) error {
	authSubscriptionClient := client.NewAuthSubscriptionClient(cfg)
	authSubscriptionHandler := handler.NewAuthSubscriptionHandler(authSubscriptionClient)
	routes.AuthSubscriptionRoutes(router, authSubscriptionHandler, &cfg.Token)
	return nil
}

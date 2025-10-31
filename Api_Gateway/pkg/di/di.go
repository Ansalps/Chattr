package di

import (
	"fmt"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/handler"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/routes"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/gin-gonic/gin"
)

func DependencyInjection(router *gin.Engine, cfg *config.Config) error {
	authSubscriptionClient := client.NewAuthSubscriptionClient(cfg)
	if authSubscriptionClient == nil {
		return fmt.Errorf("failed to initialize AuthSubscriptionClient")
	}
	fmt.Println("auth_subscription_Client",authSubscriptionClient)
	authSubscriptionHandler := handler.NewAuthSubscriptionHandler(authSubscriptionClient)
	fmt.Println("authSubscriptionHandler",authSubscriptionHandler)
	routes.AuthSubscriptionRoutes(router, authSubscriptionHandler)
	return nil
}

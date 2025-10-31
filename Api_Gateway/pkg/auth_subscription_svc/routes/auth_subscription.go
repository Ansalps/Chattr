package routes

import (
	"fmt"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/handler"
	"github.com/gin-gonic/gin"
)

func AuthSubscriptionRoutes(router *gin.Engine, authSubscriptionHandler *handler.AuthSubscriptionHandler) {
	router.POST("/admin/login", authSubscriptionHandler.AdminLogin)
	fmt.Println("is it reaching in registering routes")
}

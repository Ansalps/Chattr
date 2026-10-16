package di

import (
	"github.com/Ansalps/Chattr_Notification_Service/pkg/client"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/config"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/handler"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/db"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/logger"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/websockethub"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/repository"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/routes"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/usecase"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/usecase/interfacesUsecase"
	"github.com/gin-gonic/gin"
)

func DependencyInjection(router *gin.Engine, cfg *config.Config, hub *websockethub.Hub,log logger.Logger) (interfacesUsecase.NotificationUsecase, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}
	authClient, err := client.InitAuthSubscriptionServiceClient(cfg)
	if err != nil {
		return nil,err
	}
	NotificationRepository := repository.NewNotificationRepository(gormDB)
	NotificationUsecase := usecase.NewNotificationUsecase(NotificationRepository, hub,authClient)
	NotificationHandler := handler.NewNotificationHandler(NotificationUsecase, hub,log)
	routes.NotificationServiceRoutes(router, NotificationHandler)
	return NotificationUsecase, nil
}

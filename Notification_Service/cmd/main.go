package main

import (
	"github.com/Ansalps/Chattr_Notification_Service/pkg/config"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/di"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/kafka"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/logger"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/websockethub"
	"github.com/gin-gonic/gin"
)

func main() {
	log, err := logger.NewZapLogger()
	if err != nil {
		panic(err)
	}
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load configuration:",
			logger.Field{Key: "error", Value: err})
	}
	hub := websockethub.NewHub()
	go hub.Run()
	router := gin.New()
	router.Use(gin.Recovery())
	notificationUsecase, err := di.DependencyInjection(router, config, hub,log)
	if err != nil {
		log.Fatal("Cannot Start server due to failure in DependencyInjectin:",
			logger.Field{Key: "error", Value: err})
	}
	// NEW: Start the bridge between Kafka and WebSockets
	// Use the field from your config (e.g., config.Kafka.Brokers)
	go kafka.StartNotificationConsumer(config.Kafka.Brokers, "post-events", hub, notificationUsecase,log)
	go kafka.StartNotificationConsumer(config.Kafka.Brokers, "user-events", hub, notificationUsecase,log)
	go kafka.StartNotificationConsumer(config.Kafka.Brokers, "chat-events", hub, notificationUsecase,log)

	err = router.Run(config.PortMngr.RunnerPort)
	if err != nil {
		log.Fatal("Error starting server:",
			logger.Field{Key: "error", Value: err})
	}
}

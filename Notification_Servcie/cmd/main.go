package main

import (
	"log"

	"github.com/Ansalps/Chattr_Notification_Service/pkg/config"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/di"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/kafka"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/websockethub"
	"github.com/gin-gonic/gin"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config", err)
	}
	hub := websockethub.NewHub()
	go hub.Run()
	router := gin.New()
	notificationUsecase,err := di.DependencyInjection(router, config, hub)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
	// NEW: Start the bridge between Kafka and WebSockets
	// Use the field from your config (e.g., config.Kafka.Brokers)
	go kafka.StartNotificationConsumer(config.Kafka.Brokers, "post-events", hub,notificationUsecase)
	go kafka.StartNotificationConsumer(config.Kafka.Brokers,"user-events",hub,notificationUsecase)
	go kafka.StartNotificationConsumer(config.Kafka.Brokers,"chat-events",hub,notificationUsecase)

	

	err = router.Run(config.PortMngr.RunnerPort)
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}

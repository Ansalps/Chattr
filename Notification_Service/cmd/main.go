package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load configuration:",
			logger.Field{Key: "error", Value: err})
	}
	hub := websockethub.NewHub()
	go hub.Run()
	router := gin.New()
	router.Use(gin.Recovery())
	notificationUsecase, err := di.DependencyInjection(router, cfg, hub, log)
	if err != nil {
		log.Fatal("Cannot Start server due to failure in DependencyInjectin:",
			logger.Field{Key: "error", Value: err})
	}
	// NEW: Start the bridge between Kafka and WebSockets
	// Use the field from your config (e.g., config.Kafka.Brokers)
	go kafka.StartNotificationConsumer(cfg.Kafka.Brokers, "post-events", hub, notificationUsecase, log)
	go kafka.StartNotificationConsumer(cfg.Kafka.Brokers, "user-events", hub, notificationUsecase, log)
	go kafka.StartNotificationConsumer(cfg.Kafka.Brokers, "chat-events", hub, notificationUsecase, log)

	// err = router.Run(config.PortMngr.RunnerPort)
	// if err != nil {
	// 	log.Fatal("Error starting server:",
	// 		logger.Field{Key: "error", Value: err})
	// }
	srv := &http.Server{
		Addr:    cfg.PortMngr.RunnerPort,
		Handler: router,
	}

	log.Info("Starting Notification Service",
		logger.Field{Key: "port", Value: cfg.PortMngr.RunnerPort})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed",
				logger.Field{Key: "error", Value: err})
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("forced shutdown",
			logger.Field{Key: "error", Value: err})
	}
}

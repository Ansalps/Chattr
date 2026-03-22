package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ansalps/Chattr_Chat_Service/logger"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/config"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/di"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	log, err := logger.NewZapLogger()
	if err != nil {
		panic(err)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load configuration:",
			logger.Field{Key: "error", Value: err})
	}

	err = di.DependencyInjection(router, cfg, log)
	if err != nil {
		log.Fatal("Cannot Start server due to failure in DependencyInjectin:",
			logger.Field{Key: "error", Value: err})
	}
	handler.StartHub(log)
	// err = router.Run(config.PortMngr.RunnerPort)
	// if err != nil {
	// 	log.Fatal("Error starting server:",
	// 		logger.Field{Key: "error", Value: err})
	// }
	srv := &http.Server{
		Addr:    cfg.PortMngr.RunnerPort,
		Handler: router,
	}

	log.Info("Starting Chat Service",
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
	// ✅ stop websocket hub
	handler.StopHub()

	log.Info("shutdown complete")
}

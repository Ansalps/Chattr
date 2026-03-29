package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ansalps/Chattr_Api_Gateway/infrastructure/logger"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/di"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log, err := logger.NewZapLogger()
	if err != nil {
		panic(err)
	}

	// ✅ start pprof server
	go func() {
		log.Info("pprof running on 127.0.0.1:6060")
		if err := http.ListenAndServe("127.0.0.1:6060", nil); err != nil {
			log.Error("pprof server failed",
				logger.Field{Key: "error", Value: err})
		}
	}()

	
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware(log))

	cfg, err := config.LoadConfig()
	if err != nil {
		//log.Fatalf("cannot load configuration: %v", err)
		log.Fatal("cannot load configuration:",
			logger.Field{Key: "error", Value: err})
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5500",
			"http://127.0.0.1:5500",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	router.LoadHTMLGlob("./cmd/templates/*")
	err = di.DependencyInjection(router, cfg)
	if err != nil {
		//log.Fatalf("Cannot Start server due to failure in DependencyInjectin: %v", err)
		log.Fatal("Cannot Start server due to failure in DependencyInjectin:",
			logger.Field{Key: "error", Value: err})
	}
	// err = router.Run(cfg.Port)
	// if err != nil {
	// 	//log.Fatalf("Error starting server: %v\n", err)
	// 	log.Fatal("Error starting server:",
	// 		logger.Field{Key: "error", Value: err})
	// }
	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: router,
	}

	log.Info("Starting API Gateway",
		logger.Field{Key: "port", Value: cfg.Port})

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

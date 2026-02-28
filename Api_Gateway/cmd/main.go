package main

import (
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

	//fmt.Println("Config.Port", config.Port, "Config.AuthSubscriptionSvcUrl", config.AuthSubscriptionSvcUrl)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware(log))

	config, err := config.LoadConfig()
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
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	router.LoadHTMLGlob("./cmd/templates/*")
	err = di.DependencyInjection(router, config)
	if err != nil {
		//log.Fatalf("Cannot Start server due to failure in DependencyInjectin: %v", err)
		log.Fatal("Cannot Start server due to failure in DependencyInjectin:",
			logger.Field{Key: "error", Value: err})
	}
	err = router.Run(config.Port)
	if err != nil {
		//log.Fatalf("Error starting server: %v\n", err)
		log.Fatal("Error starting server:",
			logger.Field{Key: "error", Value: err})
	}
}

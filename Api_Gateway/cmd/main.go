package main

import (
	"fmt"
	"log"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/di"
	"github.com/gin-gonic/gin"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("cannot load configuration: %v", err)
	}
	fmt.Println("Config.Port", config.Port, "Config.AuthSubscriptionSvcUrl", config.AuthSubscriptionSvcUrl)
	router := gin.New()
	err = di.DependencyInjection(router, config)
	if err != nil {
		log.Fatalf("Cannot Start server due to failure in DependencyInjectin: %v", err)
	}
	err = router.Run(config.Port)
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}

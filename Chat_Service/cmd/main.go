package main

import (
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

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load configuration:",
			logger.Field{Key: "error", Value: err})
	}

	err = di.DependencyInjection(router, config,log)
	if err != nil {
		log.Fatal("Cannot Start server due to failure in DependencyInjectin:",
			logger.Field{Key: "error", Value: err})
	}
	handler.StartHub()
	err = router.Run(config.PortMngr.RunnerPort)
	if err != nil {
		log.Fatal("Error starting server:",
			logger.Field{Key: "error", Value: err})
	}

}

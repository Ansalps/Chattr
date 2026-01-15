package main

import (
	"log"

	"github.com/Ansalps/Chattr_Chat_Service/pkg/config"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/di"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config", err)
	}
	router := gin.New()
	 err = di.DependencyInjection(router,config)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
	handler.StartHub()
	err = router.Run(config.PortMngr.RunnerPort)
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}

}

package main

import (
	"log"

	"github.com/Ansalps/Chattr_Post_Relation_service/pkg/config"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config", err)
	}
	PostRelationServiceServer, err := di.DependencyIndjection(config)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
	lis, err := net.Listen("tcp", config.PortMngr.RunnerPort)
	if err != nil {
		log.Fatal(err)
	}
}

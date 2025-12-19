package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/di"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/pb"
	"google.golang.org/grpc"
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
	fmt.Println("Post_Relation_Service started on:", config.PortMngr.RunnerPort)
	grpcServer := grpc.NewServer()
	pb.RegisterPostRelationServiceServer(grpcServer, PostRelationServiceServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}

}

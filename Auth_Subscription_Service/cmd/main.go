package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/di"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/pb"
	"google.golang.org/grpc"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config", err)
	}
	AuthSubscriptionServiceServer, err := di.DependencyIndjection(config)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
	lis, err := net.Listen("tcp", config.Port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Auth_Subscription_Service started on:", config.Port)
	grpcServer := grpc.NewServer()
	pb.RegisterAuthSubscriptionServiceServer(grpcServer, AuthSubscriptionServiceServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}

}

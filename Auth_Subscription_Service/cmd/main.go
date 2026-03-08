package main

import (
	"net"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/logger"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/middleware"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/di"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/pb"
	"google.golang.org/grpc"
	"github.com/grpc-ecosystem/go-grpc-middleware"
)

func main() {

	// 1️⃣ Initialize logger
	log, err := logger.NewZapLogger()
	if err != nil {
		panic(err)
	}

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config",
			logger.Field{Key: "error", Value: err},
		)
	}
	AuthSubscriptionServiceServer, err := di.DependencyIndjection(config)
	if err != nil {
		log.Fatal("cannot start server",
			logger.Field{Key: "error", Value: err},
		)
	}
	lis, err := net.Listen("tcp", config.PortMngr.RunnerPort)
	if err != nil {
		log.Fatal("failed to listen",
			logger.Field{Key: "error", Value: err},
		)
	}
	log.Info("Auth Subscription Service started",
		logger.Field{Key: "port", Value: config.PortMngr.RunnerPort},
	)
	// 5️⃣ Create gRPC server with logging interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				middleware.RecoveryInterceptor,
				middleware.LoggerInterceptor(log),
			),
		),
	)
	pb.RegisterAuthSubscriptionServiceServer(grpcServer, AuthSubscriptionServiceServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to start grpc server",
			logger.Field{Key: "error", Value: err},
		)
	}

}

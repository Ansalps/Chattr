package main

import (
	"net"

	"github.com/Ansalps/Chattr_Post_Relation_Service/infrastructure/logger"
	"github.com/Ansalps/Chattr_Post_Relation_Service/middleware"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/di"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/pb"
	"google.golang.org/grpc"
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
	PostRelationServiceServer, err := di.DependencyIndjection(config)
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
	log.Info("Post Relation Service started",
		logger.Field{Key: "port", Value: config.PortMngr.RunnerPort},
	)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.LoggerInterceptor(log)),
	)
	pb.RegisterPostRelationServiceServer(grpcServer, PostRelationServiceServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to start grpc server",
			logger.Field{Key: "error", Value: err},
		)
	}

}

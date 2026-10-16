package main

import (
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/logger"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/middleware"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/di"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/pb"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

func main() {

	// 1️⃣ Initialize logger
	log, err := logger.NewZapLogger()
	if err != nil {
		panic(err)
	}

	// ✅ start pprof server
	go func() {
		log.Info("pprof running on 127.0.0.1:6061")
		if err := http.ListenAndServe("127.0.0.1:6061", nil); err != nil {
			log.Error("pprof server failed",
				logger.Field{Key: "error", Value: err})
		}
	}()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config",
			logger.Field{Key: "error", Value: err},
		)
	}

	AuthSubscriptionServiceServer, err := di.DependencyIndjection(cfg)
	if err != nil {
		log.Fatal("cannot start server",
			logger.Field{Key: "error", Value: err},
		)
	}
	lis, err := net.Listen("tcp", cfg.PortMngr.RunnerPort)
	if err != nil {
		log.Fatal("failed to listen",
			logger.Field{Key: "error", Value: err},
		)
	}
	// log.Info("Auth Subscription Service started",
	// 	logger.Field{Key: "port", Value: cfg.PortMngr.RunnerPort},
	// )
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
	// if err := grpcServer.Serve(lis); err != nil {
	// 	log.Fatal("failed to start grpc server",
	// 		logger.Field{Key: "error", Value: err},
	// 	)
	// }
	log.Info("Auth Subscription Service started",
		logger.Field{Key: "port", Value: cfg.PortMngr.RunnerPort},
	)

	// start server in goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to start grpc server",
				logger.Field{Key: "error", Value: err})
		}
	}()

	// wait for signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down gRPC server...")

	grpcServer.GracefulStop()

	log.Info("gRPC server stopped")
}

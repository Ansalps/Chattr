package client

import (
	"log"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitPostRelationServiceClient(cfg *config.Config) (pb.PostRelationServiceClient, error) {
	grpcConnection, err := grpc.NewClient(cfg.PortMngr.PostRelationSvcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	grpcClient := pb.NewPostRelationServiceClient(grpcConnection)
	return grpcClient, nil
}

package client

import (
	"fmt"

	"github.com/Ansalps/Chattr_Notification_Service/pkg/config"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitAuthSubscriptionServiceClient(cfg *config.Config) (pb.AuthSubscriptionServiceClient, error) {
	grpcConnection, err := grpc.NewClient(cfg.PortMngr.AuthSvcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		//log.Fatalf("could not connect: %v", err)
		return nil, fmt.Errorf("%w: %v", err, "error connection to grpc service auth subscriptioin service")
	}
	grpcClient := pb.NewAuthSubscriptionServiceClient(grpcConnection)
	return grpcClient, nil
}

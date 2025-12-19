package client

import (
	"log"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitAuthSubscriptionServiceClient(cfg *config.Config)(pb.AuthSubscriptionServiceClient,error){
	grpcConnection,err:=grpc.NewClient(cfg.PortMngr.AuthSvcUrl,grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err!=nil{
		log.Fatalf("could not connect: %v",err)
	}
	grpcClient:=pb.NewAuthSubscriptionServiceClient(grpcConnection)
	return grpcClient,nil
}
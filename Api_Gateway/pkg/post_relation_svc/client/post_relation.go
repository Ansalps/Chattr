package client

import (
	"context"
	"log"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/pb/post_relation"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/client/interfaces"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/responsemodels"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PostRelationClient struct {
	Client post_relation.PostRelationServiceClient
}

func NewPostRelationClient(cfg *config.Config) interfaces.PostRelationClient {
	grpcConnection, err := grpc.NewClient(cfg.PostRelationSvcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	grpcClient:=post_relation.NewPostRelationServiceClient(grpcConnection)
	return &PostRelationClient{
		Client: grpcClient,
	}
}

func (as *PostRelationClient)CreatePost(createPostReq requestmodels.CreatePostRequest)(responsemodels.CreatePostResponse,error){
resp,err:=as.Client.CreatePost(context.Background(),&post_relation.CreatePostRequest{
	UserId: createPostReq.UserID,
	Caption: createPostReq.Caption,
	MediaUrls: createPostReq.MediaUrls,
})
if err != nil {
	log.Printf("grpc create post call failed :%v", err)
	return responsemodels.CreatePostResponse{}, err
}
	return responsemodels.CreatePostResponse{
		PostID: resp.PostId,
	},nil
}
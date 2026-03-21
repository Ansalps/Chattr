package client

import (
	"context"
	"fmt"
	"log"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func InitAuthSubscriptionServiceClient(cfg *config.Config) (pb.AuthSubscriptionServiceClient, error) {
	grpcConnection, err := grpc.NewClient(cfg.PortMngr.AuthSvcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	grpcClient := pb.NewAuthSubscriptionServiceClient(grpcConnection)
	return grpcClient, nil
}

func FetchUserMetaData(
	client pb.AuthSubscriptionServiceClient,
	userIDs []uint64,
) (*pb.BatchUserMetadataResponse, error) {

	userResp, err := client.FetchUserMetaData(
		context.Background(),
		&pb.UserDataReq{
			UserId: userIDs,
		},
	)

	if err != nil {

		st, ok := status.FromError(err)
		if !ok {
			return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
		}

		switch st.Code() {

		case codes.NotFound:
			return nil, domain.ErrUsersNotFound

		case codes.Internal:
			return nil, fmt.Errorf("%w: %v", domain.ErrDatabase, err)

		default:
			return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
		}
	}

	// var usermetada []responsemodels.UserMetaData

	// for _, v := range userResp.Users {
	// 	usermetada = append(usermetada, responsemodels.UserMetaData{
	// 		UserID:        v.UserId,
	// 		UserName:      v.UserName,
	// 		Name:          v.Name,
	// 		ProfileImgUrl: v.ProfileImgUrl,
	// 		BlueTick:      v.BlueTick,
	// 	})
	// }

	return userResp, nil
}

func CheckUserExists(
	client pb.AuthSubscriptionServiceClient,
	userID uint64,
) error {

	resp, err := client.CheckUserExists(
		context.Background(),
		&pb.CheckUserExistsRequest{
			UserId: userID,
		},
	)

	if err != nil {

		st, ok := status.FromError(err)
		if !ok {
			return fmt.Errorf("%w: %v", domain.ErrInternal, err)
		}

		switch st.Code() {

		case codes.NotFound:
			return domain.ErrUserNotFound

		case codes.Internal:
			return fmt.Errorf("%w: %v", domain.ErrDatabase, err)

		default:
			return fmt.Errorf("%w: %v", domain.ErrInternal, err)
		}
	}

	if !resp.Exists {
		return domain.ErrUserNotFound
	}

	return nil
}

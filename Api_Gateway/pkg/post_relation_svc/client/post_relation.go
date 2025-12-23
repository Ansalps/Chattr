package client

import (
	"context"
	"fmt"
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

func NewPostRelationClient(cfg *config.Config) interfaces.PostRelationClientInterface {
	grpcConnection, err := grpc.NewClient(cfg.PostRelationSvcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	grpcClient := post_relation.NewPostRelationServiceClient(grpcConnection)
	return &PostRelationClient{
		Client: grpcClient,
	}
}

func (as *PostRelationClient) CreatePost(createPostReq requestmodels.CreatePostRequest) (responsemodels.CreatePostResponse, error) {
	resp, err := as.Client.CreatePost(context.Background(), &post_relation.CreatePostRequest{
		UserId:    createPostReq.UserID,
		Caption:   createPostReq.Caption,
		MediaUrls: createPostReq.MediaUrls,
	})
	if err != nil {
		log.Printf("grpc create post call failed :%v", err)
		return responsemodels.CreatePostResponse{}, err
	}
	return responsemodels.CreatePostResponse{
		PostID: resp.PostId,
	}, nil
}

func (as *PostRelationClient) EditPost(editPostReq requestmodels.EditPostRequest) (responsemodels.EditPostResponse, error) {
	resp, err := as.Client.EditPost(context.Background(), &post_relation.EditPostRequest{
		UserId:  editPostReq.UserID,
		PostId:  editPostReq.PostID,
		Caption: editPostReq.Caption,
	})
	if err != nil {
		log.Printf("grpc edit post call failed:%v", err)
		return responsemodels.EditPostResponse{}, err
	}
	return responsemodels.EditPostResponse{
		Caption: resp.Caption,
	}, nil
}

func (as *PostRelationClient) DeletePost(deletePostReq requestmodels.DeletePostRequest) (responsemodels.DeletePostResponse, error) {
	resp, err := as.Client.DeletePost(context.Background(), &post_relation.DeletePostRequest{
		UserId: deletePostReq.UserID,
		PostId: deletePostReq.PostID,
	})
	if err != nil {
		log.Printf("grpc call failed for delete post, error: %v", err)
		return responsemodels.DeletePostResponse{}, err
	}
	return responsemodels.DeletePostResponse{
		PostID: resp.PostId,
	}, nil
}

func (as *PostRelationClient) LikePost(likePostReq requestmodels.LikePostRequest) (responsemodels.LikePostResponse, error) {
	resp, err := as.Client.LikePost(context.Background(), &post_relation.LikePostRequest{
		UserId: likePostReq.UserID,
		PostId: likePostReq.PostID,
	})
	if err != nil {
		log.Println("grpc call failed for like post, error: ", err)
		return responsemodels.LikePostResponse{}, err
	}
	return responsemodels.LikePostResponse{
		PostID: resp.PostId,
	}, nil
}

func (as *PostRelationClient) UnlikePost(unlikePostReq requestmodels.UnlikePostRequest) (responsemodels.UnlikePostResponse, error) {
	resp, err := as.Client.UnlikePost(context.Background(), &post_relation.UnlikePostRequest{
		UserId: unlikePostReq.UserID,
		PostId: unlikePostReq.PostID,
	})
	if err != nil {
		log.Println("grpc call failed for unlike post", err)
		return responsemodels.UnlikePostResponse{}, err
	}
	return responsemodels.UnlikePostResponse{
		PostID: resp.PostId,
	}, nil
}
func (as *PostRelationClient) AddComment(addCommentReq requestmodels.AddCommentRequest) (responsemodels.AddCommentResponse, error) {
	var resp *post_relation.AddCommentResponse
	resp, err := as.Client.AddComment(context.Background(), &post_relation.AddCommentRequest{
		UserId:          addCommentReq.UserID,
		PostId:          addCommentReq.PostID,
		CommentText:     addCommentReq.CommentText,
		ParentCommentId: addCommentReq.ParentCommentId,
	})
	if err != nil {
		log.Println("grpc call failed for add comment", err)
		return responsemodels.AddCommentResponse{}, err
	}
	fmt.Println("resp in client in api gateway", resp)
	return responsemodels.AddCommentResponse{
		UserID:          resp.UserId,
		PostID:          resp.PostId,
		CommentID:       resp.CommentId,
		CommentText:     resp.CommentText,
		ParentCommentId: resp.ParentCommentId,
	}, nil
}

func (as *PostRelationClient) EditComment(editCommentReq requestmodels.EditCommentRequest) (responsemodels.EditCommentResponse, error) {
	//var resp *post_relation.EditCommentResponse
	resp, err := as.Client.EditComment(context.Background(), &post_relation.EditCommentRequest{
		UserId:      editCommentReq.UserID,
		PostId:      editCommentReq.PostID,
		CommentId:   editCommentReq.CommentID,
		CommentText: editCommentReq.CommentText,
	})
	if err != nil {
		log.Println("grpc edit Comment call failed", err)
		return responsemodels.EditCommentResponse{}, err
	}
	fmt.Println("check in ",resp)
	return responsemodels.EditCommentResponse{
		PostID: resp.PostId,
		CommentID:   resp.CommentId,
		CommentText: resp.CommentText,
	}, nil
}

func (as *PostRelationClient) DeleteComment(deleteCommentReq requestmodels.DeleteCommentRequest) (responsemodels.DeleteCommentResponse, error) {
	resp, err := as.Client.DeleteComment(context.Background(), &post_relation.DeleteCommentRequest{
		UserId:    deleteCommentReq.UserID,
		PostId:    deleteCommentReq.PostID,
		CommentId: deleteCommentReq.CommentID,
	})
	if err != nil {
		log.Println("grpc call failed for delete comment", err)
		return responsemodels.DeleteCommentResponse{}, err
	}
	return responsemodels.DeleteCommentResponse{
		CommentID: resp.CommentId,
	}, nil

}

func (as *PostRelationClient) Follow(followReq requestmodels.FollowRequest) (responsemodels.FollowResponse, error) {
	resp, err := as.Client.Follow(context.Background(), &post_relation.FollowRequest{
		UserId:          followReq.UserID,
		FollowingUserId: followReq.FollowingUserID,
	})
	if err != nil {
		log.Println("grpc call fialed for follow", err)
		return responsemodels.FollowResponse{}, err
	}
	return responsemodels.FollowResponse{
		FollowingUserID: resp.FollowingUserId,
	}, nil
}

func (as *PostRelationClient) Unfollow(unfollowReq requestmodels.UnfollowRequest) (responsemodels.UnfollowResponse, error) {
	resp, err := as.Client.Unfollow(context.Background(), &post_relation.UnfollowRequest{
		UserId:            unfollowReq.UserID,
		UnfollowingUserId: unfollowReq.UnfollowingUserID,
	})
	if err != nil {
		log.Println("error in grpc calling of unfollow, error: ", err)
		return responsemodels.UnfollowResponse{}, err
	}
	return responsemodels.UnfollowResponse{
		UnfollowingUserID: resp.UnfollowingUserId,
	}, nil
}

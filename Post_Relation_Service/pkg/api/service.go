package services

import (
	"context"
	"fmt"
	"log"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/pb"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase/interfacesUsecase"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostRelationServer struct {
	PostRelationUsecase interfacesUsecase.PostRelationUsecase
	pb.PostRelationServiceServer
}

func NewPostRelationSever(useCase interfacesUsecase.PostRelationUsecase) *PostRelationServer {
	return &PostRelationServer{
		PostRelationUsecase: useCase,
	}
}

func (as *PostRelationServer) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {
	fmt.Println("in service print print urls request",req.MediaUrls)
	createPostReq := requestmodels.CreatePostRequest{
		UserID:    req.UserId,
		Caption:   req.Caption,
		MediaUrls: req.MediaUrls,
	}
	createPostRes, err := as.PostRelationUsecase.CreatePost(createPostReq)
	if err != nil {
		return &pb.CreatePostResponse{}, nil
	}
	return &pb.CreatePostResponse{
		PostId: createPostRes.PostID,
	}, nil
}

func (as *PostRelationServer) EditPost(ctx context.Context, req *pb.EditPostRequest) (*pb.EditPostResponse, error) {
	editPostReq := requestmodels.EditPostRequest{
		UserID:  req.UserId,
		PostID:  req.PostId,
		Caption: req.Caption,
	}
	editPostRes, err := as.PostRelationUsecase.EditPost(editPostReq)
	if err != nil {
		if err == usecase.ErrPostNotFound {
			return nil, status.Error(codes.NotFound, "post not found")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &pb.EditPostResponse{
		Caption: editPostRes.Caption,
	}, nil
}
func (as *PostRelationServer) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	deletePostReq := requestmodels.DeletePostRequest{
		UserID: req.UserId,
		PostID: req.PostId,
	}
	deletPostRes, err := as.PostRelationUsecase.DeletePost(deletePostReq)
	if err != nil {
		if err == usecase.ErrPostNotFound {
			return nil, status.Error(codes.NotFound, "post not found")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &pb.DeletePostResponse{
		PostId: deletPostRes.PostID,
	}, nil
}

func (as *PostRelationServer) LikePost(ctx context.Context, req *pb.LikePostRequest) (*pb.LikePostResponse, error) {
	likePostReq := requestmodels.LikePostRequest{
		UserID: req.UserId,
		PostID: req.PostId,
	}
	likePostRes, err := as.PostRelationUsecase.LikePost(likePostReq)
	if err != nil {

		log.Println("internal error", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &pb.LikePostResponse{
		PostId: likePostRes.PostID,
	}, nil
}

func (as *PostRelationServer) UnlikePost(ctx context.Context, req *pb.UnlikePostRequest) (*pb.UnlikePostResponse, error) {
	unlikePostReq := requestmodels.UnlikePostRequest{
		UserID: req.UserId,
		PostID: req.PostId,
	}
	unlikePostResponse, err := as.PostRelationUsecase.UnlikePost(unlikePostReq)
	if err != nil {
		if err == usecase.ErrPostLikeNotFound {
			return nil, status.Error(codes.NotFound, "post like not found")
		}
		log.Println("error in service", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &pb.UnlikePostResponse{
		PostId: unlikePostResponse.PostID,
	}, nil
}

func (as *PostRelationServer) AddComment(ctx context.Context, req *pb.AddCommentRequest) (*pb.AddCommentResponse, error) {
	addCommentReq := requestmodels.AddCommentRequest{
		UserID:          req.UserId,
		PostID:          req.PostId,
		CommentText:     req.CommentText,
		ParentCommentId: req.ParentCommentId,
	}
	fmt.Println("befor going to usecase")
	addCommentRes, err := as.PostRelationUsecase.AddComment(addCommentReq)
	if err != nil {
		switch err {
		case usecase.ErrRecursiveComment:
			return nil, status.Error(codes.FailedPrecondition, "can't reply to a comment reply")
		case domain.ErrForeignKeyViolationCommentPost:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
		// if err == usecase.ErrRecursiveComment {
		// 	return nil, status.Error(codes.FailedPrecondition, "can't reply to a comment reply")
		// } else if err==domain.ErrForeignKeyViolationCommentPost{
		// 	return nil,status.Error(codes.NotFound,err.Error())
		// }
	}
	fmt.Println("is it reaching here", addCommentRes)

	return &pb.AddCommentResponse{
		UserId:          addCommentRes.UserID,
		PostId:          addCommentRes.PostID,
		CommentText:     addCommentRes.CommentText,
		CommentId:       addCommentRes.CommentID,
		ParentCommentId: addCommentRes.ParentCommentId,
	}, nil
}

func (as *PostRelationServer) EditComment(ctx context.Context, req *pb.EditCommentRequest) (*pb.EditCommentResponse, error) {
	editCommentReq := requestmodels.EditCommentRequest{
		UserID:      req.UserId,
		PostID:      req.PostId,
		CommentID:   req.CommentId,
		CommentText: req.CommentText,
	}
	editCommentRes, err := as.PostRelationUsecase.EditComment(editCommentReq)
	if err != nil {
		log.Println("error in servic :", err)
		if err == usecase.ErrCommentNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}
	fmt.Println("check in api for response", editCommentRes)
	return &pb.EditCommentResponse{
		PostId:      editCommentReq.PostID,
		CommentId:   editCommentRes.CommentID,
		CommentText: editCommentRes.CommentText,
	}, nil
}
func (as *PostRelationServer) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	fmt.Println("what is actually the issue?")
	deleteCommentReq := requestmodels.DeleteCommentRequest{
		UserID:    req.UserId,
		PostID:    req.PostId,
		CommentID: req.CommentId,
	}
	deleteCommentRes, err := as.PostRelationUsecase.DeleteComment(deleteCommentReq)
	if err != nil {
		if err == usecase.ErrCommentNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}
	return &pb.DeleteCommentResponse{
		CommentId: deleteCommentRes.CommentID,
	}, nil
}

func (as *PostRelationServer) Follow(ctx context.Context, req *pb.FollowRequest) (*pb.FollowResponse, error) {
	followReq := requestmodels.FollowRequest{
		UserID:          req.UserId,
		FollowingUserID: req.FollowingUserId,
	}
	followResponse, err := as.PostRelationUsecase.Follow(followReq)
	if err != nil {
		log.Println("error in service", err)
	}
	return &pb.FollowResponse{
		FollowingUserId: followResponse.FollowingUserID,
	}, nil
}
func (as *PostRelationServer) Unfollow(ctx context.Context, req *pb.UnfollowRequest) (*pb.UnfollowResponse, error) {
	unfollowReq := requestmodels.UnfollowRequest{
		UserID:            req.UserId,
		UnfollowingUserID: req.UnfollowingUserId,
	}
	unfollowResponse, err := as.PostRelationUsecase.Unfollow(unfollowReq)
	if err != nil {
		log.Println("error in service")
	}
	return &pb.UnfollowResponse{
		UnfollowingUserId: unfollowResponse.UnfollowingUserID,
	}, nil
}
func (as *PostRelationServer) FetchComments(ctx context.Context, req *pb.FetchCommentsRequest) (*pb.FetchCommentsResponse, error) {
	fetchCommentsReq := requestmodels.FetchCommentsReqeust{
		PostID: req.PostId,
	}
	fetchCommentsResponse, err := as.PostRelationUsecase.FetchComments(fetchCommentsReq)
	if err != nil {
		if err == usecase.ErrNoComments {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		log.Println("error in service", err)
		return nil, err
	}
	comments := make([]*pb.Comment, len(fetchCommentsResponse.Comments))

	for i, v := range fetchCommentsResponse.Comments {
		var childComments []*pb.Comment
		if len(v.ChildComment) > 0 {
			childComments = make([]*pb.Comment, len(v.ChildComment))
			for i, v := range v.ChildComment {
				childComments[i] = &pb.Comment{
					Id:          v.CommentID,
					CommentText: v.CommentText,
					CreatedAt:   utils.ToProtoTimestamp(v.CreatedAt),
					UserDetails: &pb.UserMetaData{
						UserId:        v.UserDetails.UserID,
						UserName:      v.UserDetails.UserName,
						Name:          v.UserDetails.Name,
						ProfileImgUrl: v.UserDetails.ProfileImgUrl,
						BlueTick:      v.UserDetails.BlueTick,
					},
					ParentCommentId: v.ParentCommentID,
				}
			}
		}
		comments[i] = &pb.Comment{
			Id:          v.CommentID,
			CommentText: v.CommentText,
			CreatedAt:   utils.ToProtoTimestamp(v.CreatedAt),
			CommentAge:  v.CommentAge,
			UserDetails: &pb.UserMetaData{
				UserId:        v.UserDetails.UserID,
				UserName:      v.UserDetails.UserName,
				Name:          v.UserDetails.Name,
				ProfileImgUrl: v.UserDetails.ProfileImgUrl,
				BlueTick:      v.UserDetails.BlueTick,
			},
			ParentCommentId:   v.ParentCommentID,
			ChildCommentCount: uint64(len(v.ChildComment)),
			ChildComment:      childComments,
		}
	}
	return &pb.FetchCommentsResponse{
		Comments: comments,
	}, nil
}

func (as *PostRelationServer) PostFollowCount(ctx context.Context, req *pb.PostFollowCountRequest) (*pb.PostFollowCountResponse, error) {
	resp, err := as.PostRelationUsecase.PostFollowCount(req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.PostFollowCountResponse{
		PostCount:      resp.PostCount,
		FollowerCount:  resp.FollowerCount,
		FollowingCount: resp.FollowingCount,
	}, nil
}
func (as *PostRelationServer) FetchAllPosts(ctx context.Context, req *pb.FetchAllPostsRequest) (*pb.FetchAllPostsResponse, error) {
	_, err := as.PostRelationUsecase.FetchAllPosts(req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.FetchAllPostsResponse{}, nil
}

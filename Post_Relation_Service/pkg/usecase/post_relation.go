package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/pb"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/respository/interfacesRepository"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase/interfacesUsecase"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/utils"
	"gorm.io/gorm"
)

type PostRelationUsecase struct {
	PostRelationRepository interfacesRepository.PostRelationRepository
	AuthSubscriptionClient pb.AuthSubscriptionServiceClient
}

var (
	ErrPostNotFound     = errors.New("Post Not found or user does not have permission")
	ErrPostLikeNotFound = errors.New("Post Like Not found")
	ErrRecursiveComment = errors.New("can't reply to a comment reply")
	ErrCommentNotFound  = errors.New("comment doesn't exist or post doesn't exist or user does not have permission")
	ErrFollowOwn        = errors.New("can't follow yourself")
	ErrUsertNotFound    = errors.New("User not found")
	ErrUnfollowOwn      = errors.New("can't unfollow yourself")
	ErrNoComments       = errors.New("No comments to Fetch for the Post or Post doesn't exist")
	ErrNoPosts          = errors.New("No Posts to Fetch")
)

func NewPostRelationUsecase(repository interfacesRepository.PostRelationRepository, authSubClient pb.AuthSubscriptionServiceClient) interfacesUsecase.PostRelationUsecase {
	return &PostRelationUsecase{
		PostRelationRepository: repository,
		AuthSubscriptionClient: authSubClient,
	}
}

func (as *PostRelationUsecase) CreatePost(createPostReq requestmodels.CreatePostRequest) (responsemodels.CreatePostResponse, error) {
	createPostRes, err := as.PostRelationRepository.CreatePost(createPostReq)
	if err != nil {
		return responsemodels.CreatePostResponse{}, nil
	}
	return responsemodels.CreatePostResponse{
		PostID: createPostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) EditPost(editPostReq requestmodels.EditPostRequest) (responsemodels.EditPostResponse, error) {
	editPostRes, err := as.PostRelationRepository.EditPostById(editPostReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.EditPostResponse{}, ErrPostNotFound
		}
		return responsemodels.EditPostResponse{}, err
	}
	return responsemodels.EditPostResponse{
		Caption: editPostRes.Caption,
	}, nil
}

func (as *PostRelationUsecase) DeletePost(deletePostReq requestmodels.DeletePostRequest) (responsemodels.DeletePostResponse, error) {
	deletePostRes, err := as.PostRelationRepository.DeletePostById(deletePostReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.DeletePostResponse{}, ErrPostNotFound
		}
		return responsemodels.DeletePostResponse{}, err
	}
	return responsemodels.DeletePostResponse{
		PostID: deletePostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) LikePost(likePostReq requestmodels.LikePostRequest) (responsemodels.LikePostResponse, error) {
	likePostRes, err := as.PostRelationRepository.LikePostById(likePostReq)
	if err != nil {
		return responsemodels.LikePostResponse{}, err
	}
	return responsemodels.LikePostResponse{
		PostID: likePostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) UnlikePost(unlikePostReq requestmodels.UnlikePostRequest) (responsemodels.UnlikePostResponse, error) {
	unlikePostRes, err := as.PostRelationRepository.UnlikePostById(unlikePostReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.UnlikePostResponse{}, ErrPostLikeNotFound
		}
		return responsemodels.UnlikePostResponse{}, err
	}
	return responsemodels.UnlikePostResponse{
		PostID: unlikePostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) AddComment(addCommentReq requestmodels.AddCommentRequest) (responsemodels.AddCommentResponse, error) {
	if addCommentReq.ParentCommentId != nil {
		fmt.Println("is reaching in here in add comment where parent comment Id not nil")
		isReplytoReply, err := as.PostRelationRepository.CheckCommentHieracrchy(addCommentReq.ParentCommentId)
		if err != nil {
			return responsemodels.AddCommentResponse{}, err
		}
		fmt.Println("print the truth :", isReplytoReply)
		if isReplytoReply {
			fmt.Println("it is true")
			return responsemodels.AddCommentResponse{}, ErrRecursiveComment
			//fmt.Println("here 1")
		}
		fmt.Println("here 2")
	}
	fmt.Println("here 3")
	addCommentRes, err := as.PostRelationRepository.AddComment(addCommentReq)
	if err != nil {
		return responsemodels.AddCommentResponse{}, err
	}
	return addCommentRes, nil
}
func (as *PostRelationUsecase) EditComment(editCommentReq requestmodels.EditCommentRequest) (responsemodels.EditCommentResponse, error) {
	resp, err := as.PostRelationRepository.EditComment(editCommentReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.EditCommentResponse{}, ErrCommentNotFound
		}
		return responsemodels.EditCommentResponse{}, err
	}
	return resp, nil
}
func (as *PostRelationUsecase) DeleteComment(deleteCommentReq requestmodels.DeleteCommentRequest) (responsemodels.DeleteCommentResponse, error) {
	deleteCommentRes, err := as.PostRelationRepository.DeleteCommentById(deleteCommentReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.DeleteCommentResponse{}, ErrCommentNotFound
		}
		return responsemodels.DeleteCommentResponse{}, err
	}
	return responsemodels.DeleteCommentResponse{
		CommentID: deleteCommentRes.CommentID,
	}, nil
}
func (as *PostRelationUsecase) Follow(followReq requestmodels.FollowRequest) (responsemodels.FollowResponse, error) {
	if followReq.UserID == followReq.FollowingUserID {
		return responsemodels.FollowResponse{}, ErrFollowOwn
	}
	_, err := as.AuthSubscriptionClient.CheckUserExists(context.Background(), &pb.CheckUserExistsRequest{
		UserId: followReq.FollowingUserID,
	})
	if err != nil {
		log.Println("inter service call for check user exist failed, error: ", err)
		return responsemodels.FollowResponse{}, err
	}
	followRes, err := as.PostRelationRepository.Follow(followReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.FollowResponse{}, ErrUsertNotFound
		}
		return responsemodels.FollowResponse{}, err
	}
	return responsemodels.FollowResponse{
		FollowingUserID: followRes.FollowingUserID,
	}, nil
}

func (as *PostRelationUsecase) Unfollow(unfollowReq requestmodels.UnfollowRequest) (responsemodels.UnfollowResponse, error) {
	if unfollowReq.UserID == unfollowReq.UnfollowingUserID {
		return responsemodels.UnfollowResponse{}, ErrUnfollowOwn
	}
	_, err := as.AuthSubscriptionClient.CheckUserExists(context.Background(), &pb.CheckUserExistsRequest{
		UserId: unfollowReq.UnfollowingUserID,
	})
	if err != nil {
		log.Println("inter service call for check user exist failed, error: ", err)
		return responsemodels.UnfollowResponse{}, err
	}
	unfollowRes, err := as.PostRelationRepository.UnfollowUserById(unfollowReq)
	if err != nil {
		return responsemodels.UnfollowResponse{}, err
	}
	return responsemodels.UnfollowResponse{
		UnfollowingUserID: unfollowRes.UnfollowingUserID,
	}, nil
}

func (as *PostRelationUsecase) FetchComments(fetchCommentsReq requestmodels.FetchCommentsReqeust) (responsemodels.FetchCommentsResponse, error) {
	commentsRes, err := as.PostRelationRepository.FetchCommentsByPostId(fetchCommentsReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.FetchCommentsResponse{}, ErrNoComments
		}
		return responsemodels.FetchCommentsResponse{}, err
	}
	userIDs := make(map[uint64]bool)
	for _, v := range commentsRes {
		userIDs[v.UserID] = true
	}
	userids := make([]uint64, len(userIDs))
	i := 0
	for k, _ := range userIDs {
		userids[i] = k
		i++
	}
	userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
		UserId: userids,
	})
	//v:=userResp[userIDs]
	if err != nil {
		log.Println("error calling service auth_subcription", err)
		return responsemodels.FetchCommentsResponse{}, err
	}
	var comments []responsemodels.Comment
	for i, v := range commentsRes {
		if commentsRes[i].ParentCommentID == nil {
			comments = append(comments, responsemodels.Comment{
				CommentID:   v.ID,
				CommentText: v.CommentText,
				CreatedAt:   v.CreatedAt,
				CommentAge:  utils.CalcuateCommentAge(v.CreatedAt),
				UserDetails: responsemodels.UserMetaData{
					UserID:        userResp.Users[v.UserID].UserId,
					UserName:      userResp.Users[v.UserID].UserName,
					Name:          userResp.Users[v.UserID].Name,
					ProfileImgUrl: userResp.Users[v.UserID].ProfileImgUrl,
					BlueTick:      userResp.Users[v.UserID].BlueTick,
				},
				ParentCommentID: v.ParentCommentID,
			})
		}
	}
	// create index lookup for parents
	parentIndex := make(map[uint64]int)
	for i, c := range comments {
		parentIndex[c.CommentID] = i
	}
	for i, v := range commentsRes {
		if commentsRes[i].ParentCommentID != nil {
			parentIdx, ok := parentIndex[*v.ParentCommentID]
			if !ok {
				return responsemodels.FetchCommentsResponse{}, errors.New("invalid parent comment id")
			}

			comments[parentIdx].ChildComment = append(comments[parentIdx].ChildComment, responsemodels.Comment{
				CommentID:   v.ID,
				CommentText: v.CommentText,
				CreatedAt:   v.CreatedAt,
				CommentAge:  utils.CalcuateCommentAge(v.CreatedAt),
				UserDetails: responsemodels.UserMetaData{
					UserID:        userResp.Users[v.UserID].UserId,
					UserName:      userResp.Users[v.UserID].UserName,
					Name:          userResp.Users[v.UserID].Name,
					ProfileImgUrl: userResp.Users[v.UserID].ProfileImgUrl,
					BlueTick:      userResp.Users[v.UserID].BlueTick,
				},
				ParentCommentID: v.ParentCommentID,
			})
		}
	}
	return responsemodels.FetchCommentsResponse{
		Comments: comments,
	}, nil
}

func (as *PostRelationUsecase) PostFollowCount(userid uint64) (responsemodels.PostFollowCountResponse, error) {
	postCount, err := as.PostRelationRepository.FetchPostCountByUserId(userid)
	if err != nil {
		return responsemodels.PostFollowCountResponse{}, err
	}
	fmt.Println("print post Count in usecase", postCount)
	resp, err := as.PostRelationRepository.FetchFollowCountByUserId(userid)
	if err != nil {
		return responsemodels.PostFollowCountResponse{}, err
	}
	fmt.Println("resp print first in usecase", resp)
	resp.PostCount = postCount
	fmt.Println("resp print second in usecase", resp, resp.PostCount)
	return resp, nil
}
func (as *PostRelationUsecase) FetchAllPosts(userid uint64) ([]responsemodels.PostWithCounts, error) {
	resp, err := as.PostRelationRepository.FetchAllPosts(userid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNoPosts
		}
		return nil, err
	}
	for i := range resp {
		resp[i].Age = utils.CalcuateCommentAge(resp[i].CreatedAt)
	}
	return resp, nil
}

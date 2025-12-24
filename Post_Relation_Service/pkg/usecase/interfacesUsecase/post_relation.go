package interfacesUsecase

import (
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/responsemodels"
)

type PostRelationUsecase interface {
	CreatePost(requestmodels.CreatePostRequest) (responsemodels.CreatePostResponse, error)
	FetchAllPosts(userid uint64) ([]responsemodels.PostWithCounts, error)
	EditPost(requestmodels.EditPostRequest) (responsemodels.EditPostResponse, error)
	DeletePost(requestmodels.DeletePostRequest) (responsemodels.DeletePostResponse, error)

	LikePost(requestmodels.LikePostRequest) (responsemodels.LikePostResponse, error)
	UnlikePost(requestmodels.UnlikePostRequest) (responsemodels.UnlikePostResponse, error)

	AddComment(requestmodels.AddCommentRequest) (responsemodels.AddCommentResponse, error)
	FetchComments(requestmodels.FetchCommentsReqeust) (responsemodels.FetchCommentsResponse, error)
	EditComment(requestmodels.EditCommentRequest) (responsemodels.EditCommentResponse, error)
	DeleteComment(requestmodels.DeleteCommentRequest) (responsemodels.DeleteCommentResponse, error)

	Follow(requestmodels.FollowRequest) (responsemodels.FollowResponse, error)
	Unfollow(requestmodels.UnfollowRequest) (responsemodels.UnfollowResponse, error)

	PostFollowCount(uint64) (responsemodels.PostFollowCountResponse, error)
}

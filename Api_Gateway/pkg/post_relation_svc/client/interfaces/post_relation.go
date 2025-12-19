package interfaces

import (
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/requestmodels"
)

type PostRelationClientInterface interface {
	CreatePost(requestmodels.CreatePostRequest) (responsemodels.CreatePostResponse,error)
	EditPost(requestmodels.EditPostRequest)(responsemodels.EditPostResponse,error)
	DeletePost(requestmodels.DeletePostRequest)(responsemodels.DeletePostResponse,error)
	LikePost(requestmodels.LikePostRequest)(responsemodels.LikePostResponse,error)
	UnlikePost(requestmodels.UnlikePostRequest)(responsemodels.UnlikePostResponse,error)

	AddComment(requestmodels.AddCommentRequest)(responsemodels.AddCommentResponse,error)
	FetchComments(requestmodels.FetchCommentsReqeust)(responsemodels.FetchCommentsResponse,error)
	EditComment(requestmodels.EditCommentRequest)(responsemodels.EditCommentResponse,error)
	DeleteComment(requestmodels.DeleteCommentRequest)(responsemodels.DeleteCommentResponse,error)

	FetchCommentsOfComment(requestmodels.FetchCommentsOfCommentReqeust)(responsemodels.FetchCommentsOfCommentResponse,error)

	Follow(requestmodels.FollowRequest)(responsemodels.FollowResponse,error)
	Unfollow(requestmodels.UnfollowRequest)(responsemodels.UnfollowResponse,error)
}

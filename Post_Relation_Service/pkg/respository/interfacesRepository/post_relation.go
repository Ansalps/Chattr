package interfacesRepository

import (
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/responsemodels"
)

type PostRelationRepository interface {
	CreatePost(requestmodels.CreatePostRequest) (responsemodels.CreatePostResponse, error)
	FetchAllPosts(currentuserid uint64,targetuserid uint64) ([]responsemodels.PostWithCounts, error)
	EditPostById(requestmodels.EditPostRequest) (responsemodels.EditPostResponse, error)
	DeletePostById(requestmodels.DeletePostRequest) (responsemodels.DeletePostResponse, error)

	LikePostById(requestmodels.LikePostRequest) (responsemodels.LikePostResponse, error)
	UnlikePostById(requestmodels.UnlikePostRequest) (responsemodels.UnlikePostResponse, error)
	CheckCommentHieracrchy(*uint64) (bool, error)
	AddComment(requestmodels.AddCommentRequest) (responsemodels.AddCommentResponse, error)
	EditComment(requestmodels.EditCommentRequest) (responsemodels.EditCommentResponse, error)
	DeleteCommentById(requestmodels.DeleteCommentRequest) (responsemodels.DeleteCommentResponse, error)

	Follow(requestmodels.FollowRequest) (responsemodels.FollowResponse, error)
	UnfollowUserById(requestmodels.UnfollowRequest) (responsemodels.UnfollowResponse, error)
	FetchFollowersUserIds(userid uint64)([]responsemodels.FollowerIds,error)
	FetchFollowingUserIds(userid uint64)([]responsemodels.FollowingIds,error)

	FetchCommentsByPostId(requestmodels.FetchCommentsReqeust) ([]responsemodels.Comments, error)

	FetchPostCountByUserId(uint64) (uint64, error)
	FetchFollowCountByUserId(uint64) (responsemodels.PostFollowCountResponse, error)
}

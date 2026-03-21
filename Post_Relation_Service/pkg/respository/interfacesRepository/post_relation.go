package interfacesRepository

import (
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/responsemodels"
)

type PostRelationRepository interface {
	CreatePost(requestmodels.CreatePostRequest) (responsemodels.CreatePostResponse, error)
	FetchAllPosts(requestmodels.FetchAllPostsReq) ([]responsemodels.PostWithCounts, error)
	EditPostById(requestmodels.EditPostRequest) (responsemodels.EditPostResponse, error)
	DeletePostById(requestmodels.DeletePostRequest) (responsemodels.DeletePostResponse, error)

	LikePostById(requestmodels.LikePostRequest) (responsemodels.LikePostResponse, error)
	FetchPostOwnerIdByPostId(uint64)(uint64,error)
	UnlikePostById(requestmodels.UnlikePostRequest) (responsemodels.UnlikePostResponse, error)
	CheckCommentHieracrchy(*uint64) (bool, error)
	AddComment(requestmodels.AddCommentRequest) (responsemodels.AddCommentResponse, error)
	EditComment(requestmodels.EditCommentRequest) (responsemodels.EditCommentResponse, error)
	DeleteCommentById(requestmodels.DeleteCommentRequest) (responsemodels.DeleteCommentResponse, error)

	Follow(requestmodels.FollowRequest) (responsemodels.FollowResponse, error)
	UnfollowUserById(requestmodels.UnfollowRequest) (responsemodels.UnfollowResponse, error)
	FetchFollowersUserIds(uint64)([]responsemodels.FollowerIds,error)
	FetchFollowersUserIds1(requestmodels.FetchFollowersRequest)([]responsemodels.FollowerIds,error)
	FetchFollowingUserIds(requestmodels.FetchFollowingRequest)([]responsemodels.FollowingIds,error)

	FetchCommentsByPostId(requestmodels.FetchCommentsReqeust) ([]responsemodels.Comments, error)

	FetchPostCountByUserId(uint64) (uint64, error)
	FetchFollowCountByUserId(uint64) (responsemodels.PostFollowCountResponse, error)

	//FetchPostDataForNewsFeed(requestmodels.FetchNewsFeedRequest)([]responsemodels.PostWithStatus,error)
	FetchNormalPostData(newsfeedReq requestmodels.FetchNewsFeedRequest) ([]responsemodels.PostWithStatus, error)
	GetFollowedCelebrityIDs(userID uint64) ([]uint64, error)
	FetchPostsByIDs(postIDs []uint64, viewerID uint64) ([]responsemodels.PostWithStatus, error)
	FetchCelebrityPostIDsFromSQL(celebIDs []uint64, lastID uint64, limit int) ([]uint64, error)
	FetchLatestPostIDsByUserID(userID uint64, limit int) ([]uint64, error)

	PromoteToCelebrity(userid uint64)error
	DepromoteToNormalUser(userid uint64)error

	FetchGlobalTrendingSQL(requestmodels.GlobalNewsFeedRequest)([]responsemodels.PostWithStatusWithTrendingScore,error)

	UpdatFollowCountOnFollow(uint64,uint64)(uint64,error)
	UpdatFollowCountOnUnFollow(uint64,uint64)(uint64,error)

	InsertUserIntoFollowCount(userid uint64)(error)
}

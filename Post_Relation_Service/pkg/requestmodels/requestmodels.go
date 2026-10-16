package requestmodels

type CreatePostRequest struct {
	UserID    uint64
	Caption   string
	MediaUrls []string
}
type EditPostRequest struct{
	UserID uint64
	PostID uint64
	Caption string
}
type DeletePostRequest struct{
	UserID uint64
	PostID uint64
}
type LikePostRequest struct{
	UserID uint64
	PostID uint64
}
type UnlikePostRequest struct{
	UserID uint64
	PostID uint64
}

type AddCommentRequest struct{
	UserID uint64
	PostID uint64
	CommentText string `json:"comment_text" validate:"required,min=1"`
	ParentCommentId *uint64 `json:"parent_comment_id" validate:"omitempty"`
}
type EditCommentRequest struct{
	UserID uint64
	PostID uint64
	CommentID uint64
	CommentText string `json:"comment_text" validate:"required,min=1"`
}
type DeleteCommentRequest struct{
	UserID uint64
	PostID uint64
	CommentID uint64
}
type FollowRequest struct{
	UserID uint64
	FollowingUserID uint64
}
type UnfollowRequest struct{
	UserID uint64
	UnfollowingUserID uint64
}
type FetchFollowersRequest struct{
	UserID uint64
	Limit int
	Offset int
}
type FetchFollowingRequest struct{
	UserID uint64
	Limit int
	Offset int
}
type FetchCommentsReqeust struct{
	PostID uint64
	Limit int
	Offset int
}
type FetchPostByPostIDRequest struct{
	UserID uint64
	PostID uint64
}
type FetchAllPostsReq struct{
	CurrentUserID uint64
	TargetUserID uint64
	Limit int
	Offset int
}

type FetchNewsFeedRequest struct{
	UserID uint64
	Limit int64
	LastID uint64
	PullToRefresh bool
}
type GlobalNewsFeedRequest struct{
	UserID uint64
	Limit int
	//LastScore float64
	Offset int
}

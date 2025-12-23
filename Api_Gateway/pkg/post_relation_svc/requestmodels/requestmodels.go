package requestmodels

type CreatePostRequest struct {
	UserID    uint64
	Caption   string 
	MediaUrls []string
}

type EditPostRequest struct{
	UserID uint64
	PostID uint64
	Caption string `json:"caption"`
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
type FetchCommentsReqeust struct{
	PostID uint64
}
type FetchCommentsOfCommentReqeust struct{
	PostID uint64
	ParentCommentId uint64
}
type FetchAllPostsReq struct{
	UserID uint64
}
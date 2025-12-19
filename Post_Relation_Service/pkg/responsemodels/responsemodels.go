package responsemodels

type CreatePostResponse struct{
	PostID uint64
}

type EditPostResponse struct{
	Caption string
}
type DeletePostResponse struct{
	PostID uint64
}
type LikePostResponse struct{
	PostID uint64
}
type UnlikePostResponse struct{
	PostID uint64
}
type AddCommentResponse struct{
	UserID uint64
	PostID uint64
	CommentText string 
	ParentCommentId *uint64 
}
type EditCommentResponse struct{
	CommentID uint64
	CommentText string
}
type DeleteCommentResponse struct{
	CommentID uint64
}
type FollowResponse struct{
	FollowingUserID uint64
}
type UnfollowResponse struct{
	UnfollowingUserID uint64
}
type Comment struct{
	ID uint64
	CommentText string
}
type FetchCommentsResponse struct{
	Comments []Comment
}
type FetchCommentsOfCommentResponse struct{
	Comments []Comment
}
type PostFollowCountResponse struct{
	PostCount uint64
	FollowerCount uint64
	FollowingCount uint64
}
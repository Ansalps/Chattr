package responsemodels

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Type  string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}
type CreatePostResponse struct {
	PostID uint64
}

type EditPostResponse struct {
	Caption string
}

type DeletePostResponse struct {
	PostID uint64
}
type LikePostResponse struct {
	PostID uint64
}
type UnlikePostResponse struct {
	PostID uint64
}
type AddCommentResponse struct {
	UserID          uint64
	PostID          uint64
	CommentID uint64
	CommentText     string
	ParentCommentId *uint64
}
type EditCommentResponse struct {
	PostID uint64
	CommentID   uint64
	CommentText string
}
type DeleteCommentResponse struct {
	CommentID uint64
}
type FollowResponse struct {
	FollowingUserID uint64
}
type UnfollowResponse struct {
	UnfollowingUserID uint64
}
type UserMetaData struct {
	UserID        uint64
	UserName      string
	Name          string
	ProfileImgUrl string 
	BlueTick      bool
}
type Comment struct {
	CommentID         uint64
	CommentText       string
	CreatedAt         time.Time
	CommentAge        string
	UserDetails       UserMetaData
	ParentCommentID   *uint64
	ChildCommentCount uint64
	ChildComment      []Comment
}
type FetchCommentsResponse struct {
	Comments []Comment
}
type Post struct{
	PostID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID uint64
	Caption string
	MediaUrls []string
	LikeCount     uint64
	CommentsCount uint64
	PostAge       string
}
type FetchAllPostsResponse struct{
	Posts []Post
}

type PostData struct{
	PostID uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID uint64
	Caption string
	MediaUrls []string
	LikeCount     uint64
	CommentsCount uint64
	PostAge       string
	IsLiked bool
	UserData UserMetaData
}

package domain

import "errors"

var (
	ErrForeignKeyViolationCommentPost = errors.New("Post Not found")
	ErrNoFollowers=errors.New("No Followers to Fetch")
	ErrNoFollowing=errors.New("No Following to Fetch")
	ErrNoFollowingNoPost=errors.New("User has No following or following users has no posts or no more data")
	ErrAlreadyFollowing=errors.New("Already following")
	ErrNoFollower=errors.New("not a follower to unfollow")
	ErrNoPostGlobally=errors.New("Global users have no posts")

	ErrPostNotFound     = errors.New("Post Not found or user does not have permission")
	ErrPostIdNotFound=errors.New("Post Not found")
	ErrPostLikeNotFound = errors.New("Post Not found or post has never been liked by the user")
	ErrRecursiveComment = errors.New("can't reply to a comment reply")
	ErrCommentNotFound  = errors.New("comment doesn't exist or post doesn't exist")
	ErrCommentEditDenied=errors.New("user does not have permission to edit comment")
	ErrCommentDeleteDenied=errors.New("user does not have permission to delete comment")
	ErrCommentIdNotFound=errors.New("comment not found")
	ErrFollowOwn        = errors.New("can't follow yourself")
	ErrUserNotFound    = errors.New("User not found")
	ErrUnfollowOwn      = errors.New("can't unfollow yourself")
	ErrNoComments       = errors.New("No comments to Fetch for the Post or Post doesn't exist")
	ErrNoPosts          = errors.New("No Posts to Fetch")

	ErrDatabase=errors.New("Database error: ")
	ErrInvalidParentCommentId=errors.New("Invalid parent comment id")

	ErrInternal=errors.New("internal server error")
	ErrUsersNotFound=errors.New("No users are found")

	CelebPostsNotFound=errors.New("Celeb posts not foun")
	NormalUserPostsNotFound=errors.New("Normal user posts not foun")
)

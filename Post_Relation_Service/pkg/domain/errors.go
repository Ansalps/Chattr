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
	ErrPostLikeNotFound = errors.New("Post Not found or post has never been liked by the user")
	ErrRecursiveComment = errors.New("can't reply to a comment reply")
	ErrCommentNotFound  = errors.New("comment doesn't exist or post doesn't exist or user does not have permission")
	ErrFollowOwn        = errors.New("can't follow yourself")
	ErrUsertNotFound    = errors.New("User not found")
	ErrUnfollowOwn      = errors.New("can't unfollow yourself")
	ErrNoComments       = errors.New("No comments to Fetch for the Post or Post doesn't exist")
	ErrNoPosts          = errors.New("No Posts to Fetch")

	ErrDatabase=errors.New("Database error: ")
)

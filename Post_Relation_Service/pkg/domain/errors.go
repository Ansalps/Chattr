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
)

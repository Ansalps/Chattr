package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotCreatorId          = errors.New("user is not the creator of the group since can't add members")
	ErrUserNotPresent        = errors.New("Member to remove is not present in the group")
	ErrNotGroupMember        = errors.New("cannot add or remove members if you are already not a group member")
	ErrUserNotInConversation = errors.New("unauthorized: user not in conversation")
	ErrGroupNotFound         = errors.New("Group id does not exist")
	ErrNoUsersFound          = errors.New("No users exist or found")
	//ErrNontExistingUsers=errors.New("All or some of the users are non-existent")
	ErrContentTypeNil=errors.New("content type is nil")
)

type NonExistingUsersError struct {
	UserIDs []uint64
}

func (e *NonExistingUsersError) Error() string {
	return fmt.Sprintf("the following user IDs do not exist: %v", e.UserIDs)
}

package domain

import "errors"

var (
	ErrNotCreatorId=errors.New("user is not the creator of the group since can't add members")
	ErrUserNotPresent=errors.New("user is not present in the group")
)

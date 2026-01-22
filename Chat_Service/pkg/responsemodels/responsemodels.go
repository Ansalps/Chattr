package responsemodels

import "time"

type CreateGroupResponse struct {
	GroupID string
}

type AddMembersResponse struct {
	GroupID      string   `json:"group_id" validate:"required"`
	GroupMembers []uint64 `json:"group_members" validate:"required"`
	//CreatorID    uint64   `json:"creator_id"`
	UserID uint64 `json:"user_id"`
}
type RemoveMemberResponse struct {
	GroupID   string `json:"group_id" validate:"required"`
	MemberID  uint64 `json:"group_members" validate:"required"`
	CreatorID uint64 `json:"creator_id"`
}
type ChatProfileResponse struct {
	ChatID          string    `json:"chat_id"`
	ChatName        string    `json:"chat_name"`
	ChatImage       string    `json:"chat_image"` // URL from Auth Service OR Default Group Icon
	LastMessage     string    `json:"last_message"`
	LastMessageTime time.Time `json:"last_message_time"`
	IsGroup         bool      `json:"is_group"`
}

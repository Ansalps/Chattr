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
type GetGroupMembersResponse struct{
	UserID uint64
	UserName string
	ProfileImgUrl string
}
type GetGroupMembers struct{
	GetGroupMembers []GetGroupMembersResponse
	Pagination PaginationDetails
}
type PaginationDetails struct{
	CurrentPage int
	PageSize int
}
type ChatProfileResponse struct {
	ChatID          string    `json:"chat_id"`
	ChatName        string    `json:"chat_name"`
	ChatImage       string    `json:"chat_image"` // URL from Auth Service OR Default Group Icon
	LastMessage     string    `json:"last_message"`
	LastMessageTime time.Time `json:"last_message_time"`
	IsGroup         bool      `json:"is_group"`
	
}

type ChatProfileFinalResponse struct{
	ChatProfiles []ChatProfileResponse
	Pagination PaginationDetails
}

type MessageResponse struct {
    MessageID      string    `json:"message_id"`
    SenderID       uint64    `json:"sender_id"`
    SenderName     string    `json:"sender_name"`
    SenderProfileImgUrl  string    `json:"sender_profile"`
    Content        string    `json:"content"`
    CreatedAt      time.Time `json:"created_at"`
    Status         string    `json:"status"`
}

type GetChatResponse struct {
    ConversationID string            `json:"conversation_id"`
    Messages       []MessageResponse `json:"messages"`
    HasMore        bool              `json:"has_more"`
	Pagination PaginationDetails
}

type GroupMeta struct {
	Name     string
	ImageURL string
}

package requestmodels

import (
	"time"
)

type CreateGroupRequest struct {
	GroupID      string
	GroupMembers []uint64 `json:"group_members"`
	GroupName    string   `json:"group_name"`
	CreatorID    uint64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AddMembersRequest struct {
	GroupID      string   `json:"group_id" validate:"required"`
	GroupMembers []uint64 `json:"group_members" validate:"required"`
	UserID       uint64   `json:"user_id"`
}

type RemoveMemberRequest struct {
	GroupID  string `json:"group_id" validate:"required"`
	MemberID uint64 `json:"member_id" validate:"required"`
	UserID   uint64 `json:"user_id"`
}

// type DirectMessage struct {
// 	Type    string `json:"string"`
// 	To      uint64 `json:"to"`
// 	Message string `json:"message"`
// }

// type ChatMessage struct {
// 	ID          primitive.ObjectID `bson:"_id,omitempty"`
// 	SenderID    string             `json:"sender_id" validate:"required"`
// 	RecipientID string             `json:"recipient_id" validate:"required"`
// 	//Type string `json:"type"`
// 	Content   string    `json:"content"`
// 	CreatedAt time.Time `json:"created_at"`
// }

type MessageRequest struct {
	SenderID    uint64    `json:"sender_id" validate:"required"`
	RecipientID uint64    `json:"recipient_id"`
	Content     string    `json:"content" `
	CreatedAt   time.Time `json:"created_at"`
	Type        string    `json:"type" validate:"required"`
	//Status      string    `json:"Status"`
	GroupID string `json:"group_id"`
	//TypingStat  bool
}

type RecentChatProfilesRequest struct{
	UserID uint64
	Limit int
	Offset int
}
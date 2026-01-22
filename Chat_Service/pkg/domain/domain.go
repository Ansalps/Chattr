package domain

import (
	"time"
)

type Message struct {
	MessageID   string
	SenderID    uint64    `json:"sender_id" validate:"required"`
	RecipientID uint64    `json:"recipient_id"`
	GroupID string `json:"group_id"`
	Content     string    `json:"content" `
	CreatedAt   time.Time `json:"created_at"`
	Type        string    `json:"type" validate:"required"`
}

type Conversation struct{
	ConversationID string `bson:"conversation_id"`
	Participants []uint64 `bson:"participants"`
	GroupID string `bson:"group_id,omitempty"`
	LastMessage string `bson:"last_message"`
	LastMessageTime time.Time `bson:"last_message_time"`
	Type string `bson:"type"`
}
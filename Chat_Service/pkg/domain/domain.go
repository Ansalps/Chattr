package domain

import (
	"time"
)

type Message struct {
	MessageID   string	`json:"message_id" bson:"message_id"`
	ConversationID string	`json:"conversation_id" bson:"conversation_id"`
	SenderID    uint64    `json:"sender_id" bson:"sender_id" validate:"required"`
	RecipientID uint64    `json:"recipient_id" bson:"recipient_id,omitempty"`
	GroupID string `json:"group_id" bson:"group_id,omit_empty"`
	Content     string    `json:"content" bson:"content"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	Type        string    `json:"type" bson:"type" validate:"required"`
	Status         string    `json:"status" bson:"status"`
}

type Conversation struct{
	ConversationID string `bson:"conversation_id"`
	Participants []uint64 `bson:"participants"`
	GroupID string `bson:"group_id,omitempty"`
	LastMessage string `bson:"last_message"`
	LastMessageTime time.Time `bson:"last_message_time"`
	Type string `bson:"type"`
}
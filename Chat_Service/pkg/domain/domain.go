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
	ConversationID string
	Participants []uint64
	GroupID string
	LastMessage string
	LastMessageTime time.Time
	Type string
}
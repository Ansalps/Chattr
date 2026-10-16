package events

type PostCreatedEvent struct {
	Type      string `json:"type"`
	UserID    uint64 `json:"userId"`
	PostID    uint64 `json:"postId"`
	CreatedAt int64  `json:"createdAt"`
}
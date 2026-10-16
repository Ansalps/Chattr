package requestmodels

type PostEvent struct {
	Type        string `json:"type"`
	ActorID     uint64 `json:"actorId"`
	PostID      uint64 `json:"postId"`
	PostOwnerID uint64 `json:"postOwnerId"`
}

type UserEvent struct {
	Type        string `json:"type"`
	ActorID     uint64 `json:"actorId"`
	FollowingID      uint64 `json:"followingId"`
	//PostOwnerID uint64 `json:"postOwnerId"`
}

type DirectMessageEvent struct{
	Type string `json:"type"`
	ActorID     uint64 `json:"actorId"`
	RecipientID      uint64 `json:"recipientId"`
	ConversationID string `json:"conversationId"`
}

type GroupMessageEvent struct{
	Type string `json:"type"`
	ActorID     uint64 `json:"actorId"`
	RecipientID      []uint64 `json:"recipientId"`
	GroupName string `json:"groupName"`
	ConversationID string `json:"conversationId"`
}

type GetNotificationsequest struct{
	UserID uint64
	Limit int
	Offset int
}
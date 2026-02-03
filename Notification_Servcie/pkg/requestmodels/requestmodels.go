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
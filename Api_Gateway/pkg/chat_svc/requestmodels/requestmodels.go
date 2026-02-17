package requestmodels

type CreateGroupRequest struct {
	//GroupID uuid.UUID
	GroupMembers []uint64 `json:"group_members"`
	GroupName    string   `json:"group_name"`
	CreatorID    uint64
	// CreatedAt    time.Time
	// UpdatedAt    time.Time
}

type AddMembersRequest struct {
	GroupID      string   `json:"group_id"`
	GroupMembers []uint64 `json:"group_members" validate:"required"`
}

type RemoveMembersRequest struct {
	GroupID      string `json:"group_id"`
	UserID uint64 `json:"user_id"`  
	MemberID uint64 `json:"member_id" validate:"required"`
}

type GroupProfileImageRequest struct{
    ContentType string `json:"content_type" binding:"required"`
    Image       []byte	`json:"image" binding:"required"`
}
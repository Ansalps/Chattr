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
	GroupID      string   
	GroupMembers []uint64 `json:"group_members" validate:"required"`
}

type RemoveMembersRequest struct {
	GroupID      string   
	MemberID uint64 `json:"member_id" validate:"required"`
}

package responsemodels

type CreateGroupResponse struct {
	GroupID string
}

type AddMembersResponse struct {
	GroupID      string   `json:"group_id" validate:"required"`
	GroupMembers []uint64 `json:"group_members" validate:"required"`
	CreatorID    uint64   `json:"creator_id"`
}
type RemoveMemberResponse struct {
	GroupID   string `json:"group_id" validate:"required"`
	MemberID  uint64 `json:"group_members" validate:"required"`
	CreatorID uint64 `json:"creator_id"`
}

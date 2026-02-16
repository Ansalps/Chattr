package interfacesrepository

import (
	"context"

	"github.com/Ansalps/Chattr_Chat_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/responsemodels"
)

type ChatRepository interface {
	CreateGroup(requestmodels.CreateGroupRequest) (responsemodels.CreateGroupResponse, error)
	ExistingMembers(string) ([]uint64, error)
	AddMembers(requestmodels.AddMembersRequest) (responsemodels.AddMembersResponse, error)
	CreatorID(string) (uint64, error)
	RemoveMember(requestmodels.RemoveMemberRequest) (responsemodels.RemoveMemberResponse, error)
	FetchMembersOfGroup(groupId string) ([]uint64, error)

	StoreIndividualChatInMessages(domain.Message) error
	StoreOrUpdateIndividualChatInConversation(conversation domain.Conversation) (string, error)

	StoreGroupChatInMessages(domain.Message) error
	StoreOrUpdateGroupChatInConversation(domain.Conversation) (string, error)

	GetUserConversation(req requestmodels.RecentChatProfilesRequest) ([]domain.Conversation, error)

	GetGroupNamesBatch(groupIDs []string) (map[string]string, error)

	GetUserMessagesByConversationId(req requestmodels.GetChatRequest) ([]domain.Message, error)

	IsUserInConversation(convID string, userID uint64) (bool, error)

	GetGroupNameByGroupID(groupID string) (string, error)

	GroupExists(ctx context.Context, groupID string) (bool, error)
}

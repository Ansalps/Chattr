package interfacesUsecase

import (
	"github.com/Ansalps/Chattr_Chat_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/responsemodels"
)

type ChatUsecase interface {
	CreateGroup(requestmodels.CreateGroupRequest) (responsemodels.CreateGroupResponse, error)
	AddMembers(requestmodels.AddMembersRequest) (responsemodels.AddMembersResponse, error)
	RemoveMember(requestmodels.RemoveMemberRequest) (responsemodels.RemoveMemberResponse, error)
	FetchMembersOfGroup(groupId string)([]uint64,error)

	StoreIndividualChatInMessages(domain.Message)(error)
	StoreOrUpdateIndividualChatInConversation(domain.Conversation)(string,error)

	StoreGroupChatInMessages(domain.Message)error
	StoreOrUpdateGroupChatInConversation(domain.Conversation)(string,error)

	GetRecentChatProfiles(requestmodels.RecentChatProfilesRequest)([]responsemodels.ChatProfileResponse,error)
	GetChat(requestmodels.GetChatRequest)(responsemodels.GetChatResponse,error)

	//PublishEvent(topic string, message interface{}) error
	GetGroupName(groupID string) (string, error)
}

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
	StoreOrUpdateIndividualChatInConversation(domain.Conversation)(error)

	StoreGroupChatInMessages(domain.Message)error
	StoreOrUpdateGroupChatInConversation(domain.Conversation)(error)

	GetRecentChatProfiles(requestmodels.RecentChatProfilesRequest)([]responsemodels.ChatProfileResponse,error)

}

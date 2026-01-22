package usecase

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/Ansalps/Chattr_Chat_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/pb"
	interfacesrepository "github.com/Ansalps/Chattr_Chat_Service/pkg/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/usecase/interfacesUsecase"
)

type ChatUsecase struct {
	ChatRepository interfacesrepository.ChatRepository
	AuthClient     pb.AuthSubscriptionServiceClient
}

func NewChatUsecase(repository interfacesrepository.ChatRepository, authClient pb.AuthSubscriptionServiceClient) interfacesUsecase.ChatUsecase {
	return &ChatUsecase{
		ChatRepository: repository,
		AuthClient:     authClient,
	}
}

func (as *ChatUsecase) CreateGroup(req requestmodels.CreateGroupRequest) (responsemodels.CreateGroupResponse, error) {
	resp, err := as.ChatRepository.CreateGroup(req)
	if err != nil {
		log.Println(err)
		return responsemodels.CreateGroupResponse{}, err
	}
	return responsemodels.CreateGroupResponse{
		GroupID: resp.GroupID,
	}, nil
}

func (as *ChatUsecase) AddMembers(req requestmodels.AddMembersRequest) (responsemodels.AddMembersResponse, error) {
	resp1, err := as.ChatRepository.ExistingMembers(req.GroupID)
	if err != nil {
		return responsemodels.AddMembersResponse{}, err
	}

	if !slices.Contains(resp1, req.UserID) {
		return responsemodels.AddMembersResponse{}, domain.ErrNotGroupMember
	}
	resp, err := as.AuthClient.CheckUserListExists(context.Background(), &pb.UserDataReq{
		UserId: req.GroupMembers,
	})
	if err != nil {
		log.Println(err)
		return responsemodels.AddMembersResponse{}, err
	}

	// fmt.Println("three")
	// req.GroupMembers = slices.DeleteFunc(resp1, func(id uint64) bool {
	// 	return slices.Contains(resp.UserId, id)
	// })
	// var groupmembers []int64
	// for i:=range resp.UserId{
	// 	groupmembers = append(groupmembers, int64(req.GroupMembers[i]))
	// }
	req.GroupMembers = resp.UserId
	//fmt.Println("group members",groupmembers)
	//fmt.Println("four")
	resp2, err := as.ChatRepository.AddMembers(req)
	if err != nil {
		return responsemodels.AddMembersResponse{}, err
	}
	resp2.UserID = req.UserID
	fmt.Println("resp2", resp2)
	return resp2, nil
}

func (as *ChatUsecase) RemoveMember(req requestmodels.RemoveMemberRequest) (responsemodels.RemoveMemberResponse, error) {
	// creatorId, err := as.ChatRepository.CreatorID(req.GroupID)
	// if err != nil {
	// 	return responsemodels.RemoveMemberResponse{}, nil
	// }
	// if creatorId != req.UserID {
	// 	return responsemodels.RemoveMemberResponse{}, domain.ErrNotCreatorId
	// }
	resp1, err := as.ChatRepository.ExistingMembers(req.GroupID)
	if err != nil {
		return responsemodels.RemoveMemberResponse{}, err
	}
	if !slices.Contains(resp1, req.UserID) {
		return responsemodels.RemoveMemberResponse{}, domain.ErrNotGroupMember
	}
	if !slices.Contains(resp1, req.MemberID) {
		return responsemodels.RemoveMemberResponse{}, domain.ErrUserNotPresent
	}
	resp2, err := as.ChatRepository.RemoveMember(req)
	if err != nil {
		return responsemodels.RemoveMemberResponse{}, err
	}
	return resp2, nil
}

func (as *ChatUsecase) FetchMembersOfGroup(groupId string) ([]uint64, error) {
	userIds, err := as.ChatRepository.FetchMembersOfGroup(groupId)
	if err != nil {
		return nil, err
	}
	return userIds, nil
}

func (as *ChatUsecase) StoreIndividualChatInMessages(dm domain.Message) error {
	err := as.ChatRepository.StoreIndividualChatInMessages(dm)
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}
func (as *ChatUsecase) StoreOrUpdateIndividualChatInConversation(conversation domain.Conversation) error {
	err := as.ChatRepository.StoreOrUpdateIndividualChatInConversation(conversation)
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}

func (as *ChatUsecase) StoreGroupChatInMessages(gm domain.Message) error {
	err := as.ChatRepository.StoreGroupChatInMessages(gm)
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}

func (as *ChatUsecase) StoreOrUpdateGroupChatInConversation(conversation domain.Conversation) error {
	err := as.ChatRepository.StoreOrUpdateGroupChatInConversation(conversation)
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}

func (as *ChatUsecase) GetRecentChatProfiles(req requestmodels.RecentChatProfilesRequest) ([]responsemodels.ChatProfileResponse, error) {
	fmt.Println("req.UserID",req.UserID)
	convs, err := as.ChatRepository.GetUserConversation(req)
	if err != nil {
		return nil, err
	}

	// 2. Separate Group IDs and unique User IDs
    var individualUserIDs []uint64
	var groupIDs []string
    userIDsMap := make(map[uint64]bool)
	fmt.Println("convs",convs)
    for _, c := range convs {
        if c.Type == "individual" {
            for _, pID := range c.Participants {
                if pID != req.UserID && !userIDsMap[pID] {
                    individualUserIDs = append(individualUserIDs, pID)
                    userIDsMap[pID] = true
                }
            }
        } else if c.Type == "group" && c.GroupID != "" {
            groupIDs = append(groupIDs, c.GroupID)
        }
    }

	// 2. One Batch Call for Groups
    groupNames, err := as.ChatRepository.GetGroupNamesBatch(groupIDs) // Map[ID]Name
	if err!=nil{
		log.Println("error fetching group names")
	}

	// 3. Batch Call to Auth Service
    // Request: { user_ids: [10, 11, ...] }
    // Response: { user_metadata_map: { "10": {name: "Ansal", img: "..."}, "11": {...} } }
	fmt.Println("individualUserIDs",individualUserIDs)
    authRes, err := as.AuthClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
		UserId: individualUserIDs})
    if err != nil {
        log.Printf("Auth service batch call failed: %v", err)
    }

	// 4. Build the final response list
    var finalProfiles []responsemodels.ChatProfileResponse
    for _, conv := range convs {
        profile := responsemodels.ChatProfileResponse{
            ChatID:          conv.ConversationID,
            LastMessage:     conv.LastMessage,
            LastMessageTime: conv.LastMessageTime,
            IsGroup:         conv.Type == "group",
        }

        if conv.Type == "group" {
            profile.ChatName = groupNames[conv.GroupID]
        } else {
            // Find the "other" person in this chat
            for _, pID := range conv.Participants {
                if pID != req.UserID {
                    if meta, ok := authRes.Users[pID]; ok {
                        profile.ChatName = meta.UserName
                        profile.ChatImage = meta.ProfileImgUrl
                    }
                    break
                }
            }
        }
        finalProfiles = append(finalProfiles, profile)
    }

    return finalProfiles, nil
}

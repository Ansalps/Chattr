package usecase

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/Ansalps/Chattr_Chat_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/pb"
	interfacesrepository "github.com/Ansalps/Chattr_Chat_Service/pkg/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/usecase/interfacesUsecase"
	"github.com/google/uuid"
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


func (as *ChatUsecase) GetGroupName(groupID string) (string, error) {
	if groupID == "" {
		return "", fmt.Errorf("groupID cannot be empty")
	}

	return as.ChatRepository.GetGroupNameByGroupID(groupID)
}

func (as *ChatUsecase) CreateGroup(req requestmodels.CreateGroupRequest) (responsemodels.CreateGroupResponse, error) {
	// 1. Create the Group entry in the 'groups' collection
	resp, err := as.ChatRepository.CreateGroup(req)
	if err != nil {
		log.Println(err)
		return responsemodels.CreateGroupResponse{}, err
	}

	// 2. Initialize the Conversation record
	// This ensures the group shows up in RecentChatProfiles for all members
	syncTime := time.Now()
	conversationStruct := domain.Conversation{
		ConversationID:  uuid.NewString(),
		Participants:    req.GroupMembers, // List of UserIDs
		GroupID:         resp.GroupID,
		LastMessage:     "Group created", // Optional preview text
		LastMessageTime: syncTime,
		Type:            "group",
	}

	// We call the Repository function we modified earlier to return a string
	_, err = as.ChatRepository.StoreOrUpdateGroupChatInConversation(conversationStruct)
	if err != nil {
		// We log the error but still return the GroupID because the group was created
		log.Printf("Warning: Group %s created but conversation sync failed: %v", resp.GroupID, err)
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
func (as *ChatUsecase) StoreOrUpdateIndividualChatInConversation(conversation domain.Conversation) (string, error) {
	convID, err := as.ChatRepository.StoreOrUpdateIndividualChatInConversation(conversation)
	if err != nil {
		log.Printf("Usecase Error: StoreOrUpdateIndividualChatInConversation failed: %v", err)
		return "", err
	}
	return convID, nil
}

func (as *ChatUsecase) StoreGroupChatInMessages(gm domain.Message) error {
	err := as.ChatRepository.StoreGroupChatInMessages(gm)
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}

func (as *ChatUsecase) StoreOrUpdateGroupChatInConversation(conversation domain.Conversation) (string, error) {
	convID, err := as.ChatRepository.StoreOrUpdateGroupChatInConversation(conversation)
	if err != nil {
		log.Println("Usecase Error:", err)
		return "", err
	}
	return convID, nil
}

func (as *ChatUsecase) GetRecentChatProfiles(req requestmodels.RecentChatProfilesRequest) ([]responsemodels.ChatProfileResponse, error) {
	fmt.Println("req.UserID", req.UserID)
	convs, err := as.ChatRepository.GetUserConversation(req)
	if err != nil {
		return nil, err
	}

	// 2. Separate Group IDs and unique User IDs
	var individualUserIDs []uint64
	var groupIDs []string
	userIDsMap := make(map[uint64]bool)
	fmt.Println("convs", convs)
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
	if err != nil {
		log.Println("error fetching group names")
	}

	// 3. Batch Call to Auth Service
	// Request: { user_ids: [10, 11, ...] }
	// Response: { user_metadata_map: { "10": {name: "Ansal", img: "..."}, "11": {...} } }
	fmt.Println("individualUserIDs", individualUserIDs)
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

func (as *ChatUsecase)GetChat(req requestmodels.GetChatRequest)(responsemodels.GetChatResponse,error){
	fmt.Println("req.UserID", req.UserID)

	// 1. SECURITY VALIDATION
    // Check if the UserID is a participant in this Conversation
    isParticipant, err := as.ChatRepository.IsUserInConversation(req.ConvID, req.UserID)
    if err != nil || !isParticipant {
        log.Printf("Unauthorized access attempt: User %d for ConvID %s", req.UserID, req.ConvID)
        return responsemodels.GetChatResponse{}, domain.ErrUserNotInConversation
    }

	// 1. Request one extra message to check if there's more data
    originalLimit := req.Limit
    req.Limit = originalLimit + 1


	messages, err := as.ChatRepository.GetUserMessagesByConversationId(req)
	if err != nil {
		log.Printf("Error fetching messages for ConvID %s: %v", req.ConvID, err)
		return responsemodels.GetChatResponse{}, err
	}

	hasMore := false
    // 2. If we got back more than the original limit, we know a next page exists
    if len(messages) > originalLimit {
        hasMore = true
        messages = messages[:originalLimit] // Remove the extra message before returning
    }

	// 3. Reverse the slice so oldest is first, newest is last (UI friendly)
    for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
        messages[i], messages[j] = messages[j], messages[i]
    }

	// 2. Collect unique SenderIDs to batch-fetch names/profiles
    senderIDsMap := make(map[uint64]bool)
    var uniqueIDs []uint64
    for _, msg := range messages {
        if !senderIDsMap[msg.SenderID] {
            uniqueIDs = append(uniqueIDs, msg.SenderID)
            senderIDsMap[msg.SenderID] = true
        }
    }
	// 3. Batch Call to Auth Service (gRPC)
    var authRes *pb.BatchUserMetadataResponse // Replace with your actual gRPC generated package name
    if len(uniqueIDs) > 0 {
        authRes, err = as.AuthClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
            UserId: uniqueIDs,
        })
        if err != nil {
            log.Printf("Warning: Auth service call failed: %v", err)
            // We continue so the user sees message text even if profiles fail
        }
    }

	// 4. Map to response models and hydrate sender data
    var messageResponses []responsemodels.MessageResponse
    for _, msg := range messages {
        mResp := responsemodels.MessageResponse{
            MessageID:   msg.MessageID,
            SenderID:    msg.SenderID,
            Content:     msg.Content,
            CreatedAt:   msg.CreatedAt,
            Status:      msg.Status,
        }

		// Fill Name and Profile if available from gRPC response
        if authRes != nil && authRes.Users != nil {
            if user, ok := authRes.Users[msg.SenderID]; ok {
                mResp.SenderName = user.UserName
                mResp.SenderProfileImgUrl = user.ProfileImgUrl
            }
        }
		messageResponses = append(messageResponses, mResp)
		
	}
	// 5. Final Response
	return responsemodels.GetChatResponse{
		ConversationID: req.ConvID,
		Messages:       messageResponses,
		HasMore:        hasMore, // Frontend now knows when to stop scrolling!
	}, nil
}
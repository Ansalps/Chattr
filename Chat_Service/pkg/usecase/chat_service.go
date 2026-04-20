package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Ansalps/Chattr_Chat_Service/pkg/client"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/config"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/pb"
	interfacesrepository "github.com/Ansalps/Chattr_Chat_Service/pkg/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/usecase/interfacesUsecase"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatUsecase struct {
	ChatRepository interfacesrepository.ChatRepository
	AuthClient     pb.AuthSubscriptionServiceClient
	AwsS3Client    *s3.Client
	AwsBucket      string
	Config         *config.Config
}

func NewChatUsecase(repository interfacesrepository.ChatRepository, authClient pb.AuthSubscriptionServiceClient, awsS3Client *s3.Client, awsBucket string, config *config.Config) interfacesUsecase.ChatUsecase {
	return &ChatUsecase{
		ChatRepository: repository,
		AuthClient:     authClient,
		AwsS3Client:    awsS3Client,
		AwsBucket:      awsBucket,
		Config:         config,
	}
}

func (as *ChatUsecase) SetGroupProfileImage(req requestmodels.GroupProfileImageRequest) (string, error) {
	exists, err := as.ChatRepository.GroupExists(context.Background(), req.GroupID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return "", fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	if !exists {
		return "", domain.ErrGroupNotFound
	}
	userIds, err := as.ChatRepository.FetchMembersOfGroup(req.GroupID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", fmt.Errorf("%w: %v", domain.ErrGroupNotFound, err)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return "", fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return "", fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	if !slices.Contains(userIds, req.UserID) {
		return "", domain.ErrNotGroupMember
	}
	ct := req.ContentType

	ct = strings.TrimPrefix(ct, "image/")

	filename := fmt.Sprintf("%d_%d.%s", req.UserID, time.Now().Unix(), ct)
	//fmt.Println("file name", filename)
	key := "group_profiles/" + filename
	//fmt.Println("inside usecase type", setProfileImageReq.ContentType)
	if req.ContentType == "" {
		return "", domain.ErrContentTypeNil /*fmt.Errorf("content type is nil")*/
	}
	//fmt.Println("hi hello",aws.String(as.AwsBucket),aws.String(key),as.AwsBucket,key)
	uploader := manager.NewUploader(as.AwsS3Client)
	_, err = uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(as.AwsBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(req.Image),
		ContentType: aws.String(req.ContentType),
	})
	if err != nil {
		//fmt.Println("is it here")
		return "", fmt.Errorf("%w: %v", domain.ErrS3UploadFail, err)
	}
	// Construct URL
	imageURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", as.AwsBucket, as.Config.Aws.AwsRegion, key)
	// Save to DB
	err = as.ChatRepository.SetGroupProfileImage(req.GroupID, imageURL)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}
		if errors.Is(err, domain.ErrGroupNotFound) {
			return "", domain.ErrGroupNotFound
		}

		return "", fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	return imageURL, nil
}

func (as *ChatUsecase) DoesUserExist(userid uint64) (bool, error) {
	resp, err := as.AuthClient.CheckUserExists(context.Background(), &pb.CheckUserExistsRequest{
		UserId: userid,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Not a gRPC error
			return false,
				fmt.Errorf("%w: %v", domain.ErrInternal, err)
		}

		switch st.Code() {

		case codes.NotFound:
			return false, domain.ErrUserNotFound

		case codes.Internal:
			return false,
				fmt.Errorf("%w: %v", domain.ErrDatabase, err)

		default:
			return false,
				fmt.Errorf("%w: %v", domain.ErrInternal, err)
		}
	}
	return resp.Exists, nil
}

func (as *ChatUsecase) GetGroupName(groupID string) (string, error) {
	if groupID == "" {
		return "", fmt.Errorf("groupID cannot be empty")
	}

	return as.ChatRepository.GetGroupNameByGroupID(groupID)
}

func (as *ChatUsecase) CreateGroup(req requestmodels.CreateGroupRequest) (responsemodels.CreateGroupResponse, error) {
	allUsersNotExists, err := as.AuthClient.CheckAllUsersExists(context.Background(), &pb.UserDataReq{
		UserId: req.GroupMembers,
	})
	if err != nil {
		return responsemodels.CreateGroupResponse{}, fmt.Errorf("%w: %v: %v", domain.ErrInternal, "failed internal service call on auth service", err)
	}
	if len(allUsersNotExists.UserId) != 0 {
		return responsemodels.CreateGroupResponse{}, &domain.NonExistingUsersError{
			UserIDs: allUsersNotExists.UserId,
		}
	}

	resp, err := as.ChatRepository.CreateGroup(req)
	if err != nil {
		//log.Println(err)
		if errors.Is(err, context.DeadlineExceeded) {
			return responsemodels.CreateGroupResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return responsemodels.CreateGroupResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
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
		//log.Printf("Warning: Group %s created but conversation sync failed: %v", resp.GroupID, err)
		if errors.Is(err, context.DeadlineExceeded) {
			return responsemodels.CreateGroupResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}
		if errors.Is(err, mongo.ErrNoDocuments) {
			return responsemodels.CreateGroupResponse{}, fmt.Errorf("%w: %v", domain.ErrInvalidGroupID, err)
		}
		return responsemodels.CreateGroupResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	return responsemodels.CreateGroupResponse{
		GroupID: resp.GroupID,
	}, nil
}

func (as *ChatUsecase) AddMembers(req requestmodels.AddMembersRequest) (responsemodels.AddMembersResponse, error) {
	allUsersNotExists, err := as.AuthClient.CheckAllUsersExists(context.Background(), &pb.UserDataReq{
		UserId: req.GroupMembers,
	})
	if err != nil {
		return responsemodels.AddMembersResponse{}, fmt.Errorf("%w: %v: %v", domain.ErrInternal, "failed internal service call on auth service", err)
	}
	if len(allUsersNotExists.UserId) != 0 {
		return responsemodels.AddMembersResponse{}, &domain.NonExistingUsersError{
			UserIDs: allUsersNotExists.UserId,
		}
	}
	exists, err := as.ChatRepository.GroupExists(context.Background(), req.GroupID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return responsemodels.AddMembersResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return responsemodels.AddMembersResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	if !exists {
		return responsemodels.AddMembersResponse{}, domain.ErrGroupNotFound
	}
	resp1, err := as.ChatRepository.ExistingMembers(req.GroupID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return responsemodels.AddMembersResponse{}, fmt.Errorf("%w: %v", domain.ErrGroupNotFound, err)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return responsemodels.AddMembersResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return responsemodels.AddMembersResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	if !slices.Contains(resp1, req.UserID) {
		return responsemodels.AddMembersResponse{}, domain.ErrNotGroupMember
	}
	// resp, err := as.AuthClient.CheckUserListExists(context.Background(), &pb.UserDataReq{
	// 	UserId: req.GroupMembers,
	// })
	// if err != nil {
	// 	log.Println(err)
	// 	return responsemodels.AddMembersResponse{}, err
	// }
	// req.GroupMembers = slices.DeleteFunc(resp1, func(id uint64) bool {
	// 	return slices.Contains(resp.UserId, id)
	// })
	// var groupmembers []int64
	// for i:=range resp.UserId{
	// 	groupmembers = append(groupmembers, int64(req.GroupMembers[i]))
	// }
	//req.GroupMembers = resp.UserId
	resp2, err := as.ChatRepository.AddMembers(req)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return responsemodels.AddMembersResponse{}, domain.ErrGroupNotFound
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return responsemodels.AddMembersResponse{}, domain.ErrDatabaseTimeout
		}

		return responsemodels.AddMembersResponse{}, domain.ErrInternal
	}
	resp2.UserID = req.UserID
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
	exists, err := as.ChatRepository.GroupExists(context.Background(), req.GroupID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return responsemodels.RemoveMemberResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return responsemodels.RemoveMemberResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	if !exists {
		return responsemodels.RemoveMemberResponse{}, domain.ErrGroupNotFound
	}
	resp1, err := as.ChatRepository.ExistingMembers(req.GroupID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return responsemodels.RemoveMemberResponse{}, fmt.Errorf("%w: %v", domain.ErrGroupNotFound, err)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return responsemodels.RemoveMemberResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return responsemodels.RemoveMemberResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	if !slices.Contains(resp1, req.UserID) {
		return responsemodels.RemoveMemberResponse{}, domain.ErrNotGroupMember
	}
	if !slices.Contains(resp1, req.MemberID) {
		return responsemodels.RemoveMemberResponse{}, domain.ErrUserNotPresent
	}
	resp2, err := as.ChatRepository.RemoveMember(req)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return responsemodels.RemoveMemberResponse{}, fmt.Errorf("%w: %v", domain.ErrGroupNotFound, err)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return responsemodels.RemoveMemberResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return responsemodels.RemoveMemberResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	return resp2, nil
}

func (as *ChatUsecase) GetGroupMembers(req requestmodels.GetGroupMembersRequest) ([]responsemodels.GetGroupMembersResponse, error) {

	exists, err := as.ChatRepository.GroupExists(context.Background(), req.GroupID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	if !exists {
		return nil, domain.ErrGroupNotFound
	}
	userIds, err := as.ChatRepository.FetchMembersOfGroup(req.GroupID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: %v", domain.ErrGroupNotFound, err)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	if !slices.Contains(userIds, req.UserID) {
		return nil, domain.ErrNotGroupMember
	}
	// ✅ Apply pagination
	start := req.Offset
	end := req.Offset + req.Limit

	if start > len(userIds) {
		return []responsemodels.GetGroupMembersResponse{}, nil
	}

	if end > len(userIds) {
		end = len(userIds)
	}

	userIds = userIds[start:end]
	authRes, err := client.FetchUserMetaData(
		as.AuthClient,
		userIds,
	)
	if err != nil {
		return nil, err
	}
	// authRes, err := as.AuthClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
	// 	UserId: userIds})
	// if err != nil {
	// 	st, ok := status.FromError(err)
	// 	if !ok {
	// 		// Not a gRPC error
	// 		return nil,
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}

	// 	switch st.Code() {

	// 	case codes.NotFound:
	// 		return nil, domain.ErrUsersNotFound

	// 	case codes.Internal:
	// 		return nil,
	// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)

	// 	default:
	// 		return nil,
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}
	// }
	//fmt.Println("",authRes.Users)

	resp := make([]responsemodels.GetGroupMembersResponse, 0)
	for _, v := range userIds {
		r := responsemodels.GetGroupMembersResponse{
			UserID:        authRes.Users[v].UserId,
			UserName:      authRes.Users[v].UserName,
			ProfileImgUrl: authRes.Users[v].ProfileImgUrl,
		}
		resp = append(resp, r)
	}
	return resp, nil
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
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	return nil
}
func (as *ChatUsecase) StoreOrUpdateIndividualChatInConversation(conversation domain.Conversation) (string, error) {
	convID, err := as.ChatRepository.StoreOrUpdateIndividualChatInConversation(conversation)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}

		return "", fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	return convID, nil
}

func (as *ChatUsecase) StoreGroupChatInMessages(gm domain.Message) error {
	err := as.ChatRepository.StoreGroupChatInMessages(gm)
	if err != nil {
		//log.Println(err)
		return err
	}
	return err
}

func (as *ChatUsecase) StoreOrUpdateGroupChatInConversation(conversation domain.Conversation) (string, error) {
	convID, err := as.ChatRepository.StoreOrUpdateGroupChatInConversation(conversation)
	if err != nil {
		//log.Println("Usecase Error:", err)
		return "", err
	}
	return convID, nil
}

func (as *ChatUsecase) GetRecentChatProfiles(req requestmodels.RecentChatProfilesRequest) ([]responsemodels.ChatProfileResponse, error) {
	fmt.Println("req.UserID", req.UserID)
	convs, err := as.ChatRepository.GetUserConversation(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	if len(convs)==0{
		return []responsemodels.ChatProfileResponse{},nil
	}
	// 2. Separate Group IDs and unique User IDs
	var individualUserIDs []uint64
	var groupIDs []string
	userIDsMap := make(map[uint64]bool)
	//fmt.Println("convs", convs)
	for _, c := range convs {
		if c.Type == "individual" {
			for _, pID := range c.Participants {
				if pID != req.UserID && !userIDsMap[pID] {
					individualUserIDs = append(individualUserIDs, pID)
					individualUserIDs=append(individualUserIDs, req.UserID)
					userIDsMap[pID] = true
				}
			}
		} else if c.Type == "group" && c.GroupID != "" {
			groupIDs = append(groupIDs, c.GroupID)
		}
	}
	var groupMeta map[string]responsemodels.GroupMeta
	if len(groupIDs) > 0 {
		// 2. One Batch Call for Groups
		//groupNames, err := as.ChatRepository.GetGroupNamesBatch(groupIDs) // Map[ID]Name
		groupMeta, err = as.ChatRepository.GetGroupMetaBatch(groupIDs)

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
			}
			if err != mongo.ErrNoDocuments {
				return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
			}
		}
	}

	var authRes *pb.BatchUserMetadataResponse
	if len(individualUserIDs)>0{
	authRes, err = client.FetchUserMetaData(
		as.AuthClient,
		individualUserIDs,
	)
	if err != nil {
		return nil, err
	}
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
			meta := groupMeta[conv.GroupID]

			profile.ChatName = meta.Name
			profile.ChatImage = meta.ImageURL // empty if not set
			profile.GroupID=conv.GroupID
		} else {
			// Find the "other" person in this chat
			for i, pID := range conv.Participants {
				if pID != req.UserID {
					if meta, ok := authRes.Users[pID]; ok {
						profile.ChatName = meta.UserName
						profile.ChatImage = meta.ProfileImgUrl
					}
					break
				} else if conv.Participants[i]==req.UserID && conv.Participants[i+1]==req.UserID{
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

func (as *ChatUsecase) GetChat(req requestmodels.GetChatRequest) (responsemodels.GetChatResponse, error) {
	// 1. SECURITY VALIDATION
	// Check if the UserID is a participant in this Conversation
	isParticipant, err := as.ChatRepository.IsUserInConversation(req.ConvID, req.UserID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return responsemodels.GetChatResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabaseTimeout, err)
		}
		//log.Printf("Unauthorized access attempt: User %d for ConvID %s", req.UserID, req.ConvID)
		return responsemodels.GetChatResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	if !isParticipant {
		return responsemodels.GetChatResponse{}, domain.ErrUserNotInConversation
	}

	// 1. Request one extra message to check if there's more data
	originalLimit := req.Limit
	req.Limit = originalLimit + 1

	messages, err := as.ChatRepository.GetUserMessagesByConversationId(req)
	if err != nil {
		return responsemodels.GetChatResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
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
		authRes, err = client.FetchUserMetaData(
			as.AuthClient,
			uniqueIDs,
		)
		if err != nil {
			return responsemodels.GetChatResponse{}, err
		}
		// authRes, err = as.AuthClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
		// 	UserId: uniqueIDs,
		// })
		// if err != nil {
		// 	st, ok := status.FromError(err)
		// 	if !ok {
		// 		// Not a gRPC error
		// 		return responsemodels.GetChatResponse{},
		// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
		// 	}

		// 	switch st.Code() {

		// 	case codes.NotFound:
		// 		return responsemodels.GetChatResponse{}, domain.ErrUsersNotFound

		// 	case codes.Internal:
		// 		return responsemodels.GetChatResponse{},
		// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)

		// 	default:
		// 		return responsemodels.GetChatResponse{},
		// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
		// 	}
		// }
	}

	// 4. Map to response models and hydrate sender data
	var messageResponses []responsemodels.MessageResponse
	for _, msg := range messages {
		mResp := responsemodels.MessageResponse{
			MessageID: msg.MessageID,
			SenderID:  msg.SenderID,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
			Status:    msg.Status,
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

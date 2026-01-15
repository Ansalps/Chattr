package usecase

import (
	"log"
	"slices"

	"github.com/Ansalps/Chattr_Chat_Service/pkg/domain"
	interfacesrepository "github.com/Ansalps/Chattr_Chat_Service/pkg/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/usecase/interfacesUsecase"
)

type ChatUsecase struct {
	ChatRepository interfacesrepository.ChatRepository
}

func NewChatUsecase(repository interfacesrepository.ChatRepository) interfacesUsecase.ChatUsecase {
	return &ChatUsecase{
		ChatRepository: repository,
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
	creatorId, err := as.ChatRepository.CreatorID(req.GroupID)
	if err != nil {
		return responsemodels.AddMembersResponse{}, nil
	}
	if creatorId != req.UserID {
		return responsemodels.AddMembersResponse{}, domain.ErrNotCreatorId
	}
	resp1, err := as.ChatRepository.ExistingMembers(req.GroupID)
	if err != nil {
		return responsemodels.AddMembersResponse{}, err
	}
	req.GroupMembers = slices.DeleteFunc(req.GroupMembers, func(id uint64) bool {
		return slices.Contains(resp1, id)
	})
	resp2, err := as.ChatRepository.AddMembers(req)
	if err != nil {
		return responsemodels.AddMembersResponse{}, err
	}
	return resp2, nil
}

func (as *ChatUsecase) RemoveMember(req requestmodels.RemoveMemberRequest) (responsemodels.RemoveMemberResponse, error) {
	creatorId, err := as.ChatRepository.CreatorID(req.GroupID)
	if err != nil {
		return responsemodels.RemoveMemberResponse{}, nil
	}
	if creatorId != req.UserID {
		return responsemodels.RemoveMemberResponse{}, domain.ErrNotCreatorId
	}
	resp1, err := as.ChatRepository.ExistingMembers(req.GroupID)
	if err != nil {
		return responsemodels.RemoveMemberResponse{}, err
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

func (as *ChatUsecase)FetchMembersOfGroup(groupId string)([]uint64,error){
	userIds,err:=as.ChatRepository.FetchMembersOfGroup(groupId)
	if err!=nil{
		return nil,err
	}
	return userIds,nil
}

func (as *ChatUsecase)StoreIndividualChatInMessages(dm domain.Message)(error){
	err:=as.ChatRepository.StoreIndividualChatInMessages(dm)
	if err!=nil{
		log.Println(err)
		return err
	}
	return err
}
func (as *ChatUsecase)StoreOrUpdateIndividualChatInConversation(conversation domain.Conversation)(error){
	err:=as.ChatRepository.StoreOrUpdateIndividualChatInConversation(conversation)
	if err!=nil{
		log.Println(err)
		return err
	}
	return err
}

func (as *ChatUsecase)StoreGroupChatInMessages(gm domain.Message)error{
	err:=as.ChatRepository.StoreGroupChatInMessages(gm)
	if err!=nil{
		log.Println(err)
		return err
	}
	return err
}

func (as *ChatUsecase)StoreOrUpdateGroupChatInConversation(conversation domain.Conversation)(error){
	err:=as.ChatRepository.StoreOrUpdateGroupChatInConversation(conversation)
	if err!=nil{
		log.Println(err)
		return err
	}
	return err
}
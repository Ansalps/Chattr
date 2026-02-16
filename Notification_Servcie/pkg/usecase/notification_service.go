package usecase

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Ansalps/Chattr_Notification_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/websockethub"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/pb"
	interfacesrepository "github.com/Ansalps/Chattr_Notification_Service/pkg/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/usecase/interfacesUsecase"
)

type NotificationUsecase struct {
	NotificationRepository interfacesrepository.NotificationRepository
	Hub                    *websockethub.Hub
	AuthSubscriptionClient pb.AuthSubscriptionServiceClient
	logger domain.Logger
}

func NewNotificationUsecase(repository interfacesrepository.NotificationRepository, hub *websockethub.Hub, authSubClient pb.AuthSubscriptionServiceClient,logger domain.Logger) interfacesUsecase.NotificationUsecase {
	return &NotificationUsecase{
		NotificationRepository: repository,
		Hub:                    hub,
		AuthSubscriptionClient: authSubClient,
		logger: logger,
	}
}

func (uc *NotificationUsecase) ProcessLikeEvent(event requestmodels.PostEvent) {
	// 1. Get Actor Name from User Service (via gRPC or Client)
	actor, err := uc.AuthSubscriptionClient.UserPublicData(context.Background(), &pb.UserPublicDataRequest{
		UserId: event.ActorID,
	})
	if err != nil {
		log.Println("failed to fetch user name from auth_subscripiton service")
	}

	// 2. Format the message
	displayMsg := fmt.Sprintf("%s liked your post!", actor.Name)

	strPostID := strconv.FormatUint(event.PostID, 10)
	// 3. Save to DB (optional)
	err = uc.NotificationRepository.SaveNotification(domain.Notification{
		UserID:    event.PostOwnerID,
		ActorID:   event.ActorID,
		TargetID:  strPostID,
		Type:      event.Type,
		Message:   displayMsg,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Println("notification failed to store in db", err)
	}

	// 4. Push to Hub
	uc.Hub.Notification <- websockethub.NotificationMessage{
		UserID:  event.PostOwnerID,
		Payload: []byte(displayMsg),
	}
}

func (uc *NotificationUsecase) ProcessCommentEvent(event requestmodels.PostEvent) {
	// 1. Get Actor Name from User Service (via gRPC or Client)
	actor, err := uc.AuthSubscriptionClient.UserPublicData(context.Background(), &pb.UserPublicDataRequest{
		UserId: event.ActorID,
	})
	if err != nil {
		log.Println("failed to fetch user name from auth_subscripiton service")
	}

	// 2. Format the message
	displayMsg := fmt.Sprintf("%s commented on your post!", actor.Name)

	strPostID := strconv.FormatUint(event.PostID, 10)

	// 3. Save to DB (optional)
	err = uc.NotificationRepository.SaveNotification(domain.Notification{
		UserID:    event.PostOwnerID,
		ActorID:   event.ActorID,
		TargetID:  strPostID,
		Type:      event.Type,
		Message:   displayMsg,
		CreatedAt: time.Now(),
	})

	if err != nil {
		log.Println("notification failed to store in db", err)
	}

	// 4. Push to Hub
	uc.Hub.Notification <- websockethub.NotificationMessage{
		UserID:  event.PostOwnerID,
		Payload: []byte(displayMsg),
	}
}

func (uc *NotificationUsecase) ProcessFollowEvent(event requestmodels.UserEvent) {
	// 1. Get Actor Name from User Service (via gRPC or Client)
	actor, err := uc.AuthSubscriptionClient.UserPublicData(context.Background(), &pb.UserPublicDataRequest{
		UserId: event.ActorID,
	})
	if err != nil {
		log.Println("failed to fetch user name from auth_subscripiton service")
	}
	// 2. Format the message
	displayMsg := fmt.Sprintf("%s started following you!", actor.Name)

	strActorID := strconv.FormatUint(event.ActorID, 10)

	// 3. Save to DB (optional)
	err = uc.NotificationRepository.SaveNotification(domain.Notification{
		UserID:    event.FollowingID,
		ActorID:   event.ActorID,
		TargetID:  strActorID,
		Type:      event.Type,
		Message:   displayMsg,
		CreatedAt: time.Now(),
	})

	if err != nil {
		log.Println("notification failed to store in db", err)
	}

	// 4. Push to Hub
	uc.Hub.Notification <- websockethub.NotificationMessage{
		UserID:  event.FollowingID,
		Payload: []byte(displayMsg),
	}

}

func (uc *NotificationUsecase) ProcessDirectMessageEvent(event requestmodels.DirectMessageEvent) {
	// 1. Get Actor Name from User Service (via gRPC or Client)
	actor, err := uc.AuthSubscriptionClient.UserPublicData(context.Background(), &pb.UserPublicDataRequest{
		UserId: event.ActorID,
	})
	if err != nil {
		log.Println("failed to fetch user name from auth_subscripiton service")
	}
	// 2. Format the message
	displayMsg := fmt.Sprintf("%s messaged you!", actor.Name)

	// 3. Save to DB (optional)
	err = uc.NotificationRepository.SaveNotification(domain.Notification{
		UserID:    event.RecipientID,
		ActorID:   event.ActorID,
		TargetID:  event.ConversationID,
		Type:      event.Type,
		Message:   displayMsg,
		CreatedAt: time.Now(),
	})

	if err != nil {
		log.Println("notification failed to store in db", err)
	}

	// 4. Push to Hub
	uc.Hub.Notification <- websockethub.NotificationMessage{
		UserID:  event.RecipientID,
		Payload: []byte(displayMsg),
	}

}

func (uc *NotificationUsecase) ProcessGroupMessageEvent(event requestmodels.GroupMessageEvent) {
	// 1. Get Actor Name from User Service (via gRPC or Client)
	actor, err := uc.AuthSubscriptionClient.UserPublicData(context.Background(), &pb.UserPublicDataRequest{
		UserId: event.ActorID,
	})
	if err != nil {
		log.Println("failed to fetch user name from auth_subscripiton service")
	}

	// 2. Format the message
	displayMsg := fmt.Sprintf("%s messaged in the group \"%s\"!", actor.Name,event.GroupName)

	for _,v:=range event.RecipientID{
		// 3. Save to DB (optional)
		err = uc.NotificationRepository.SaveNotification(domain.Notification{
			UserID:    v,
			ActorID:   event.ActorID,
			TargetID:  event.ConversationID,
			Type:      event.Type,
			Message:   displayMsg,
			CreatedAt: time.Now(),
		})
		if err != nil {
			log.Println("notification failed to store in db", err)
		}
		// 4. Push to Hub
		uc.Hub.Notification <- websockethub.NotificationMessage{
			UserID:  v,
			Payload: []byte(displayMsg),
		}
	}
}

func (uc *NotificationUsecase)GetAllNotifications(req requestmodels.GetNotificationsequest)([]domain.Notification,error){
	resp,err:=uc.NotificationRepository.GetAllNotifications(req)
	if err!=nil{
		return nil,err
	}
	return resp,nil
}

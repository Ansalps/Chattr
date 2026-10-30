package usecase

import (
	"context"
	"encoding/json"
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
}

func (uc *NotificationUsecase) pushNotification(notification domain.Notification) {
	payload, err := json.Marshal(notification)
	if err != nil {
		return
	}
	uc.Hub.Notification <- websockethub.NotificationMessage{UserID: notification.UserID, Payload: payload}
}

func NewNotificationUsecase(repository interfacesrepository.NotificationRepository, hub *websockethub.Hub, authSubClient pb.AuthSubscriptionServiceClient) interfacesUsecase.NotificationUsecase {
	return &NotificationUsecase{
		NotificationRepository: repository,
		Hub:                    hub,
		AuthSubscriptionClient: authSubClient,
	}
}

func (uc *NotificationUsecase) ProcessLikeEvent(event requestmodels.PostEvent) error {
	fmt.Println("is it reaching inside process like event")
	if event.ActorID == event.PostOwnerID {
		fmt.Printf("action id: %d post owner id: %d",event.ActorID,event.PostOwnerID)
		return nil
	}
	// 1. Get Actor Name from User Service (via gRPC or Client)
	actor, err := uc.AuthSubscriptionClient.UserPublicData(context.Background(), &pb.UserPublicDataRequest{
		UserId: event.ActorID,
	})
	if err != nil {
		log.Println("failed to fetch user name from auth_subscripiton service", err)
		return err
	}

	// 2. Format the message
	displayMsg := fmt.Sprintf("%s liked your post!", actor.Name)

	fmt.Println("disply message: ",displayMsg)

	strPostID := strconv.FormatUint(event.PostID, 10)
	// 3. Save to DB (optional)
	notification := domain.Notification{
		UserID:    event.PostOwnerID,
		ActorID:   event.ActorID,
		TargetID:  strPostID,
		Type:      event.Type,
		Message:   displayMsg,
		CreatedAt: time.Now(),
	}
	err = uc.NotificationRepository.SaveNotification(notification)
	if err != nil {
		log.Println("notification failed to store in db", err)
		return err
	}
	fmt.Println("domain.notification---------", notification)
	// 4. Push to Hub
	uc.pushNotification(notification)
	return nil
}

func (uc *NotificationUsecase) ProcessCommentEvent(event requestmodels.PostEvent) error {
	if event.ActorID == event.PostOwnerID {
		return nil
	}
	// 1. Get Actor Name from User Service (via gRPC or Client)
	actor, err := uc.AuthSubscriptionClient.UserPublicData(context.Background(), &pb.UserPublicDataRequest{
		UserId: event.ActorID,
	})
	if err != nil {
		//log.Println("failed to fetch user name from auth_subscripiton service",err)
		return err
	}

	// 2. Format the message
	displayMsg := fmt.Sprintf("%s commented on your post!", actor.Name)

	strPostID := strconv.FormatUint(event.PostID, 10)

	// 3. Save to DB (optional)
	notification := domain.Notification{
		UserID:    event.PostOwnerID,
		ActorID:   event.ActorID,
		TargetID:  strPostID,
		Type:      event.Type,
		Message:   displayMsg,
		CreatedAt: time.Now(),
	}
	err = uc.NotificationRepository.SaveNotification(notification)

	if err != nil {
		//log.Println("notification failed to store in db", err)
		return err
	}

	// 4. Push to Hub
	uc.pushNotification(notification)
	return nil
}

func (uc *NotificationUsecase) ProcessFollowEvent(event requestmodels.UserEvent) error {
	if event.ActorID == event.FollowingID {
		return nil
	}
	// 1. Get Actor Name from User Service (via gRPC or Client)
	actor, err := uc.AuthSubscriptionClient.UserPublicData(context.Background(), &pb.UserPublicDataRequest{
		UserId: event.ActorID,
	})
	if err != nil {
		//log.Println("failed to fetch user name from auth_subscripiton service",err)
		return err
	}
	// 2. Format the message
	displayMsg := fmt.Sprintf("%s started following you!", actor.Name)

	strActorID := strconv.FormatUint(event.ActorID, 10)

	// 3. Save to DB (optional)
	notification := domain.Notification{
		UserID:    event.FollowingID,
		ActorID:   event.ActorID,
		TargetID:  strActorID,
		Type:      event.Type,
		Message:   displayMsg,
		CreatedAt: time.Now(),
	}
	err = uc.NotificationRepository.SaveNotification(notification)

	if err != nil {
		//log.Println("notification failed to store in db", err)
		return err
	}

	// 4. Push to Hub
	uc.pushNotification(notification)
	return nil

}

func (uc *NotificationUsecase) ProcessDirectMessageEvent(event requestmodels.DirectMessageEvent) error {
	if event.ActorID == event.RecipientID {
		return nil
	}
	// 1. Get Actor Name from User Service (via gRPC or Client)
	actor, err := uc.AuthSubscriptionClient.UserPublicData(context.Background(), &pb.UserPublicDataRequest{
		UserId: event.ActorID,
	})
	if err != nil {
		//log.Println("failed to fetch user name from auth_subscripiton service")
		return err
	}
	// 2. Format the message
	displayMsg := fmt.Sprintf("%s messaged you!", actor.Name)

	// 3. Save to DB (optional)
	notification := domain.Notification{
		UserID:    event.RecipientID,
		ActorID:   event.ActorID,
		TargetID:  event.ConversationID,
		Type:      event.Type,
		Message:   displayMsg,
		CreatedAt: time.Now(),
	}
	err = uc.NotificationRepository.SaveNotification(notification)

	if err != nil {
		//log.Println("notification failed to store in db", err)
		return err
	}

	// 4. Push to Hub
	uc.pushNotification(notification)
	return nil
}

func (uc *NotificationUsecase) ProcessGroupMessageEvent(event requestmodels.GroupMessageEvent) error {
	// 1. Get Actor Name from User Service (via gRPC or Client)
	actor, err := uc.AuthSubscriptionClient.UserPublicData(context.Background(), &pb.UserPublicDataRequest{
		UserId: event.ActorID,
	})
	if err != nil {
		//log.Println("failed to fetch user name from auth_subscripiton service")
		return err
	}

	// 2. Format the message
	displayMsg := fmt.Sprintf("%s messaged in the group \"%s\"!", actor.Name, event.GroupName)

	for _, v := range event.RecipientID {
		if v == event.ActorID {
			continue
		}
		// 3. Save to DB (optional)
		notification := domain.Notification{
			UserID:    v,
			ActorID:   event.ActorID,
			TargetID:  event.ConversationID,
			Type:      event.Type,
			Message:   displayMsg,
			CreatedAt: time.Now(),
		}
		err = uc.NotificationRepository.SaveNotification(notification)
		if err != nil {
			//log.Println("notification failed to store in db", err)
			return err
		}
		// 4. Push to Hub
		uc.pushNotification(notification)
	}
	return nil
}

func (uc *NotificationUsecase) GetAllNotifications(req requestmodels.GetNotificationsequest) ([]domain.Notification, error) {
	resp, err := uc.NotificationRepository.GetAllNotifications(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

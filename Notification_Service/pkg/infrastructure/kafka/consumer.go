package kafka

import (
	"context"
	"encoding/json"
	"strings"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/logger"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/websockethub"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/usecase/interfacesUsecase"
	"github.com/segmentio/kafka-go"
)

// PostEvent must match the JSON tags from your Post Service output
// type PostEvent struct {
// 	Type        string `json:"type"`
// 	ActorID     uint64 `json:"actorId"`
// 	PostID      uint64 `json:"postId"`
// 	PostOwnerID uint64 `json:"postOwnerId"`
// }

func StartNotificationConsumer(brokerStr string, topic string, hub *websockethub.Hub, notificationUsecase interfacesUsecase.NotificationUsecase,log logger.Logger) {
	// 1. Setup the Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(brokerStr, ","),
		Topic:    topic,
		GroupID:  "notification-group", // This ensures we don't miss messages
		MinBytes: 10e3,                 // 10KB
		MaxBytes: 10e6,                 // 10MB
	})

	log.Info("Listening for Kafka messages on topic: ",
		logger.Field{Key: "topic",Value: "topic"})

	for {
		// 2. Block until a message arrives
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			//log.Printf("Kafka Read Error: %v", err)
			log.Error("Kafka Read Error: ",
			logger.Field{Key: "error",Value: err})
			continue
		}

		// 1. Determine the type first
		var base struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(m.Value, &base); err != nil {
			//log.Printf("Failed to parse event type: %v", err)
			log.Error("Failed to parse event type: ",
				logger.Field{Key: "error",Value: err})
			continue
		}

		

		switch base.Type {
		case "USER_FOLLOW":
			var event requestmodels.UserEvent
			if err := json.Unmarshal(m.Value, &event); err != nil {
				log.Error("Failed to parse event type: ",
				logger.Field{Key: "error",Value: err})
				continue
			}
			if event.Type=="USER_FOLLOW"{
				//log.Printf("New Follow detected for user %d",event.FollowingID)
				log.Info("New Follow detected for user: ",
					logger.Field{Key: "following_id",Value: event.FollowingID})
				err=notificationUsecase.ProcessFollowEvent(event)
				if err!=nil{
					log.Error("Failed to store  notification for follow event",
						logger.Field{Key: "error",Value: err})
				}
			}

		case "POST_LIKE", "POST_COMMENT":
			var event requestmodels.PostEvent
			if err := json.Unmarshal(m.Value, &event); err != nil {
				//log.Printf("Failed to parse PostEvent: %v", err)
				log.Error("Failed to parse PostEvent: ",
					logger.Field{Key: "error",Value: err})
				continue
			}
			if event.Type == "POST_LIKE" && event.PostOwnerID != 0 {
				//log.Printf("New Like detected for User %d", event.PostOwnerID)
				log.Info("New Like detected for User",
					logger.Field{Key: "error",Value: err})
				err=notificationUsecase.ProcessLikeEvent(event)
				if err!=nil{
					log.Error("failed to store notification for like event",
						logger.Field{Key: "error",Value: err})	
				}
			} else if event.Type=="POST_COMMENT" &&event.PostOwnerID!=0{
				//log.Printf("New Comment detected for User %d",event.PostOwnerID)
				log.Info("New Comment detected for User",
					logger.Field{Key: "error",Value: err})
				err=notificationUsecase.ProcessCommentEvent(event)
				if err!=nil{
					log.Error("failed to store notification for post comment",
					logger.Field{Key: "error",Value: err})
				}
			}
		case "DIRECT_MESSAGE":
			var event requestmodels.DirectMessageEvent
			if err := json.Unmarshal(m.Value, &event); err != nil {
				//log.Printf("Failed to parse PostEvent: %v", err)
				log.Error("Failed to parse PostEvent: ",
					logger.Field{Key: "error",Value: err})
				continue
			}
			if event.Type=="DIRECT_MESSAGE"{
				//log.Printf("New Message detected for user %d",event.RecipientID)
				log.Info("New Message detected for user",
					logger.Field{Key: "recipient_id",Value: event.RecipientID})
				err=notificationUsecase.ProcessDirectMessageEvent(event)
				if err!=nil{
					log.Error("faile to store notification for direct message",
						logger.Field{Key: "error",Value: err})
				}
			}
		case "GROUP_MESSAGE":
			var event requestmodels.GroupMessageEvent
			if err := json.Unmarshal(m.Value, &event); err != nil {
				//log.Printf("Failed to parse PostEvent: %v", err)
				log.Error("Failed to parse PostEvent: ",
					logger.Field{Key: "error",Value: err})
				continue
			}
			if event.Type=="GROUP_MESSAGE"{
				//log.Print("Group Message detectd for users",event.RecipientID)
				log.Info("Group Message detectd for users",
					logger.Field{Key: "recipient_id",Value: event.RecipientID})
				err=notificationUsecase.ProcessGroupMessageEvent(event)
				if err!=nil{
					log.Error("failed to store notification for group message",
						logger.Field{Key: "error",Value: err})
				}
			}
		default:
			//log.Printf("Unknown event type: %s", base.Type)
			log.Error("Unknown event type",
				logger.Field{Key: "error",Value: err})
		}
	}
}

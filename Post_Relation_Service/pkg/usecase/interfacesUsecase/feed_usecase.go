package interfacesUsecase

import "github.com/Ansalps/Chattr_Post_Relation_Service/pkg/events"

type FeedUsecase interface {
	ProcessPostCreated(event events.PostCreatedEvent) error
}

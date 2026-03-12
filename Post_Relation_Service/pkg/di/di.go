package di

import (
	"github.com/Ansalps/Chattr_Post_Relation_Service/infrastructure/logger"
	"github.com/Ansalps/Chattr_Post_Relation_Service/infrastructure/kafka"
	services "github.com/Ansalps/Chattr_Post_Relation_Service/pkg/api"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/client"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/db"
	repository "github.com/Ansalps/Chattr_Post_Relation_Service/pkg/respository"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase"
)

func DependencyIndjection(cfg *config.Config,log logger.Logger) (*services.PostRelationServer, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}
	authSubscriptionClient, err := client.InitAuthSubscriptionServiceClient(cfg)
	if err != nil {
		return nil, err
	}
	redisClient := client.NewRedisClient(cfg)
	RedisRepository := repository.NewRedisRepository(redisClient)

	// This reads much better: Config -> Kafka -> Brokers
	kafkaProducer := kafka.NewKafkaProducer([]string{cfg.Kafka.Brokers})

	PostRelationRepository := repository.NewPostRelationRepository(gormDB)
	PostRelationUsecase := usecase.NewPostRelationUsecase(PostRelationRepository, authSubscriptionClient, RedisRepository,kafkaProducer,log)
	PostRelationServiceServer := services.NewPostRelationSever(PostRelationUsecase)

	return PostRelationServiceServer, nil
}

package di

import (
	"fmt"

	"github.com/Ansalps/Chattr_Post_Relation_Service/infrastructure/kafka"
	"github.com/Ansalps/Chattr_Post_Relation_Service/infrastructure/logger"
	services "github.com/Ansalps/Chattr_Post_Relation_Service/pkg/api"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/client"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/db"
	repository "github.com/Ansalps/Chattr_Post_Relation_Service/pkg/respository"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase/interfacesUsecase"
)

func DependencyInjection(cfg *config.Config, log logger.Logger) (*services.PostRelationServer, interfacesUsecase.FeedUsecase, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		fmt.Println("error connecting database", err)
		return nil, nil, err
	}
	
	authSubscriptionClient, err := client.InitAuthSubscriptionServiceClient(cfg)
	if err != nil {
		fmt.Println("error connecting auth service", err)
		return nil, nil, err
	}
	
	redisClient := client.NewRedisClient(cfg)
	RedisRepository := repository.NewRedisRepository(redisClient)
	
	// This reads much better: Config -> Kafka -> Brokers
	// kafkaProducer := kafka.NewKafkaProducer([]string{cfg.Kafka.Brokers})
	
	kafkaProducer := kafka.NewKafkaProducer(
		[]string{cfg.Kafka.Brokers},
		[]byte(cfg.Kafka.CACert),
		[]byte(cfg.Kafka.AccessCert),
		[]byte(cfg.Kafka.AccessKey),
	)
	
	PostRelationRepository := repository.NewPostRelationRepository(gormDB)
	fmt.Println("444444")
	PostRelationUsecase := usecase.NewPostRelationUsecase(PostRelationRepository, authSubscriptionClient,
		RedisRepository, kafkaProducer, log, cfg)
		
	// ✅ NEW: Feed Usecase
	feedUsecase := usecase.NewFeedUsecase(
		PostRelationRepository,
		RedisRepository,
		cfg,
		log,
	)
	
	PostRelationServiceServer := services.NewPostRelationSever(PostRelationUsecase)
	
	return PostRelationServiceServer, feedUsecase, nil
}

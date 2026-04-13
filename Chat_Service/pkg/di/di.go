package di

import (
	"fmt"

	"github.com/Ansalps/Chattr_Chat_Service/logger"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/client"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/config"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/db"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/handler"
	awss3 "github.com/Ansalps/Chattr_Chat_Service/pkg/infrastructure/AwsS3"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/infrastructure/kafka"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/repository"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/routes"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/usecase"
	"github.com/gin-gonic/gin"
)

func DependencyInjection(router *gin.Engine, cfg *config.Config,log logger.Logger) error {
	mongoClient, err := db.ConnectMongo(cfg)
	if err != nil {
		return err
	}
	authClient, err := client.InitAuthSubscriptionServiceClient(cfg)
	if err != nil {
		return err
	}
	// This reads much better: Config -> Kafka -> Brokers
	// kafkaProducer := kafka.NewKafkaProducer([]string{cfg.Kafka.Brokers})
	kafkaProducer := kafka.NewKafkaProducer(
		[]string{cfg.Kafka.Brokers},
		[]byte(cfg.Kafka.CACert),
		[]byte(cfg.Kafka.AccessCert),
		[]byte(cfg.Kafka.AccessKey),
	)
	
	AwsS3Client, err := awss3.NewS3Client(cfg.Aws.AwsAccessKey, cfg.Aws.AwsSecretAccessKey, cfg.Aws.AwsRegion)
	if err != nil {
		return  fmt.Errorf("failed to initialize s3 client: %w", err)
	}
	ChatRepository := repository.NewChatRepository(mongoClient.Client())
	ChatUsecase := usecase.NewChatUsecase(ChatRepository, authClient, AwsS3Client, cfg.Aws.AwsBucket,cfg)
	ChatHandler := handler.NewChatHandler(ChatUsecase, kafkaProducer, cfg,log)
	routes.ChatServiceRoutes(router, ChatHandler)
	return nil
}

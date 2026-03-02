package di

import (
	"fmt"

	services "github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/api"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/db"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/infrastructure/jwt"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/infrastructure/razorpaygateway"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/infrastructure/smtp"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/repository"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/usecase"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils/AwsS3"
)

func DependencyIndjection(cfg *config.Config) (*services.AuthSubscriptionServer, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}
	AuthSubscriptionRepository := repository.NewAuthSubscriptionRepository(gormDB)
	SmtpProvider := smtp.NewSmtpProvider(&cfg.Smtp)
	JwtProvider := jwt.NewJwtProvider()

	AwsS3Client, err := AwsS3.NewS3Client(cfg.Aws.AwsAccessKey, cfg.Aws.AwsSecretAccessKey, cfg.Aws.AwsRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize s3 client: %w", err)
	}
	//razorpayClient := utils.NewRazorpayClient(cfg.Razorpay.KeyId, cfg.Razorpay.KeySecret)
	razorpayGateway := razorpaygateway.NewRazorpayGateway(cfg.Razorpay.KeyId, cfg.Razorpay.KeySecret)
	AuthSubscriptionUsecase := usecase.NewAuthSubscriptionUsecase(AuthSubscriptionRepository,
		SmtpProvider, cfg, JwtProvider, razorpayGateway, AwsS3Client, cfg.Aws.AwsBucket)
	AuthSubscriptionServiceServer := services.NewAuthSubscriptionServer(AuthSubscriptionUsecase)

	return AuthSubscriptionServiceServer, nil
}

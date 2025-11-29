package di

import (
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/api"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/db"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/repository"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/usecase"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils/jwt"
	randomnumber "github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils/randomNumber"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils/smtp"
)

func DependencyIndjection(cfg *config.Config) (*services.AuthSubscriptionServer, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}
	AuthSubscriptionRepository := repository.NewAuthSubscriptionRepository(gormDB)
	SmtpUtil:=smtp.NewSmtpUtil(&cfg.Smtp)
	JwtUtil:=jwt.NewJwtUtil()
	RandomUtil:=randomnumber.NewRandomNumberUtil()
	razorpayClient:=utils.NewRazorpayClient(cfg.Razorpay.KeyId,cfg.Razorpay.KeySecret)
	AuthSubscriptionUsecase := usecase.NewAuthSubscriptionUsecase(AuthSubscriptionRepository,RandomUtil,SmtpUtil,&cfg.Token,JwtUtil,/*&cfg.Razorpay,*/razorpayClient)
	AuthSubscriptionServiceServer := services.NewAuthSubscriptionServer(AuthSubscriptionUsecase)

	
	return AuthSubscriptionServiceServer, nil
}

package di

import (
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/api"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/config"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/db"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/repository"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/usecase"
)

func DependencyIndjection(cfg *config.Config) (*services.AuthSubscriptionServer, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}
	AuthSubscriptionRepository := repository.NewAuthSubscriptionRepository(gormDB)
	AuthSubscriptionUsecase := usecase.NewAuthSubscriptionUsecase(AuthSubscriptionRepository)
	AuthSubscriptionServiceServer := services.NewAuthSubscriptionServer(AuthSubscriptionUsecase)
	return AuthSubscriptionServiceServer, nil
}

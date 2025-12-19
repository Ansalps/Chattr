package di

import (
	services "github.com/Ansalps/Chattr_Post_Relation_Service/pkg/api"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/client"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/db"
	repository "github.com/Ansalps/Chattr_Post_Relation_Service/pkg/respository"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase"
)

func DependencyIndjection(cfg *config.Config) (*services.PostRelationServer, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}
	authSubscriptionClient,err:=client.InitAuthSubscriptionServiceClient(cfg)
	if err!=nil{
		return nil,err
	}

	PostRelationRepository := repository.NewPostRelationRepository(gormDB)
	PostRelationUsecase := usecase.NewPostRelationUsecase(PostRelationRepository,authSubscriptionClient)
	PostRelationServiceServer := services.NewPostRelationSever(PostRelationUsecase)

	return PostRelationServiceServer, nil
}

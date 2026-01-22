package di

import (
	"github.com/Ansalps/Chattr_Chat_Service/pkg/client"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/config"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/db"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/handler"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/repository"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/routes"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/usecase"
	"github.com/gin-gonic/gin"
)

func DependencyInjection(router *gin.Engine, cfg *config.Config) ( error) {
	mongoClient, err := db.ConnectMongo()
	if err != nil {
		return err
	}
	authClient,err:=client.InitAuthSubscriptionServiceClient(cfg)
	if err!=nil{
		return err
	}
	ChatRepository := repository.NewChatRepository(mongoClient.Client())
	ChatUsecase := usecase.NewChatUsecase(ChatRepository,authClient)
	ChatHandler := handler.NewChatHandler(ChatUsecase)
	routes.ChatServiceRoutes(router,ChatHandler)
	return  nil
}

package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	mongoClient *mongo.Client
}

func (m *MongoClient) Client() *mongo.Client {
	return m.mongoClient
}

func ConnectMongo() (*MongoClient,error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	//ping to check connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	log.Println("conncect to MongoDB")
	return &MongoClient{
		mongoClient: client,
	},nil
}

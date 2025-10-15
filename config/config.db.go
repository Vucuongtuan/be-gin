package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
var (
    URI      string
    DATABASE string = "chat_app" 
)
var MongoClient *mongo.Client
func ConnectionDB() (*mongo.Client,error){
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}
	clientUri := options.Client().ApplyURI(databaseURL)
	client,err := mongo.NewClient(clientUri)

	if err != nil {
return nil,fmt.Errorf("Can't connect to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(),10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil,fmt.Errorf("Can't connect to database: , %v", err)
	}
	err = client.Ping(ctx,nil)

	if err != nil {
		return nil,fmt.Errorf("Can't connect to database: , %v", err)
	}

	log.Println("Connected to MongoDB database successfully")
	MongoClient = client
	return client, nil
}

func GetCollection(collectionName string) *mongo.Collection {
	return MongoClient.Database(DATABASE).Collection(collectionName)
}
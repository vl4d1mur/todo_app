package mongo

import (
	"context"
	"log"
	"time"

	"todo/internal/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

func ConnectMongo() {
	client, err := mongo.Connect(options.Client().ApplyURI(config.MongoUri))
	if err != nil {
		log.Fatal("Connection error:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Ping error:", err)
	}

	MongoClient = client
	MongoDB = client.Database(config.MongoDbName)

	log.Println("Connected to:", config.MongoDbName)
}

func CloseMongoDB() {
	if MongoClient != nil {
		if err := MongoClient.Disconnect(context.Background()); err != nil {
			log.Println("Disconnect error:", err)
		}
		log.Println("Connection closed")
	}
}

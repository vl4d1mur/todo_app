package mongo

import (
	"context"
	"time"

	"task_service/internal/config"
	"task_service/pkg/log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

func ConnectMongo() {
	client, err := mongo.Connect(options.Client().ApplyURI(config.MongoUri))
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("Connection error:")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Logger.Fatal().Err(err).Msg("Ping error:")
	}

	MongoClient = client
	MongoDB = client.Database(config.MongoDbName)

	log.Logger.Info().Str("Connected to:", config.MongoDbName)
}

func CloseMongoDB() {
	if MongoClient != nil {
		if err := MongoClient.Disconnect(context.Background()); err != nil {
			log.Logger.Error().Err(err).Msg("Disconnect error:")
		}
		log.Logger.Info().Msg("Connection closed")
	}
}

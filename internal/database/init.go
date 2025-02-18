package mongodb

import (
	"context"
	"erply_test/internal/logger"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConnectionConfig struct {
	Host   string
	Port   string
	DbName string
	User   string
	Pass   string
}

func ConnectDB(c ConnectionConfig, logger logger.LoggerInterface) *mongo.Database {
	var mongoURI = "mongodb://" + c.Host + ":" + c.Port

	clientOptions := options.Client().ApplyURI(mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Could not ping MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB!")

	return client.Database(c.DbName)
}

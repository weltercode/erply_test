package mongodb

import (
	"context"
	"erply_test/internal/logger"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConnectionConfig struct {
	URI    string
	DbName string
	User   string
	Pass   string
}

func ConnectDB(c ConnectionConfig, logger logger.LoggerInterface) *mongo.Database {
	clientOptions := options.Client().ApplyURI(c.URI)

	logger.Info(fmt.Sprintf("Used URI: %s", c.URI))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error("Error connecting to MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error("Could not ping MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB!")

	return client.Database(c.DbName)
}

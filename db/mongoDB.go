package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func Close(ctx context.Context, client *mongo.Client, cancel context.CancelFunc) {
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func Connect(uri string) (*MongoDB, context.Context, context.CancelFunc) {
	const timeOut = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	database := client.Database("project-consumer")

	return &MongoDB{
		Client:   client,
		Database: database,
	}, ctx, cancel
}

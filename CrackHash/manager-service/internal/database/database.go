package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnection struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewMongoConnection(ctx context.Context, url string, dbName string) (*MongoConnection, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(dbName)

	return &MongoConnection{
		Client: client,
		DB:     db,
	}, nil
}

func (m *MongoConnection) Disconnect(ctx context.Context) {
	if m.Client != nil {
		m.Client.Disconnect(ctx)
	}
}

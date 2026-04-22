package main

import (
	"context"
	"crackhash/pkg/models"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrStorageFull = errors.New("Not enough memory at TasksStorage")
)

type RequestState struct {
	ReqID           string               `bson:"_id"`
	Hash            string               `bson:"hash"`
	MaxLength       int                  `bson:"maxLength"`
	Status          models.RequestStatus `bson:"status"`
	Data            []string             `bson:"data"`
	WorkersFinished int                  `bson:"workersFinished"`
	TotalWorkers    int                  `bson:"totalWorkers"`
	CreatedAt       time.Time            `bson:"createdAt"`
}

type MongoStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoStorage(ctx context.Context, url string, dbName string) (*MongoStorage, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection("tasks")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "hash", Value: 1}, {Key: "maxLength", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, _ = collection.Indexes().CreateOne(ctx, indexModel)

	return &MongoStorage{
		client:     client,
		collection: collection,
	}, nil
}

func (s *MongoStorage) Disconnect(ctx context.Context) {
	if s.client != nil {
		s.client.Disconnect(ctx)
	}
}

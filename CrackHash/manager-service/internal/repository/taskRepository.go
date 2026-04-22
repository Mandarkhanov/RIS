package repository

import (
	"context"
	"crackhash/pkg/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RequestState struct {
	ReqID           string               `bson:"_id"`
	Hash            string               `bson:"hash"`
	MaxLength       int                  `bson:"maxLength"`
	Status          domain.RequestStatus `bson:"status"`
	Data            []string             `bson:"data"`
	WorkersFinished int                  `bson:"workersFinished"`
	TotalWorkers    int                  `bson:"totalWorkers"`
	CreatedAt       time.Time            `bson:"createdAt"`
}

type TaskRepository struct {
	collection *mongo.Collection
}

func NewTaskRepository(ctx context.Context, db *mongo.Database) (*TaskRepository, error) {
	collection := db.Collection("tasks")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "hash", Value: 1}, {Key: "maxLength", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, err
	}

	return &TaskRepository{
		collection: collection,
	}, nil
}

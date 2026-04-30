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
	ReqID         string               `bson:"_id"`
	Hash          string               `bson:"hash"`
	MaxLength     int                  `bson:"maxLength"`
	Status        domain.RequestStatus `bson:"status"`
	Data          []string             `bson:"data"`
	FinishedParts []int                `bson:"finishedParts"`
	TotalWorkers  int                  `bson:"totalWorkers"`
	PendingParts  []int                `bson:"pendingParts"`
	CreatedAt     time.Time            `bson:"createdAt"`
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

func (r *TaskRepository) StoreOrGetExists(ctx context.Context, state *RequestState) (string, error) {
	_, err := r.collection.InsertOne(ctx, state)
	if mongo.IsDuplicateKeyError(err) {
		var existing RequestState
		errFind := r.collection.FindOne(ctx, bson.M{"hash": state.Hash, "maxLength": state.MaxLength}).Decode(&existing)
		if errFind != nil {
			return "", errFind
		}
		return existing.ReqID, nil
	}
	if err != nil {
		return "", err
	}
	return state.ReqID, nil
}

func (r *TaskRepository) Get(ctx context.Context, reqID string) (*RequestState, bool) {
	var state RequestState
	err := r.collection.FindOne(ctx, bson.M{"_id": reqID}).Decode(&state)
	if err != nil {
		return nil, false
	}
	return &state, true
}

func (r *TaskRepository) AddWorkerResult(ctx context.Context, reqID string, partNumber int, words []string) (*RequestState, error) {
	addToSet := bson.M{"finishedParts": partNumber}
	if len(words) > 0 {
		addToSet["data"] = bson.M{"$each": words}
	}

	update := bson.M{"$addToSet": addToSet}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedState RequestState
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": reqID}, update, opts).Decode(&updatedState)
	if err != nil {
		return nil, err
	}

	if len(updatedState.FinishedParts) == updatedState.TotalWorkers {
		_, _ = r.collection.UpdateByID(ctx, reqID, bson.M{"$set": bson.M{"status": domain.StatusReady}})
		updatedState.Status = domain.StatusReady
	}
	return &updatedState, nil
}

func (r *TaskRepository) GetAndMarkTimedOutTasks(ctx context.Context, timeout time.Duration) ([]string, error) {
	deadline := time.Now().Add(-timeout)
	filter := bson.M{"status": domain.StatusInProgress, "createdAt": bson.M{"$lt": deadline}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var timedOutTasks []RequestState
	if err := cursor.All(ctx, &timedOutTasks); err != nil {
		return nil, err
	}

	var cancelledIDs []string
	for _, task := range timedOutTasks {
		newStatus := domain.StatusError
		if len(task.Data) > 0 {
			newStatus = domain.StatusPartialReady
		}
		_, _ = r.collection.UpdateByID(ctx, task.ReqID, bson.M{"$set": bson.M{"status": newStatus}})
		cancelledIDs = append(cancelledIDs, task.ReqID)
	}
	return cancelledIDs, nil
}

func (r *TaskRepository) MarkPartDispatched(ctx context.Context, reqID string, partNumber int) error {
	update := bson.M{"$pull": bson.M{"pendingParts": partNumber}}
	_, err := r.collection.UpdateByID(ctx, reqID, update)
	return err
}

func (r *TaskRepository) GetTasksWithPendingParts(ctx context.Context) ([]RequestState, error) {
	filter := bson.M{"pendingParts.0": bson.M{"$exists": true}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []RequestState
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

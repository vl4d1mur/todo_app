package repository

import (
	"context"
	"time"

	"notifier_service/internal/config"
	"notifier_service/internal/db/mongo"
	"notifier_service/internal/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var _ DeadlineRepository = (*DeadlineRepositoryImpl)(nil)

type DeadlineRepositoryImpl struct{}

func NewDeadlineRepository() *DeadlineRepositoryImpl {
	return &DeadlineRepositoryImpl{}
}

func (r *DeadlineRepositoryImpl) collection() *mongodriver.Collection {
	return mongo.MongoDB.Collection(config.MongoDeadlineCollection)
}

func (r *DeadlineRepositoryImpl) Upsert(ctx context.Context, d *models.TaskDeadline) error {
	now := time.Now()
	d.UpdatedAt = now

	var existing models.TaskDeadline
	err := r.collection().FindOne(ctx, bson.M{"task_id": d.TaskID}).Decode(&existing)

	deadlineChanged := err == mongodriver.ErrNoDocuments || !existing.Deadline.Equal(d.Deadline)

	setFields := bson.M{
		"user_id":    d.UserID,
		"title":      d.Title,
		"deadline":   d.Deadline,
		"updated_at": d.UpdatedAt,
	}

	if deadlineChanged {
		setFields["notified"] = false
	}

	update := bson.M{
		"$set": setFields,
		"$setOnInsert": bson.M{
			"task_id":    d.TaskID,
			"created_at": d.CreatedAt,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err = r.collection().UpdateOne(ctx, bson.M{"task_id": d.TaskID}, update, opts)
	return err
}

func (r *DeadlineRepositoryImpl) GetPending(ctx context.Context) ([]models.TaskDeadline, error) {
	now := time.Now()
	threshold := now.Add(24 * time.Hour)

	filter := bson.M{
		"deadline": bson.M{
			"$gte": now,
			"$lte": threshold,
		},
		"notified": false,
	}

	cursor, err := r.collection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	deadlines := []models.TaskDeadline{}
	if err := cursor.All(ctx, &deadlines); err != nil {
		return nil, err
	}
	return deadlines, nil
}

func (r *DeadlineRepositoryImpl) MarkNotified(ctx context.Context, taskID uuid.UUID) error {
	_, err := r.collection().UpdateOne(
		ctx,
		bson.M{"task_id": taskID},
		bson.M{"$set": bson.M{"notofied": true}},
	)
	return err
}

func (r *DeadlineRepositoryImpl) DeleteByTaskID(ctx context.Context, taskID uuid.UUID) error {
	_, err := r.collection().DeleteOne(ctx, bson.M{"task_id": taskID})
	return err
}

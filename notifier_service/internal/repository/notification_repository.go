package repository

import (
	"context"
	"errors"
	"time"

	"notifier_service/internal/config"
	"notifier_service/internal/db/mongo"
	"notifier_service/internal/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var(
	_ NotificationRepository = (*NotificationRepositoryImpl)(nil)
	ErrNotificationNotFound = errors.New("notification not found")
)


type NotificationRepositoryImpl struct {}

func NewNotificationRepository() *NotificationRepositoryImpl {
	return &NotificationRepositoryImpl{}
}

func (r *NotificationRepositoryImpl) collection() *mongodriver.Collection {
	return mongo.MongoDB.Collection(config.MongoDbCollection)
}

func (r *NotificationRepositoryImpl) Create(ctx context.Context, n *models.Notification) error {
	n.CreatedAt = time.Now()

	_, err := r.collection().InsertOne(ctx, n)
	return err
}

func (r *NotificationRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.Notification, int64, error) {
	filter := bson.M{"user_id": userID}

	total, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	notifications := []models.Notification{}
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}
	return notifications, total, nil
}

func (r *NotificationRepositoryImpl) UpdateStatus(ctx context.Context, id string, status models.NotificationStatus, errorMsg string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	if status == models.StatusSent {
		now := time.Now()
		update["$set"].(bson.M)["sent_at"] = now
	}

	if errorMsg != "" {
		update["$set"].(bson.M)["error"] = errorMsg
	}

	result, err := r.collection().UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrNotificationNotFound
	}
	return nil
}
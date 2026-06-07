package repository

import (
	"context"
	"time"

	"task_service/internal/db/mongo"
	"task_service/internal/models"
	"task_service/internal/config"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var _ NoteRepository = (*NoteRepositoryImpl)(nil)

type NoteRepositoryImpl struct {}

func NewNoteRepository() *NoteRepositoryImpl {
	return &NoteRepositoryImpl{}
}

func (r *NoteRepositoryImpl) CreateNote(ctx context.Context, note *models.Note) error {
	note.CreatedAt = time.Now()

	_, err := mongo.MongoDB.Collection(config.MongoDbCollection).InsertOne(ctx, note)
	return err
}

func (r *NoteRepositoryImpl) DeleteNote(ctx context.Context, noteID bson.ObjectID, userID uuid.UUID) error {
	filter := bson.M{
		"_id": noteID,
		"author_id": userID,
	}

	_, err := mongo.MongoDB.Collection(config.MongoDbCollection).DeleteOne(ctx, filter)
	return err
}

func (r *NoteRepositoryImpl) GetAllByTask(ctx context.Context, taskID uuid.UUID) ([]models.Note, error) {
	filter := bson.M{"task_id": taskID}

	cursor, err := mongo.MongoDB.Collection(config.MongoDbCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notes []models.Note
	if err = cursor.All(ctx, &notes); err != nil {
		return nil, err
	}

	return notes, nil
}
package repository

import (
	"context"

	"task_service/internal/models"
	"task_service/internal/dto"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *models.Task) error
    GetAllByUser(ctx context.Context, userID uuid.UUID) ([]models.Task, error)
    GetTaskByID(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error)
    UpdateTask(ctx context.Context, taskID, userID uuid.UUID, req dto.UpdateTaskRequest) error
    DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error
}

type NoteRepository interface {
    CreateNote(ctx context.Context, note *models.Note) error
    DeleteNote(ctx context.Context, noteID bson.ObjectID, userID uuid.UUID) error
    GetAllByTask(ctx context.Context, taskID uuid.UUID) ([]models.Note, error)
}

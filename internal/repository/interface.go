package repository

import (
	"context"

	"todo/internal/models"
	"todo/internal/dto"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	ExistByEmail(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user *models.User) error
}

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

type SessionRepository interface {
    CreateSession(ctx context.Context, session *models.Session) error
    GetSessionByToken(ctx context.Context, refreshToken string) (*models.Session, error)
    DeleteSessionByToken(ctx context.Context, refreshToken string) error
    DeleteAllSessionsByUser(ctx context.Context, userID uuid.UUID) error
}
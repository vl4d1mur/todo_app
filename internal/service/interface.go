package service

import (
    "context"

    "todo/internal/dto"
    "todo/internal/models"
    "todo/pkg/pagination"

    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/v2/bson"
)

type AuthServiceInterface interface {
    Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error)
    Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenPair, error)
    GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
    RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenPair, error)
    Logout(ctx context.Context, refreshToken string) error
}

type TaskServiceInterface interface {
    CreateTask(ctx context.Context, userID uuid.UUID, req dto.CreateTaskRequest) (*models.Task, error)
    UpdateTask(ctx context.Context, taskID, userID uuid.UUID, req dto.UpdateTaskRequest) (*models.Task, error)
    DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error
    GetAllByUser(ctx context.Context, userID uuid.UUID, q pagination.Query) ([]models.Task, int64, error)
    GetTaskByID(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error)
}

type NoteServiceInterface interface {
    CreateNote(ctx context.Context, taskID, userID uuid.UUID, req dto.CreateNoteRequest) (*models.Note, error)
    DeleteNote(ctx context.Context, noteID bson.ObjectID, userID uuid.UUID) error
    GetAllByTask(ctx context.Context, taskID uuid.UUID) ([]models.Note, error)
}

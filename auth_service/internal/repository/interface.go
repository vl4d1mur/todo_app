package repository

import (
	"context"

	"auth_service/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	ExistByEmail(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user *models.User) error
}

type SessionRepository interface {
    CreateSession(ctx context.Context, session *models.Session) error
    GetSessionByToken(ctx context.Context, refreshToken string) (*models.Session, error)
    DeleteSessionByToken(ctx context.Context, refreshToken string) error
    DeleteAllSessionsByUser(ctx context.Context, userID uuid.UUID) error
}
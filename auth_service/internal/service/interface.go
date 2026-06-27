package service

import (
	"context"

	"auth_service/internal/dto"
	"auth_service/internal/models"

	"github.com/google/uuid"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenPair, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenPair, error)
	Logout(ctx context.Context, refreshToken, accessToken string) error
}

type TelegramServiceInterface interface {
	GenerateTelegramCode(ctx context.Context, userID uuid.UUID) (string, error)
}

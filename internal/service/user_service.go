package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"todo/internal/db/redisConn"
	"todo/internal/dto"
	"todo/internal/models"
	"todo/internal/repository"
	"todo/pkg/log"
	"todo/pkg/hash"
	"todo/pkg/jwt"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not")
	ErrUserAlrdeadyExists = errors.New("user with this email already exists")
	ErrInvalidPassword = errors.New("invalid email or password")
)

type AuthService struct {
    repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *AuthService {
    return &AuthService{repo: repo}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error) {
	exists, err := s.repo.ExistByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, ErrUserAlrdeadyExists
	}

	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User {
		ID: uuid.New(),
		Email: req.Email,
		Password: hashedPassword,
		Name: req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (string, *models.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", nil, ErrInvalidPassword
		}
		return "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !hash.CheckPasswordHash(req.Password, user.Password) {
		return "", nil, ErrInvalidPassword
	}

	token, err := jwt.GenerateJWT(*user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}
	return token, user, nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	userIDStr := userID.String()

	cached, err := redisConn.GetCachedUserProfile(userIDStr)
	if err == nil {
		return cached, nil
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	if err := redisConn.CacheUserProfile(userIDStr, *user); err != nil {
		log.Logger.Warn().Err(err).Msg("Failed to cache user profile")
	}

	return user, nil
}
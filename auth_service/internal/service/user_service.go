package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth_service/internal/db/redisConn"
	"auth_service/internal/dto"
	"auth_service/internal/models"
	"auth_service/internal/repository"
	"auth_service/pkg/log"
	"auth_service/pkg/hash"
	"auth_service/pkg/jwt"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserAlrdeadyExists = errors.New("user with this email already exists")
	ErrInvalidPassword = errors.New("invalid email or password")
	ErrInvalidSession = errors.New("invalid or expired refresh token")
)

type AuthService struct {
    userRepo repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository) *AuthService {
    return &AuthService{
		userRepo: userRepo,
		sessionRepo:sessionRepo,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error) {
	exists, err := s.userRepo.ExistByEmail(ctx, req.Email)
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

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenPair, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !hash.CheckPasswordHash(req.Password, user.Password) {
		return nil, ErrInvalidPassword
	}

	tokens, err := s.createTokenPair(ctx, user)
    if err != nil {
        return nil, err
    }

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
    if err := s.sessionRepo.DeleteSessionByToken(ctx, refreshToken); err != nil {
        if errors.Is(err, repository.ErrSessionNotFound) {
            return ErrInvalidSession
        }
        return fmt.Errorf("failed to delete session: %w", err)
    }
    return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	userIDStr := userID.String()

	cached, err := redisConn.GetCachedUserProfile(userIDStr)
	if err == nil {
		return cached, nil
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
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

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenPair, error) {
    if _, err := jwt.ParseRefresh(refreshToken); err != nil {
        return nil, ErrInvalidSession
    }

    session, err := s.sessionRepo.GetSessionByToken(ctx, refreshToken)
    if err != nil {
        if errors.Is(err, repository.ErrSessionNotFound) {
            return nil, ErrInvalidSession
        }
        return nil, fmt.Errorf("failed to get session: %w", err)
    }

    if err := s.sessionRepo.DeleteSessionByToken(ctx, refreshToken); err != nil {
        return nil, fmt.Errorf("failed to delete old session: %w", err)
    }

    user, err := s.userRepo.GetUserByID(ctx, session.UserID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    return s.createTokenPair(ctx, user)
}

func (s *AuthService) createTokenPair(ctx context.Context, user *models.User) (*dto.TokenPair, error) {
    accessToken, err := jwt.GenerateAccess(*user)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }

    refreshToken, err := jwt.GenerateRefresh(*user)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }

    session := &models.Session{
        ID:           uuid.New(),
        UserID:       user.ID,
        RefreshToken: refreshToken,
        ExpiresAt:    time.Now().Add(jwt.RefreshTokenTTL),
        CreatedAt:    time.Now(),
    }

    if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
        return nil, fmt.Errorf("failed to create session: %w", err)
    }

    return &dto.TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}

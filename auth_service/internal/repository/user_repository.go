package repository

import (
	"context"
	"errors"

	"auth_service/internal/db/postgres"
	"auth_service/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlrdeadyExists = errors.New("user with this email already exists")
)
var _ UserRepository = (*UserRepositoryImpl)(nil)

type UserRepositoryImpl struct{}

func NewUserRepository() *UserRepositoryImpl {
	return &UserRepositoryImpl{}
}

func (r *UserRepositoryImpl) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var u models.User
	err := postgres.DB.QueryRow(ctx, `SELECT id, email, name, telegram_chat_id, created_at, updated_at FROM users WHERE id = $1`,
		userID).Scan(&u.ID, &u.Email, &u.Name, &u.TelegramChatID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := postgres.DB.QueryRow(ctx,
		`SELECT id, email, password, name, created_at, updated_at FROM users WHERE email = $1`,
		email).Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepositoryImpl) ExistByEmail(ctx context.Context, email string) (bool, error) {
	var id uuid.UUID
	err := postgres.DB.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *UserRepositoryImpl) CreateUser(ctx context.Context, user *models.User) error {
	_, err := postgres.DB.Exec(ctx, `INSERT INTO users (id, email, password, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email, user.Password, user.Name, user.CreatedAt, user.UpdatedAt)

	return err
}

func (r *UserRepositoryImpl) UpdateTelegramChatID(ctx context.Context, userID uuid.UUID, chatID int64) error {
	_, err := postgres.DB.Exec(ctx,
		"UPDATE users SET telegram_chat_id = $1, updated_at = NOW() WHERE id = $2",
		chatID, userID)
	return err
}

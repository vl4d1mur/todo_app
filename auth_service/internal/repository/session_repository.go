package repository

import (
	"context"
	"errors"
	"time"

	"auth_service/internal/db/postgres"
	"auth_service/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrSessionNotFound = errors.New("session not found or expired")

var _ SessionRepository = (*SessionRepositoryImpl)(nil)

type SessionRepositoryImpl struct{}

func NewSessionRepository() *SessionRepositoryImpl {
	return &SessionRepositoryImpl{}
}

func (r *SessionRepositoryImpl) CreateSession(ctx context.Context, session *models.Session) error {
	_, err := postgres.DB.Exec(ctx,
		`INSERT INTO sessions (id, user_id, refresh_token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`, session.ID, session.UserID, session.RefreshToken, session.ExpiresAt, session.CreatedAt)

	return err
}

func (r *SessionRepositoryImpl) GetSessionByToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	var s models.Session
	err := postgres.DB.QueryRow(ctx, `SELECT id, user_id, refresh_token, expires_at, created_at
		FROM sessions WHERE refresh_token = $1`, refreshToken).Scan(&s.ID, &s.UserID, &s.RefreshToken, &s.ExpiresAt, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	if time.Now().After(s.ExpiresAt) {
		r.DeleteSessionByToken(ctx, refreshToken)
		return nil, ErrSessionNotFound
	}

	return &s, nil
}

func (r *SessionRepositoryImpl) DeleteSessionByToken(ctx context.Context, refreshToken string) error {
	_, err := postgres.DB.Exec(ctx, `DELETE FROM sessions WHERE refresh_token = $1`, refreshToken)
	return err
}

func (r *SessionRepositoryImpl) DeleteAllSessionsByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := postgres.DB.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

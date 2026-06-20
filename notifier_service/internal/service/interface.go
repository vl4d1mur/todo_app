package service

import (
	"context"

	"notifier_service/internal/models"
	"notifier_service/pkg/pagination"

	"github.com/google/uuid"
)

type NotifierServiceInterface interface {
	HandleStatusChanged(ctx context.Context, userID, taskID uuid.UUID, newStatus, taskTitle string) error
	HandleDeadlineApproaching(ctx context.Context, userID, taskID uuid.UUID, taskTitle, deadline string) error
	GetByUserID(ctx context.Context, userID uuid.UUID, q pagination.Query) ([]models.Notification, int64, error)
}

type AuthClient interface {
	GetUserEmail(ctx context.Context, userID string) (string, error)
}

package service

import (
	"context"
	"time"

	"notifier_service/internal/models"
	"notifier_service/pkg/pagination"

	"github.com/google/uuid"
)

type NotifierServiceInterface interface {
	HandleStatusChanged(ctx context.Context, userID, taskID uuid.UUID, newStatus, taskTitle string) error
	HandleDeadlineApproaching(ctx context.Context, userID, taskID uuid.UUID, taskTitle, deadline string) error
	GetByUserID(ctx context.Context, userID uuid.UUID, q pagination.Query) ([]models.Notification, int64, error)
	HandleTaskMutation(ctx context.Context, userID, taskID uuid.UUID, title string, deadline *time.Time) error
	HandleTaskDeletion(ctx context.Context, taskID uuid.UUID) error
}

type AuthClient interface {
	GetUserEmail(ctx context.Context, userID string) (string, error)
}

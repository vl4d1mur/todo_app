package repository

import (
	"context"

	"notifier_service/internal/models"

	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) error
	GetByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.Notification, int64, error)
	UpdateStatus(ctx context.Context, id string, status models.NotificationStatus, errorMsg string) error
}

type DeadlineRepository interface {
	Upsert(ctx context.Context, deadline *models.TaskDeadline) error
	DeleteByTaskID(ctx context.Context, taskID uuid.UUID) error
	GetPending(ctx context.Context) ([]models.TaskDeadline, error)
	MarkNotified(ctx context.Context, taskID uuid.UUID) error
}

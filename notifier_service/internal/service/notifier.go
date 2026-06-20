package service

import (
	"context"
	"fmt"
	"time"

	"notifier_service/internal/models"
	"notifier_service/internal/repository"
	"notifier_service/pkg/log"
	"notifier_service/pkg/pagination"

	"github.com/google/uuid"
)

type NotifierService struct {
	repo       repository.NotificationRepository
	authClient AuthClient
	smtp       *SMTPService
}

func NewNotifierService(
	repo repository.NotificationRepository,
	authClient AuthClient,
	smtp *SMTPService,
) *NotifierService {
	return &NotifierService{
		repo:       repo,
		authClient: authClient,
		smtp:       smtp,
	}
}

func (s *NotifierService) createAndSend(
	ctx context.Context,
	userID, taskID uuid.UUID,
	eventType, recipient, subject, body string,
) error {
	notification := &models.Notification{
		UserID:    userID,
		TaskID:    taskID,
		EventType: eventType,
		Channel:   "email",
		Recipient: recipient,
		Subject:   subject,
		Body:      body,
		Status:    models.StatusPending,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to save notification")
		return err
	}

	if err := s.smtp.Send(recipient, subject, body); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to send email")
		if updateErr := s.repo.UpdateStatus(ctx, notification.ID.Hex(), models.StatusFailed, err.Error()); updateErr != nil {
			log.Logger.Error().Err(updateErr).Msg("Failed to update notification status to failed")
		}
		return err
	}

	if err := s.repo.UpdateStatus(ctx, notification.ID.Hex(), models.StatusSent, ""); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to update notification status to sent")
	}
	_ = time.Now()
	return nil
}

func (s *NotifierService) HandleStatusChanged(ctx context.Context, userID, taskID uuid.UUID, newStatus, taskTitle string) error {
	email, err := s.authClient.GetUserEmail(ctx, userID.String())
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get user email")
		return err
	}

	subject := fmt.Sprintf("Task status changed: %s", taskTitle)
	body := fmt.Sprintf("Hello, Your task \"%s\" has changed status to: %s", taskTitle, newStatus)

	return s.createAndSend(ctx, userID, taskID, "task.status_changed", email, subject, body)
}

func (s *NotifierService) HandleDeadlineApproaching(ctx context.Context, userID, taskID uuid.UUID, taskTitle, deadline string) error {
	email, err := s.authClient.GetUserEmail(ctx, userID.String())
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get user email")
		return err
	}

	subject := fmt.Sprintf("Deadline approaching: %s", taskTitle)
	body := fmt.Sprintf("Hello, Your task \"%s\" is due soon: %s", taskTitle, deadline)

	return s.createAndSend(ctx, userID, taskID, "task.deadline_approaching", email, subject, body)
}

func (s *NotifierService) GetByUserID(ctx context.Context, userID uuid.UUID, q pagination.Query) ([]models.Notification, int64, error) {
	q = pagination.NewPaginationQuery(q.Page, q.Limit, "")
	return s.repo.GetByUserID(ctx, userID, q.Page, q.Limit)
}

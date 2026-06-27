package service

import (
	"context"
	"fmt"
	"time"

	"notifier_service/internal/models"
	"notifier_service/internal/repository"
	"notifier_service/pkg/log"
	"notifier_service/pkg/metrics"
	"notifier_service/pkg/pagination"

	"github.com/google/uuid"
)

type NotifierService struct {
	repo         repository.NotificationRepository
	deadlineRepo repository.DeadlineRepository
	authClient   AuthClient
	smtp         *SMTPService
	telegram     *TelegramChannel
}

func NewNotifierService(
	repo repository.NotificationRepository,
	deadlineRepo repository.DeadlineRepository,
	authClient AuthClient,
	smtp *SMTPService,
	telegram *TelegramChannel,
) *NotifierService {
	return &NotifierService{
		repo:         repo,
		deadlineRepo: deadlineRepo,
		authClient:   authClient,
		smtp:         smtp,
		telegram:     telegram,
	}
}

func (s *NotifierService) sendEmail(
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

	metrics.NotificationsSent.WithLabelValues("email", eventType).Inc()
	return nil
}

func (s *NotifierService) sendTelegram(ctx context.Context, userID, taskID uuid.UUID, eventType string, chatID int64, subject, body string) error {
	notification := &models.Notification{
		UserID:    userID,
		TaskID:    taskID,
		EventType: eventType,
		Channel:   "telegram",
		Recipient: fmt.Sprintf("%d", chatID),
		Subject:   subject,
		Body:      body,
		Status:    models.StatusPending,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return err
	}

	if err := s.telegram.Send(chatID, subject, body); err != nil {
		if updateErr := s.repo.UpdateStatus(ctx, notification.ID.Hex(), models.StatusFailed, err.Error()); updateErr != nil {
			log.Logger.Error().Err(updateErr).Msg("Failed to update notification status to failed")
			return err
		}
	}

	metrics.NotificationsSent.WithLabelValues("telegram", eventType).Inc()
	return s.repo.UpdateStatus(ctx, notification.ID.Hex(), models.StatusSent, "")
}

func (s *NotifierService) HandleStatusChanged(ctx context.Context, userID, taskID uuid.UUID, newStatus, taskTitle string) error {
	email, chatID, hasTelegram, err := s.authClient.GetUserContacts(ctx, userID.String())
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get user email")
		return err
	}

	_ = chatID
	_ = hasTelegram

	subject := fmt.Sprintf("Task status changed: %s", taskTitle)
	body := fmt.Sprintf("Hello, Your task \"%s\" has changed status to: %s", taskTitle, newStatus)

	return s.sendEmail(ctx, userID, taskID, "task.status_changed", email, subject, body)
}

func (s *NotifierService) HandleDeadlineApproaching(ctx context.Context, userID, taskID uuid.UUID, taskTitle, deadline string) error {
	email, chatID, hasTelegram, err := s.authClient.GetUserContacts(ctx, userID.String())
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get user contacts")
		return err
	}
	log.Logger.Info().
		Str("email", email).
		Int64("chat_id", chatID).
		Bool("has_telegram", hasTelegram).
		Msg("Got user contacts")

	subject := fmt.Sprintf("Deadline approaching: %s", taskTitle)
	body := fmt.Sprintf("Hello, Your task \"%s\" is due soon: %s", taskTitle, deadline)

	if email != "" {
		if err := s.sendEmail(ctx, userID, taskID, "task.deadline_approaching", email, subject, body); err != nil {
			log.Logger.Error().Err(err).Msg("Email channel failed")
		}
	}

	if hasTelegram {
		if err := s.sendTelegram(ctx, userID, taskID, "task.deadline_approaching", chatID, subject, body); err != nil {
			log.Logger.Error().Err(err).Msg("Telegram channel failed")
		}
	}

	return nil
}

func (s *NotifierService) HandleTaskMutation(ctx context.Context, userID, taskID uuid.UUID, title string, deadline *time.Time) error {
	if deadline == nil {
		return s.deadlineRepo.DeleteByTaskID(ctx, taskID)
	}

	deadlineEntry := &models.TaskDeadline{
		TaskID:   taskID,
		UserID:   userID,
		Title:    title,
		Deadline: *deadline,
	}

	if err := s.deadlineRepo.Upsert(ctx, deadlineEntry); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to upsert deadline")
		return err
	}

	return nil
}

func (s *NotifierService) HandleTaskDeletion(ctx context.Context, taskID uuid.UUID) error {
	if err := s.deadlineRepo.DeleteByTaskID(ctx, taskID); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to delete deadline")
		return err
	}
	return nil
}

func (s *NotifierService) GetByUserID(ctx context.Context, userID uuid.UUID, q pagination.Query) ([]models.Notification, int64, error) {
	q = pagination.NewPaginationQuery(q.Page, q.Limit, "")
	return s.repo.GetByUserID(ctx, userID, q.Page, q.Limit)
}

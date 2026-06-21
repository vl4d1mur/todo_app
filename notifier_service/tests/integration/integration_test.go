//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"notifier_service/internal/config"
	"notifier_service/internal/consumer"
	"notifier_service/internal/db/mongo"
	"notifier_service/internal/models"
	"notifier_service/internal/repository"
	"notifier_service/internal/service"
	"notifier_service/pkg/log"

	"github.com/google/uuid"
	natsclient "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tcMongo "github.com/testcontainers/testcontainers-go/modules/mongodb"
	tcNats "github.com/testcontainers/testcontainers-go/modules/nats"
	"github.com/testcontainers/testcontainers-go/wait"
)

// mockAuthClient — заглушка для gRPC к auth_service.
// В интеграционном тесте notifier мы не хотим тащить весь auth_service.
type mockAuthClient struct {
	email       string
	chatID      int64
	hasTelegram bool
}

type mockTelegramSender struct {
	sentMessages []sentMessage
	mu           sync.Mutex
}

type sentMessage struct {
	ChatID int64
	Text   string
}

func (m *mockTelegramSender) SendMessage(chatID int64, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = append(m.sentMessages, sentMessage{ChatID: chatID, Text: text})
	return nil
}

func (m *mockTelegramSender) Messages() []sentMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]sentMessage{}, m.sentMessages...)
}

func (m *mockAuthClient) GetUserContacts(ctx context.Context, userID string) (string, int64, bool, error) {
	return m.email, m.chatID, m.hasTelegram, nil
}

func (m *mockAuthClient) ActivateTelegram(ctx context.Context, code string, chatID int64) (bool, error) {
	return true, nil
}

func TestNotifier_StatusChanged_EndToEnd(t *testing.T) {
	log.InitLogger()
	ctx := context.Background()

	// 1. Поднимаем MongoDB контейнер
	mongoContainer, err := tcMongo.Run(ctx, "mongo:7")
	assert.NoError(t, err)
	defer mongoContainer.Terminate(ctx)

	mongoURI, err := mongoContainer.ConnectionString(ctx)
	assert.NoError(t, err)

	// 2. Поднимаем NATS контейнер
	natsContainer, err := tcNats.Run(ctx, "nats:2-alpine")
	assert.NoError(t, err)
	defer natsContainer.Terminate(ctx)

	natsURL, err := natsContainer.ConnectionString(ctx)
	assert.NoError(t, err)

	// 3. Поднимаем MailHog контейнер
	mailhogReq := testcontainers.ContainerRequest{
		Image:        "mailhog/mailhog",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForListeningPort("1025/tcp"),
	}
	mailhogContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: mailhogReq,
		Started:          true,
	})
	assert.NoError(t, err)
	defer mailhogContainer.Terminate(ctx)

	mailhogHost, err := mailhogContainer.Host(ctx)
	assert.NoError(t, err)
	mailhogSMTPPort, err := mailhogContainer.MappedPort(ctx, "1025")
	assert.NoError(t, err)
	mailhogAPIPort, err := mailhogContainer.MappedPort(ctx, "8025")
	assert.NoError(t, err)

	// 4. Настраиваем config для тестового окружения
	config.MongoUri = mongoURI
	config.MongoDbName = "notifier_test"
	config.MongoDbCollection = "notifications"
	config.MongoDeadlineCollection = "task_deadlines"
	config.NatsURL = natsURL
	config.SmtpHost = mailhogHost
	config.SmtpPort = mailhogSMTPPort.Port()
	config.SmtpFrom = "test@test.local"

	// 5. Подключаемся к Mongo (используя реальный коннектор)
	mongo.ConnectMongo()
	defer mongo.CloseMongoDB()

	// 6. Создаём сервис со всеми зависимостями
	notifRepo := repository.NewNotificationRepository()
	deadlineRepo := repository.NewDeadlineRepository()
	smtpSvc := service.NewSMTPService()
	mockTg := &mockTelegramSender{}
	telegramChannel := service.NewTelegramChannel(mockTg)
	mockAuth := &mockAuthClient{email: "user@example.com"}
	notifierSvc := service.NewNotifierService(notifRepo, deadlineRepo, mockAuth, smtpSvc, telegramChannel)

	// 7. Запускаем NATS consumer
	natsConsumer, err := consumer.NewConsumer(notifierSvc)
	assert.NoError(t, err)
	err = natsConsumer.Start()
	assert.NoError(t, err)
	defer natsConsumer.Close()

	time.Sleep(200 * time.Millisecond) // даём consumer'у подписаться

	// 8. Публикуем событие в NATS как делает task_service
	pubConn, err := natsclient.Connect(natsURL)
	assert.NoError(t, err)
	defer pubConn.Close()

	userID := uuid.New()
	taskID := uuid.New()
	event := models.TaskEvent{
		EventType: "task.status_changed",
		TaskID:    taskID,
		UserID:    userID,
		Timestamp: time.Now(),
		Payload: map[string]any{
			"new_status": "done",
			"title":      "Test Task",
		},
	}
	data, _ := json.Marshal(event)
	err = pubConn.Publish("task-events", data)
	assert.NoError(t, err)
	pubConn.Flush()

	// 9. Ждём что в MongoDB появится запись со статусом sent
	assert.Eventually(t, func() bool {
		notifications, _, err := notifRepo.GetByUserID(ctx, userID, 1, 10)
		if err != nil || len(notifications) == 0 {
			return false
		}
		return notifications[0].Status == models.StatusSent
	}, 10*time.Second, 200*time.Millisecond, "Notification was not saved with Sent status")

	// 10. Проверяем что письмо реально дошло до MailHog через его API
	mailhogAPIURL := fmt.Sprintf("http://%s:%s/api/v2/messages", mailhogHost, mailhogAPIPort.Port())
	resp, err := http.Get(mailhogAPIURL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	assert.Contains(t, bodyStr, "user@example.com", "Email should be sent to test user")
	assert.Contains(t, bodyStr, "Test Task", "Email body should mention task title")
	assert.True(t, strings.Contains(bodyStr, "task.status_changed") || strings.Contains(bodyStr, "Task status changed"),
		"Email should be about status change")
}

func TestNotifier_DeadlineFlow(t *testing.T) {
	log.InitLogger()
	ctx := context.Background()

	// CONTAINERS
	mongoContainer, err := tcMongo.Run(ctx, "mongo:7")
	assert.NoError(t, err)
	defer mongoContainer.Terminate(ctx)

	mongoURI, err := mongoContainer.ConnectionString(ctx)
	assert.NoError(t, err)

	natsContainer, err := tcNats.Run(ctx, "nats:2-alpine")
	assert.NoError(t, err)
	defer natsContainer.Terminate(ctx)

	natsURL, err := natsContainer.ConnectionString(ctx)
	assert.NoError(t, err)

	// CONFIG

	config.MongoUri = mongoURI
	config.MongoDbName = "notifier_test"
	config.MongoDbCollection = "notifications"
	config.MongoDeadlineCollection = "task_deadlines"
	config.NatsURL = natsURL

	mongo.ConnectMongo()
	defer mongo.CloseMongoDB()

	// SERVICES

	notifRepo := repository.NewNotificationRepository()
	deadlineRepo := repository.NewDeadlineRepository()
	smtpSvc := service.NewSMTPService()
	mockTg := &mockTelegramSender{}
	telegramChannel := service.NewTelegramChannel(mockTg)
	mockAuth := &mockAuthClient{email: "user@example.com"}
	notifierSvc := service.NewNotifierService(notifRepo, deadlineRepo, mockAuth, smtpSvc, telegramChannel)

	natsConsumer, err := consumer.NewConsumer(notifierSvc)
	assert.NoError(t, err)
	err = natsConsumer.Start()
	assert.NoError(t, err)
	defer natsConsumer.Close()

	time.Sleep(200 * time.Millisecond)

	pubConn, err := natsclient.Connect(natsURL)
	assert.NoError(t, err)
	defer pubConn.Close()

	userID := uuid.New()
	taskID := uuid.New()

	// TASK CREATED

	firstDeadline := time.Now().Add(12 * time.Hour).UTC().Truncate(time.Second)
	createEvent := models.TaskEvent{
		EventType: "task.created",
		TaskID:    taskID,
		UserID:    userID,
		Timestamp: time.Now(),
		Payload: map[string]any{
			"title":    "Important task",
			"deadline": firstDeadline.Format(time.RFC3339),
		},
	}
	data, _ := json.Marshal(createEvent)
	err = pubConn.Publish("task-events", data)
	assert.NoError(t, err)
	pubConn.Flush()

	assert.Eventually(t, func() bool {
		deadlines, err := deadlineRepo.GetPending(ctx)
		if err != nil {
			return false
		}
		for _, d := range deadlines {
			if d.TaskID == taskID && d.Title == "Important task" {
				return true
			}
		}
		return false
	}, 10*time.Second, 200*time.Millisecond, "Deadline should be saved after task.created")

	// TASK UPDATED — DEADLINE CHANGED

	// Помечаем как уведомлённое чтобы проверить сброс флага
	err = deadlineRepo.MarkNotified(ctx, taskID)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	newDeadline := firstDeadline.Add(1 * time.Hour)
	updateEvent := models.TaskEvent{
		EventType: "task.updated",
		TaskID:    taskID,
		UserID:    userID,
		Timestamp: time.Now(),
		Payload: map[string]any{
			"title":    "Important task",
			"deadline": newDeadline.Format(time.RFC3339),
		},
	}
	data, _ = json.Marshal(updateEvent)
	err = pubConn.Publish("task-events", data)
	assert.NoError(t, err)
	pubConn.Flush()

	assert.Eventually(t, func() bool {
		deadlines, err := deadlineRepo.GetPending(ctx)
		if err != nil {
			return false
		}
		for _, d := range deadlines {
			if d.TaskID == taskID && !d.Notified {
				return true
			}
		}
		return false
	}, 5*time.Second, 200*time.Millisecond, "Notified flag should be reset after deadline change")

	// TASK DELETED

	deleteEvent := models.TaskEvent{
		EventType: "task.deleted",
		TaskID:    taskID,
		UserID:    userID,
		Timestamp: time.Now(),
	}
	data, _ = json.Marshal(deleteEvent)
	err = pubConn.Publish("task-events", data)
	assert.NoError(t, err)
	pubConn.Flush()

	assert.Eventually(t, func() bool {
		deadlines, err := deadlineRepo.GetPending(ctx)
		if err != nil {
			return false
		}
		for _, d := range deadlines {
			if d.TaskID == taskID {
				return false
			}
		}
		return true
	}, 5*time.Second, 200*time.Millisecond, "Deadline should be deleted after task.deleted")
}

func TestNotifier_DeadlineFlow_WithTelegram(t *testing.T) {
	log.InitLogger()
	ctx := context.Background()

	// Контейнеры
	mongoContainer, err := tcMongo.Run(ctx, "mongo:7")
	assert.NoError(t, err)
	defer mongoContainer.Terminate(ctx)
	mongoURI, _ := mongoContainer.ConnectionString(ctx)

	natsContainer, err := tcNats.Run(ctx, "nats:2-alpine")
	assert.NoError(t, err)
	defer natsContainer.Terminate(ctx)
	natsURL, _ := natsContainer.ConnectionString(ctx)

	mailhogReq := testcontainers.ContainerRequest{
		Image:        "mailhog/mailhog",
		ExposedPorts: []string{"1025/tcp"},
		WaitingFor:   wait.ForListeningPort("1025/tcp"),
	}
	mailhogContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: mailhogReq,
		Started:          true,
	})
	assert.NoError(t, err)
	defer mailhogContainer.Terminate(ctx)
	mailhogHost, _ := mailhogContainer.Host(ctx)
	mailhogPort, _ := mailhogContainer.MappedPort(ctx, "1025")

	// Config
	config.MongoUri = mongoURI
	config.MongoDbName = "notifier_test"
	config.MongoDbCollection = "notifications"
	config.MongoDeadlineCollection = "task_deadlines"
	config.NatsURL = natsURL
	config.SmtpHost = mailhogHost
	config.SmtpPort = mailhogPort.Port()
	config.SmtpFrom = "test@test.local"

	mongo.ConnectMongo()
	defer mongo.CloseMongoDB()

	// Services с привязанным Telegram
	notifRepo := repository.NewNotificationRepository()
	deadlineRepo := repository.NewDeadlineRepository()
	smtpSvc := service.NewSMTPService()
	mockTg := &mockTelegramSender{}
	telegramChannel := service.NewTelegramChannel(mockTg)
	mockAuth := &mockAuthClient{
		email:       "user@example.com",
		chatID:      123456789,
		hasTelegram: true,
	}
	notifierSvc := service.NewNotifierService(notifRepo, deadlineRepo, mockAuth, smtpSvc, telegramChannel)

	// Эмулируем срабатывание cron — напрямую вызываем HandleDeadlineApproaching
	userID := uuid.New()
	taskID := uuid.New()

	err = notifierSvc.HandleDeadlineApproaching(ctx, userID, taskID, "Test deadline task", "2026-12-31T23:59:59Z")
	assert.NoError(t, err)

	// Проверяем что Telegram получил сообщение
	assert.Eventually(t, func() bool {
		messages := mockTg.Messages()
		if len(messages) == 0 {
			return false
		}
		for _, m := range messages {
			if m.ChatID == 123456789 && strings.Contains(m.Text, "Test deadline task") {
				return true
			}
		}
		return false
	}, 3*time.Second, 100*time.Millisecond, "Telegram message should be sent for deadline")

	// Проверяем что в БД появилось две записи: email и telegram
	notifications, _, err := notifRepo.GetByUserID(ctx, userID, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, notifications, 2, "Should have 2 notifications: email and telegram")

	channels := []string{}
	for _, n := range notifications {
		channels = append(channels, n.Channel)
	}
	assert.Contains(t, channels, "email")
	assert.Contains(t, channels, "telegram")
}

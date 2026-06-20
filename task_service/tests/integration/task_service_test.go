//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"task_service/internal/config"
	"task_service/internal/db/mongo"
	"task_service/internal/db/postgres"
	"task_service/internal/db/redisConn"
	"task_service/internal/dto"
	"task_service/internal/events"
	"task_service/internal/models"
	"task_service/internal/repository"
	"task_service/internal/service"
	"task_service/pkg/log"
	"task_service/pkg/pagination"

	"github.com/google/uuid"
	natsclient "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tcMongo "github.com/testcontainers/testcontainers-go/modules/mongodb"
	tcNats "github.com/testcontainers/testcontainers-go/modules/nats"
	tcPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestTaskService_FullFlow(t *testing.T) {
	log.InitLogger()
	ctx := context.Background()

	// ==================== POSTGRES ====================

	_, currentFile, _, _ := runtime.Caller(0)
	//migrationsPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "migrations")
	testdataPath := filepath.Join(filepath.Dir(currentFile), "testdata", "schema.sql")
	pgContainer, err := tcPostgres.Run(ctx,
		"postgres:16-alpine",
		tcPostgres.WithDatabase("tasks_test"),
		tcPostgres.WithUsername("postgres"),
		tcPostgres.WithPassword("root"),
		tcPostgres.WithInitScripts(testdataPath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	assert.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	pgDSN, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err)

	// ==================== REDIS ====================

	redisContainer, err := tcRedis.Run(ctx, "redis:7-alpine")
	assert.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	redisAddr, err := redisContainer.Endpoint(ctx, "")
	assert.NoError(t, err)

	// ==================== MONGO ====================

	mongoContainer, err := tcMongo.Run(ctx, "mongo:7")
	assert.NoError(t, err)
	defer mongoContainer.Terminate(ctx)

	mongoURI, err := mongoContainer.ConnectionString(ctx)
	assert.NoError(t, err)

	// ==================== NATS ====================

	natsContainer, err := tcNats.Run(ctx, "nats:2-alpine")
	assert.NoError(t, err)
	defer natsContainer.Terminate(ctx)

	natsURL, err := natsContainer.ConnectionString(ctx)
	assert.NoError(t, err)

	// ==================== CONFIG ====================

	config.PostgresDSN = pgDSN
	config.RedisAddr = redisAddr
	config.RedisPassword = ""
	config.MongoUri = mongoURI
	config.MongoDbName = "tasks_test"
	config.MongoDbCollection = "notes"
	config.NatsURL = natsURL

	// ==================== CONNECT ====================

	postgres.ConnectPostgres()
	defer postgres.ClosePostgres()

	redisConn.ConnectRedis()
	defer redisConn.CloseRedis()

	mongo.ConnectMongo()
	defer mongo.CloseMongoDB()

	events.ConnectNATS()
	defer events.CloseNATS()

	// ==================== SUBSCRIBER FOR EVENTS ====================

	natsClient, err := natsclient.Connect(natsURL)
	assert.NoError(t, err)
	defer natsClient.Close()

	receivedEvents := make([]string, 0)
	var eventsMutex sync.Mutex
	_, err = natsClient.Subscribe("task-events", func(msg *natsclient.Msg) {
		var event map[string]any
		_ = json.Unmarshal(msg.Data, &event)
		if eventType, ok := event["event_type"].(string); ok {
			eventsMutex.Lock()
			receivedEvents = append(receivedEvents, eventType)
			eventsMutex.Unlock()
		}
	})
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	// ==================== SERVICES ====================

	taskRepo := repository.NewTaskRepository()
	noteRepo := repository.NewNoteRepository()

	taskService := service.NewTaskService(taskRepo)
	noteService := service.NewNoteService(noteRepo)

	userID := uuid.New()

	// ==================== 1. CREATE TASK ====================

	deadline := time.Now().Add(48 * time.Hour)
	task, err := taskService.CreateTask(ctx, userID, dto.CreateTaskRequest{
		Title:       "Integration test task",
		Description: "Created in integration test",
		Status:      models.TaskStatusTodo,
		Priority:    2,
		Deadline:    &deadline,
	})
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "Integration test task", task.Title)

	// ==================== 2. GET TASK FROM POSTGRES ====================

	fetchedTask, err := taskService.GetTaskByID(ctx, task.ID, userID)
	assert.NoError(t, err)
	assert.Equal(t, task.ID, fetchedTask.ID)

	// ==================== 3. CHECK CACHE ====================

	tasks, _, err := taskService.GetAllByUser(ctx, userID, pagination.Query{Page: 1, Limit: 10})
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)

	cachedTasks, cacheErr := redisConn.GetCachedTasksList(userID.String())
	assert.NoError(t, cacheErr)
	assert.Len(t, cachedTasks, 1, "Tasks should be cached after GetAllByUser")

	// ==================== 4. UPDATE TASK ====================

	newStatus := models.TaskStatusInProgress
	_, err = taskService.UpdateTask(ctx, task.ID, userID, dto.UpdateTaskRequest{
		Status: &newStatus,
	})
	assert.NoError(t, err)

	// Cache must be invalidated
	_, cacheErr = redisConn.GetCachedTasksList(userID.String())
	assert.Error(t, cacheErr, "Cache should be invalidated after update")

	// ==================== 5. CREATE NOTE ====================

	note, err := noteService.CreateNote(ctx, task.ID, userID, dto.CreateNoteRequest{
		Text: "Integration test note",
		Meta: map[string]any{"color": "red"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, note)

	// ==================== 6. GET NOTES FOR TASK ====================

	notes, err := noteService.GetAllByTask(ctx, task.ID)
	assert.NoError(t, err)
	assert.Len(t, notes, 1)
	assert.Equal(t, "Integration test note", notes[0].Text)

	// ==================== 7. DELETE TASK ====================

	err = taskService.DeleteTask(ctx, task.ID, userID)
	assert.NoError(t, err)

	// ==================== 8. CHECK EVENTS ====================

	time.Sleep(300 * time.Millisecond) // ждём асинхронной публикации в NATS

	eventsMutex.Lock()
	defer eventsMutex.Unlock()

	assert.Contains(t, receivedEvents, "task.created")
	assert.Contains(t, receivedEvents, "task.status_changed")
	assert.Contains(t, receivedEvents, "task.deleted")
	assert.Contains(t, receivedEvents, "task.note_added")
}

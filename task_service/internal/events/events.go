package events

import (
	"encoding/json"
	"time"

	"task_service/internal/config"
	"task_service/pkg/log"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

var NatsConn *nats.Conn

type TaskEvent struct {
	EventType string    `json:"event_type"`
	TaskID    uuid.UUID `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload,omitempty"`
}

const (
	EventTaskCreated       = "task.created"
	EventTaskStatusChanged = "task.status_changed"
	EventTaskDeleted       = "task.deleted"
	EventNoteAdded         = "task.note_added"
	EventNoteDeleted       = "task.note_deleted"
)

func ConnectNATS() {
	conn, err := nats.Connect(config.NatsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(5),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Logger.Error().Err(err).Msg("NATS disconnected:")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Logger.Warn().Str("NATS reconnected to:", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("NATS connection error:")
	}

	NatsConn = conn
	log.Logger.Info().Msg("Connected to NATS")
}

func CloseNATS() {
	if NatsConn != nil {
		if err := NatsConn.Drain(); err != nil {
			log.Logger.Error().Err(err).Msg("NATS drain error")
		}
		log.Logger.Info().Msg("NATS connection closed")
	}
}

func PublishTaskCreated(taskID, userID uuid.UUID) {
	publishTaskEvent(EventTaskCreated, taskID, userID, nil)
}

func PublishTaskStatusChanged(taskID, userID uuid.UUID, newStatus, title string) {
	publishTaskEvent(EventTaskStatusChanged, taskID, userID, map[string]any{
		"new_status": newStatus,
		"title":      title,
	})
}

func PublishTaskDeadlineApproaching(taskID, userID uuid.UUID, title, deadline string) {
	publishTaskEvent("task.deadline_approaching", taskID, userID, map[string]any{
		"title":    title,
		"deadline": deadline,
	})
}

func PublishTaskDeleted(taskID, userID uuid.UUID) {
	publishTaskEvent(EventTaskDeleted, taskID, userID, nil)
}

func PublishNoteAdded(taskID, userID uuid.UUID, noteID, text string) {
	publishTaskEvent(EventNoteAdded, taskID, userID, map[string]any{
		"note_id": noteID,
		"text":    text,
	})
}

func publishTaskEvent(eventType string, taskID, userID uuid.UUID, payload any) {
	event := TaskEvent{
		EventType: eventType,
		TaskID:    taskID,
		UserID:    userID,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Event marshal error:")
		return
	}

	if NatsConn == nil || !NatsConn.IsConnected() {
		log.Logger.Error().Msg("Publish error: NATS is not connected:")
		return
	}

	if err := NatsConn.Publish("task-events", data); err != nil {
		log.Logger.Error().Err(err).Msg("NATS publish error:")
		return
	}

	log.Logger.Info().
		Str("event_type", eventType).
		Str("task_id", taskID.String()).
		Str("user_id", userID.String()).
		Msg("Task event published")
}

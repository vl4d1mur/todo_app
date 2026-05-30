package events

import (
	"encoding/json"
	//"log"
	"time"

	"todo/internal/config"
	"todo/pkg/log"

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
			//log.Println("NATS disconnected:", err)
			log.Logger.Error().Err(err).Msg("NATS disconnected:")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			//log.Println("NATS reconnected to:", nc.ConnectedUrl())
			log.Logger.Warn().Str("NATS reconnected to:", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		//log.Fatal("NATS connection error:", err)
		log.Logger.Fatal().Err(err).Msg("NATS connection error:")
	}

	NatsConn = conn
	//log.Println("Connected to NATS")
	log.Logger.Info().Msg("Connected to NATS")
}

func CloseNATS() {
	if NatsConn != nil {
		NatsConn.Drain()
		//log.Println("NATS connection closed")
		log.Logger.Info().Msg("NATS connection closed")
	}
}

func PublishTaskEvent(eventType string, taskID, userID uuid.UUID, payload any) {
	event := TaskEvent{
		EventType: eventType,
		TaskID:    taskID,
		UserID:    userID,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	data, err := json.Marshal(event)
	if err != nil {
		//log.Println("Event marshal error:", err)
		log.Logger.Error().Err(err).Msg("Event marshal error:")
		return
	}

	if NatsConn == nil || !NatsConn.IsConnected() {
		log.Logger.Error().Msg("Publish error: NATS is not connected:")
		return
	}

	if err := NatsConn.Publish("task-events", data); err != nil {
		//log.Println("NATS publish error:", err)
		log.Logger.Error().Err(err).Msg("NATS publish error:")
		return
	}

	//log.Printf("Event published: %s task:%s", eventType, taskID)
	log.Logger.Info().
		Str("event_type", eventType).
		Str("task_id", taskID.String()).
		Str("user_id", userID.String()).
		Msg("Task event published")
}

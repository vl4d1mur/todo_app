package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskEvent struct {
	EventType string         `json:"event_type"`
	TaskID    uuid.UUID      `json:"task_id"`
	UserID    uuid.UUID      `json:"user_id"`
	Timestamp time.Time      `json:"timestamp"`
	Payload   map[string]any `json:"payload,omitempty"`
}

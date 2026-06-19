package models 

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type NotificationStatus string

const (
	StatusPending NotificationStatus = "pending"
	StatusSent NotificationStatus = "sent"
	StatusFailed NotificationStatus = ""
)

type Notification struct {
	ID         bson.ObjectID      `bson:"_id,omitempty" json:"id"`
	UserID     uuid.UUID          `bson:"user_id" json:"user_id"`
	TaskID     uuid.UUID          `bson:"task_id" json:"task_id"`
	EventType  string             `bson:"event_type" json:"event_type"`
	Channel    string             `bson:"channel" json:"channel"`
	Recipient  string             `bson:"recipient" json:"recipient"`
	Subject    string             `bson:"subject" json:"subject"`
	Body       string             `bson:"body" json:"body"`
	Status     NotificationStatus `bson:"status" json:"status"`
	SentAt     *time.Time         `bson:"sent_at,omitempty" json:"sent_at,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	Error      string             `bson:"error,omitempty" json:"error,omitempty"`
}
package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TaskDeadline struct {
	ID        bson.ObjectID `json:"id" bson:"_id,omitempty"`
	TaskID    uuid.UUID     `json:"task_id" bson:"task_id"`
	UserID    uuid.UUID     `json:"user_id" bson:"user_id"`
	Title     string        `json:"title" bson:"title"`
	Deadline  time.Time     `json:"deadline" bson:"deadline"`
	Notified  bool          `json:"notified" bson:"notified"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
}

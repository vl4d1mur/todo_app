package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Note struct {
	ID        bson.ObjectID  `json:"id" bson:"_id,omitempty"`
	TaskID    uuid.UUID      `json:"task_id" bson:"task_id"`
	AuthorID  uuid.UUID      `json:"author_id" bson:"author_id"`
	Text      string         `json:"text" bson:"text"`
	Meta      map[string]any `json:"meta" bson:"meta"`
	CreatedAt time.Time      `json:"created_at" bson:"created_at"`
}


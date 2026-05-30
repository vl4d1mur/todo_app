package dto

import (
	"time"
	
	"todo/internal/models"
)

type CreateTaskRequest struct {
	Title       string     `json:"title" validate:"required,min=3,max=200"`
	Description string     `json:"description" validate:"max=1000"`
	Status      models.TaskStatus `json:"status" validate:"oneof=todo in_progress done cancelled"`
	Priority    int        `json:"priority" validate:"min=1,max=4"`
	Deadline    *time.Time `json:"deadline"`
}

type UpdateTaskRequest struct {
	Title       *string     `json:"title,omitempty"`
	Description *string     `json:"description,omitempty"`
	Status      *models.TaskStatus `json:"status,omitempty"`
	Deadline    *time.Time  `json:"deadline,omitempty"`
	Priority    *int        `json:"priority,omitempty"`
}

type TaskFilter struct {
	Status   *models.TaskStatus `json:"status"`
	Priority *int        `json:"priority"`
	Deadline *time.Time  `json:"deadline"`
}
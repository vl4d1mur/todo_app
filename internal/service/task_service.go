package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"todo/internal/db/redisConn"
	"todo/internal/dto"
	"todo/internal/events"
	"todo/internal/models"
	"todo/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrInvalidTaskData = errors.New("invalid task data")
	ErrTaskNotFound    = errors.New("task not found")
)

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(ctx context.Context, userID uuid.UUID, req dto.CreateTaskRequest) (*models.Task, error) {
	task := models.Task{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		DeadLine:    req.Deadline,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if task.Status == "" {
		task.Status = models.TaskStatusTodo
	}

	if err := s.repo.CreateTask(ctx, &task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	redisConn.InvalidateTasksCache(userID.String())
	events.PublishTaskEvent(events.EventTaskCreated, task.ID, task.UserID, nil)

	return &task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID, userID uuid.UUID, req dto.UpdateTaskRequest) (*models.Task, error) {
	if err := s.repo.UpdateTask(ctx, taskID, userID, req); err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	redisConn.InvalidateTasksCache(userID.String())

	if req.Status != nil {
		events.PublishTaskEvent(events.EventTaskStatusChanged, taskID, userID, map[string]any{
			"new_status": *req.Status,
		})
	}

	return s.repo.GetTaskByID(ctx, taskID, userID)
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteTask(ctx, taskID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	redisConn.InvalidateTasksCache(userID.String())
	events.PublishTaskEvent(events.EventTaskDeleted, taskID, userID, nil)

	return nil
}

func (s *TaskService) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]models.Task, error) {
	if tasks, err := redisConn.GetCachedTasksList(userID.String()); err == nil {
		return tasks, nil
	}

	tasks, err := s.repo.GetAllByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	redisConn.CacheTasksList(userID.String(), tasks)
	return tasks, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error) {
	return s.repo.GetTaskByID(ctx, taskID, userID)
}

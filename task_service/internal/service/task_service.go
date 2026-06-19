package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"task_service/internal/db/redisConn"
	"task_service/internal/dto"
	"task_service/internal/events"
	"task_service/internal/models"
	"task_service/internal/repository"
	"task_service/pkg/pagination"

	"github.com/google/uuid"
)

var (
	ErrInvalidTaskData  = errors.New("invalid task data")
	ErrTaskNotFound     = errors.New("task not found")
	ErrInvalidStatus    = errors.New("invalid status")
	ErrTaskAccessDenied = errors.New("access denied")
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

	/*if task.Status != "" {
		switch task.Status {
		case models.TaskStatusTodo, models.TaskStatusInProgress, models.TaskStatusCancelled, models.TaskStatusDone:
		default:
			return nil, ErrInvalidStatus
		}
	}*/

	if err := s.repo.CreateTask(ctx, &task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	redisConn.InvalidateTasksCache(userID.String())
	events.PublishTaskCreated(task.ID, task.UserID)

	return &task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID, userID uuid.UUID, req dto.UpdateTaskRequest) (*models.Task, error) {
	err := s.repo.UpdateTask(ctx, taskID, userID, req)
	if err != nil {
		if errors.Is(err, repository.ErrTaskAccessDenied) {
			return nil, ErrTaskAccessDenied
		}
		return nil, ErrTaskNotFound
	}

	redisConn.InvalidateTasksCache(userID.String())

	task, err := s.repo.GetTaskByID(ctx, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated task: %w", err)
	}

	if req.Status != nil {
		events.PublishTaskStatusChanged(taskID, userID, string(*req.Status), task.Title)
	}

	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error {
	err := s.repo.DeleteTask(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskAccessDenied) {
			return ErrTaskAccessDenied
		}
		return ErrTaskNotFound
	}

	redisConn.InvalidateTasksCache(userID.String())
	events.PublishTaskDeleted(taskID, userID)

	return nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskAccessDenied) {
			return nil, ErrTaskAccessDenied
		}
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *TaskService) GetAllByUser(ctx context.Context, userID uuid.UUID, q pagination.Query) ([]models.Task, int64, error) {
	q = pagination.NewPaginationQuery(q.Page, q.Limit, q.Status)

	allTasks, err := redisConn.GetCachedTasksList(userID.String())
	if err != nil {
		allTasks, err = s.repo.GetAllByUser(ctx, userID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get tasks: %w", err)
		}
		redisConn.CacheTasksList(userID.String(), allTasks)
	}

	if q.Status != "" {
		switch models.TaskStatus(q.Status) {
		case models.TaskStatusTodo, models.TaskStatusInProgress, models.TaskStatusDone:

		default:
			return nil, 0, ErrInvalidStatus
		}
	}

	if q.Status != "" {
		filtered := []models.Task{}
		for _, t := range allTasks {
			if string(t.Status) == q.Status {
				filtered = append(filtered, t)
			}
		}
		allTasks = filtered
	}

	total := int64(len(allTasks))
	start := (q.Page - 1) * q.Limit
	end := start + q.Limit

	if start >= int(total) {
		return []models.Task{}, total, nil
	}
	if end > int(total) {
		end = int(total)
	}

	return allTasks[start:end], total, nil
}

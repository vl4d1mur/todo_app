package repository

import (
	"context"
	"errors"
	"fmt"

	"task_service/internal/db/postgres"
	"task_service/internal/dto"
	"task_service/internal/models"
	"task_service/pkg/log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var _ TaskRepository = (*TaskRepositoryImpl)(nil)

var (
	ErrTaskNotFound     = errors.New("task not found or access denied")
	ErrTaskAccessDenied = errors.New("access to task denied")
)

type TaskRepositoryImpl struct{}

func NewTaskRepository() *TaskRepositoryImpl {
	return &TaskRepositoryImpl{}
}

func (r *TaskRepositoryImpl) CreateTask(ctx context.Context, task *models.Task) error {
	_, err := postgres.DB.Exec(ctx,
		`INSERT INTO tasks (id, user_id, title, description, status, priority, deadline, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		task.ID, task.UserID, task.Title, task.Description, task.Status, task.Priority, task.DeadLine, task.CreatedAt, task.UpdatedAt)
	return err
}

func (r *TaskRepositoryImpl) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]models.Task, error) {
	rows, err := postgres.DB.Query(ctx, `SELECT id, user_id, title, description, status, priority, deadline, created_at, updated_at
		FROM tasks WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	tasks = []models.Task{}
	for rows.Next() {
		var t models.Task
		err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DeadLine, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			log.Logger.Warn().Err(err).Msg("Failed to scan task row")
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepositoryImpl) GetTaskByID(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := postgres.DB.QueryRow(ctx,
		`SELECT id, user_id, title, description, status, priority, deadline, created_at, updated_at FROM tasks WHERE id = $1 AND user_id = $2`,
		taskID, userID).Scan(&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.Status, &task.Priority, &task.DeadLine, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	if task.UserID != userID {
		return nil, ErrTaskAccessDenied
	}

	return &task, nil
}

func (r *TaskRepositoryImpl) UpdateTask(ctx context.Context, taskID, userID uuid.UUID, req dto.UpdateTaskRequest) error {

	var ownerID uuid.UUID
	err := postgres.DB.QueryRow(ctx,
		`SELECT user_id FROM tasks WHERE id = $1`, taskID).Scan(&ownerID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTaskNotFound
		}
		return err
	}

	if ownerID != userID {
		return ErrTaskAccessDenied
	}

	query := "UPDATE tasks SET updated_at = NOW()"
	var args []interface{}
	argCount := 1

	if req.Title != nil {
		query += fmt.Sprintf(", title = $%d", argCount)
		args = append(args, *req.Title)
		argCount++
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", argCount)
		args = append(args, *req.Description)
		argCount++
	}
	if req.Status != nil {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, *req.Status)
		argCount++
	}
	if req.Priority != nil {
		query += fmt.Sprintf(", priority = $%d", argCount)
		args = append(args, *req.Priority)
		argCount++
	}
	if req.Deadline != nil {
		query += fmt.Sprintf(", deadline = $%d", argCount)
		args = append(args, *req.Deadline)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, taskID)

	_, err = postgres.DB.Exec(ctx, query, args...)
	return err
}

func (r *TaskRepositoryImpl) DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error {
	var ownerID uuid.UUID
	err := postgres.DB.QueryRow(ctx, `
	SELECT user_id FROM tasks WHERE id = $1`, taskID).Scan(&ownerID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTaskNotFound
		}
		return err
	}

	if ownerID != userID {
		return ErrTaskAccessDenied
	}

	_, err = postgres.DB.Exec(ctx, "DELETE FROM tasks WHERE id = $1", taskID)

	return err
}


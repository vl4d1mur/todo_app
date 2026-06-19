// task_service/tests/task_service_test/task_service_test.go
package task_service_test

import (
	"context"
	"errors"
	"testing"

	"task_service/internal/dto"
	"task_service/internal/models"
	"task_service/internal/repository"
	"task_service/internal/service"
	"task_service/pkg/pagination"

	"github.com/google/uuid"
)

// ==================== MOCK ====================

type mockTaskRepo struct {
	createErr     error
	getByIDTask   *models.Task
	getByIDErr    error
	updateErr     error
	deleteErr     error
	getAllTasks   []models.Task
	getAllErr     error
}

func (m *mockTaskRepo) CreateTask(ctx context.Context, task *models.Task) error {
	return m.createErr
}
func (m *mockTaskRepo) GetTaskByID(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error) {
	return m.getByIDTask, m.getByIDErr
}
func (m *mockTaskRepo) UpdateTask(ctx context.Context, taskID, userID uuid.UUID, req dto.UpdateTaskRequest) error {
	return m.updateErr
}
func (m *mockTaskRepo) DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error {
	return m.deleteErr
}
func (m *mockTaskRepo) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]models.Task, error) {
	return m.getAllTasks, m.getAllErr
}

// ==================== CREATE ====================

func TestCreateTask_Success(t *testing.T) {
	svc := service.NewTaskService(&mockTaskRepo{})

	userID := uuid.New()
	task, err := svc.CreateTask(context.Background(), userID, dto.CreateTaskRequest{
		Title:       "Test task",
		Description: "Description",
		Priority:    2,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if task.Title != "Test task" {
		t.Errorf("Expected title 'Test task', got %v", task.Title)
	}
	if task.Status != models.TaskStatusTodo {
		t.Errorf("Expected default status 'todo', got %v", task.Status)
	}
	if task.UserID != userID {
		t.Errorf("Expected userID %v, got %v", userID, task.UserID)
	}
}

func TestCreateTask_RepoError(t *testing.T) {
	svc := service.NewTaskService(&mockTaskRepo{createErr: errors.New("db error")})

	_, err := svc.CreateTask(context.Background(), uuid.New(), dto.CreateTaskRequest{
		Title: "Test",
	})

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== UPDATE ====================

func TestUpdateTask_NotFound(t *testing.T) {
	svc := service.NewTaskService(&mockTaskRepo{updateErr: repository.ErrTaskNotFound})

	_, err := svc.UpdateTask(context.Background(), uuid.New(), uuid.New(), dto.UpdateTaskRequest{})

	if !errors.Is(err, service.ErrTaskNotFound) {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}
}

func TestUpdateTask_AccessDenied(t *testing.T) {
	svc := service.NewTaskService(&mockTaskRepo{updateErr: repository.ErrTaskAccessDenied})

	_, err := svc.UpdateTask(context.Background(), uuid.New(), uuid.New(), dto.UpdateTaskRequest{})

	if !errors.Is(err, service.ErrTaskAccessDenied) {
		t.Errorf("Expected ErrTaskAccessDenied, got %v", err)
	}
}

// ==================== DELETE ====================

func TestDeleteTask_Success(t *testing.T) {
	svc := service.NewTaskService(&mockTaskRepo{})

	err := svc.DeleteTask(context.Background(), uuid.New(), uuid.New())

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	svc := service.NewTaskService(&mockTaskRepo{deleteErr: repository.ErrTaskNotFound})

	err := svc.DeleteTask(context.Background(), uuid.New(), uuid.New())

	if !errors.Is(err, service.ErrTaskNotFound) {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}
}

func TestDeleteTask_AccessDenied(t *testing.T) {
	svc := service.NewTaskService(&mockTaskRepo{deleteErr: repository.ErrTaskAccessDenied})

	err := svc.DeleteTask(context.Background(), uuid.New(), uuid.New())

	if !errors.Is(err, service.ErrTaskAccessDenied) {
		t.Errorf("Expected ErrTaskAccessDenied, got %v", err)
	}
}

// ==================== GET BY ID ====================

func TestGetTaskByID_Success(t *testing.T) {
	expected := &models.Task{
		ID:    uuid.New(),
		Title: "Test",
	}
	svc := service.NewTaskService(&mockTaskRepo{getByIDTask: expected})

	task, err := svc.GetTaskByID(context.Background(), expected.ID, uuid.New())

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if task.ID != expected.ID {
		t.Errorf("Expected ID %v, got %v", expected.ID, task.ID)
	}
}

func TestGetTaskByID_NotFound(t *testing.T) {
	svc := service.NewTaskService(&mockTaskRepo{getByIDErr: repository.ErrTaskNotFound})

	_, err := svc.GetTaskByID(context.Background(), uuid.New(), uuid.New())

	if !errors.Is(err, service.ErrTaskNotFound) {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}
}

// ==================== GET ALL — пагинация и фильтр ====================

func TestGetAllByUser_InvalidStatus(t *testing.T) {
	svc := service.NewTaskService(&mockTaskRepo{
		getAllTasks: []models.Task{{ID: uuid.New(), Status: "todo"}},
	})

	_, _, err := svc.GetAllByUser(context.Background(), uuid.New(), pagination.Query{
		Page:   1,
		Limit:  10,
		Status: "wrong_status",
	})

	if !errors.Is(err, service.ErrInvalidStatus) {
		t.Errorf("Expected ErrInvalidStatus, got %v", err)
	}
}

func TestGetAllByUser_FilterByStatus(t *testing.T) {
	tasks := []models.Task{
		{ID: uuid.New(), Status: models.TaskStatusTodo},
		{ID: uuid.New(), Status: models.TaskStatusDone},
		{ID: uuid.New(), Status: models.TaskStatusTodo},
	}
	svc := service.NewTaskService(&mockTaskRepo{getAllTasks: tasks})

	result, total, err := svc.GetAllByUser(context.Background(), uuid.New(), pagination.Query{
		Page:   1,
		Limit:  10,
		Status: "todo",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if total != 2 {
		t.Errorf("Expected 2 tasks with status 'todo', got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 items in result, got %d", len(result))
	}
}

func TestGetAllByUser_Pagination(t *testing.T) {
	tasks := make([]models.Task, 15)
	for i := range tasks {
		tasks[i] = models.Task{ID: uuid.New(), Status: models.TaskStatusTodo}
	}
	svc := service.NewTaskService(&mockTaskRepo{getAllTasks: tasks})

	result, total, err := svc.GetAllByUser(context.Background(), uuid.New(), pagination.Query{
		Page:  2,
		Limit: 5,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(result) != 5 {
		t.Errorf("Expected 5 items on page 2, got %d", len(result))
	}
}
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"task_service/internal/dto"
	"task_service/internal/middleware"
	"task_service/internal/service"
	"task_service/pkg/log"
	"task_service/pkg/pagination"
	"task_service/pkg/validator"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *TaskHandler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid data format")
		return
	}

	if err := validator.Validate(req); err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.service.CreateTask(r.Context(), userID, req)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Task creation error:")
		middleware.RespondWithError(w, http.StatusInternalServerError, "Task creation failed")
		return
	}

	middleware.RespondWithJSON(w, http.StatusCreated, dto.SuccessResponse{
		Message: "Task created successfully",
		Data:    task,
	})
}

func (h *TaskHandler) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid task id")
		return
	}

	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid data format")
		return
	}

	if err := validator.Validate(req); err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.service.UpdateTask(r.Context(), taskID, userID, req)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			middleware.RespondWithError(w, http.StatusNotFound, "Task not found")
		} else if errors.Is(err, service.ErrTaskAccessDenied) {
			middleware.RespondWithError(w, http.StatusForbidden, "Access denied")
		} else {
			log.Logger.Error().Err(err).Msg("Update task error:")
			middleware.RespondWithError(w, http.StatusInternalServerError, "Update task failed")
		}
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "Task updated success",
		Data:    task,
	})
}

func (h *TaskHandler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	if err := h.service.DeleteTask(r.Context(), taskID, userID); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			middleware.RespondWithError(w, http.StatusNotFound, "Task not found")
		} else if errors.Is(err, service.ErrTaskAccessDenied) {
			middleware.RespondWithError(w, http.StatusForbidden, "Access denied")
		} else {
			log.Logger.Error().Err(err).Msg("Delete task error:")
			middleware.RespondWithError(w, http.StatusInternalServerError, "Delete task failed")
		}
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "Task deleted success",
	})
}

func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	status := r.URL.Query().Get("status")
	q := pagination.NewPaginationQuery(page, limit, status)

	tasks, total, err := h.service.GetAllByUser(r.Context(), userID, q)
	if err != nil {
		if errors.Is(err, service.ErrInvalidStatus) {
			middleware.RespondWithError(w, http.StatusBadRequest, "Invalid task status")
		} else {
			log.Logger.Error().Err(err).Msg("Getting tasks error:")
			middleware.RespondWithError(w, http.StatusInternalServerError, "Get tasks failed")
		}

		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, pagination.NewPaginatedResponse(tasks, total, q))
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := h.service.GetTaskByID(r.Context(), taskID, userID)
	if err != nil {
		if errors.Is(err, service.ErrTaskAccessDenied) {
			middleware.RespondWithError(w, http.StatusForbidden, "Access denied")
		} else if errors.Is(err, service.ErrTaskNotFound) {
			middleware.RespondWithError(w, http.StatusNotFound, "Task not found")
		} else {
			log.Logger.Error().Err(err).Msg("delete task error")
			middleware.RespondWithError(w, http.StatusInternalServerError, "Get task by ID error")
		}
		return
	}

	notes, err := h.noteService.GetAllByTask(r.Context(), taskID)
	if err != nil {
		log.Logger.Error().Err(err).Msg("GetNotesByTask error")
		middleware.RespondWithError(w, http.StatusInternalServerError, "Get notes from task failed")
	}

	middleware.RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "Task:",
		Data: dto.TaskWithNotes{
			Task:  *task,
			Notes: notes,
		},
	})
}

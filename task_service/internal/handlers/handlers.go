package handlers

import (
	"task_service/internal/service"
)

type Handler struct {
    Task *TaskHandler
    Note *NoteHandler
}

type TaskHandler struct {
    service service.TaskServiceInterface
    noteService service.NoteServiceInterface
}

type NoteHandler struct {
    service service.NoteServiceInterface
}

func NewHandler(
    taskSvc service.TaskServiceInterface,
    noteSvc service.NoteServiceInterface,
) *Handler {
    return &Handler{
        Task: &TaskHandler{service: taskSvc, noteService: noteSvc},
        Note: &NoteHandler{service: noteSvc},
    }
}
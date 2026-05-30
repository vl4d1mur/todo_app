package handlers

import (
	"todo/internal/service"
)

type Handler struct {
    Auth *AuthHandler
    Task *TaskHandler
    Note *NoteHandler
}

type AuthHandler struct {
    service service.AuthServiceInterface
}

type TaskHandler struct {
    service service.TaskServiceInterface
}

type NoteHandler struct {
    service service.NoteServiceInterface
}

func NewHandler(
    authSvc service.AuthServiceInterface,
    taskSvc service.TaskServiceInterface,
    noteSvc service.NoteServiceInterface,
) *Handler {
    return &Handler{
        Auth: &AuthHandler{service: authSvc},
        Task: &TaskHandler{service: taskSvc},
        Note: &NoteHandler{service: noteSvc},
    }
}
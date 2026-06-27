package routes

import (
	"github.com/gorilla/mux"

	authgrpc "task_service/internal/grpc"
	"task_service/internal/handlers"
	"task_service/internal/health"
	"task_service/internal/middleware"
	"task_service/pkg/metrics"
)

func SetupRoutes(h *handlers.Handler, health *health.Checker, authClient *authgrpc.AuthClient) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.Logger)
	r.Use(metrics.Middleware)

	r.HandleFunc("/healthz", health.Liveness).Methods("GET")
	r.HandleFunc("/readyz", health.Readiness).Methods("GET")

	r.Handle("/metrics", metrics.Handler()).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(authClient))

	api.HandleFunc("/tasks", h.Task.CreateTaskHandler).Methods("POST")
	api.HandleFunc("/tasks", h.Task.GetAllTasks).Methods("GET")
	api.HandleFunc("/tasks/{id}", h.Task.UpdateTaskHandler).Methods("PUT", "PATCH")
	api.HandleFunc("/tasks/{id}", h.Task.DeleteTaskHandler).Methods("DELETE")
	api.HandleFunc("/tasks/{id}", h.Task.GetTaskByID).Methods("GET")

	api.HandleFunc("/tasks/{id}/notes", h.Note.CreateNoteHandler).Methods("POST")
	api.HandleFunc("/tasks/{id}/notes", h.Note.GetNoteHandler).Methods("GET")
	api.HandleFunc("/notes/{noteId}", h.Note.DeleteNoteHandler).Methods("DELETE")
	return r
}

package routes

import (
	"github.com/gorilla/mux"

	authgrpc "task_service/internal/grpc"
	"task_service/internal/handlers"
	"task_service/internal/health"
	"task_service/internal/middleware"
)

func SetupRoutes(h *handlers.Handler, health *health.Checker, authClient *authgrpc.AuthClient) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.Logger)

	r.HandleFunc("/healthz", health.Liveness).Methods("GET")
    r.HandleFunc("/readyz", health.Readiness).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(authClient))

	api.HandleFunc("/tasks", h.Task.CreateTaskHandler).Methods("POST")              // создание задачи
	api.HandleFunc("/tasks", h.Task.GetAllTasks).Methods("GET")                     // получение всех таск пользователя
	api.HandleFunc("/tasks/{id}", h.Task.UpdateTaskHandler).Methods("PUT", "PATCH") // обновление
	api.HandleFunc("/tasks/{id}", h.Task.DeleteTaskHandler).Methods("DELETE")       // удаление
	api.HandleFunc("/tasks/{id}", h.Task.GetTaskByID).Methods("GET")                // получить по айдишнику


	api.HandleFunc("/tasks/{id}/notes", h.Note.CreateNoteHandler).Methods("POST")     // создание заметки
	api.HandleFunc("/tasks/{id}/notes", h.Note.GetNoteHandler).Methods("GET")         // получить по айдишнику
	api.HandleFunc("/notes/{noteId}", h.Note.DeleteNoteHandler).Methods("DELETE") // удалить
	return r
}

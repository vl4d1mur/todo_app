package routes

import (
	"github.com/gorilla/mux"

	"todo/internal/handlers"
	"todo/internal/middleware"
)

func SetupRoutes(h *handlers.Handler) *mux.Router {
	r := mux.NewRouter()

	r.Use(middleware.Logger)

	r.HandleFunc("/api/register", h.Auth.RegisterHandler).Methods("POST")
	r.HandleFunc("/api/login", h.Auth.LoginHandler).Methods("POST")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware)

	api.HandleFunc("/profile", h.Auth.GetProfileHandler).Methods("GET")
	api.HandleFunc("/tasks", h.Task.CreateTaskHandler).Methods("POST")              // создание задачи
	api.HandleFunc("/tasks/{id}", h.Task.UpdateTaskHandler).Methods("PUT", "PATCH") // обновление
	api.HandleFunc("/tasks/{id}", h.Task.DeleteTaskHandler).Methods("DELETE")       // удаление
	api.HandleFunc("/tasks/{id}", h.Task.GetTaskByID).Methods("GET")                // получить по айдишнику
	api.HandleFunc("/tasks", h.Task.GetAllTasks).Methods("GET")                     // получение всех таск пользователя

	api.HandleFunc("/tasks/{id}/notes", h.Note.CreateNoteHandler).Methods("POST")     // создание заметки
	api.HandleFunc("/tasks/{id}/notes", h.Note.GetNoteHandler).Methods("GET")         // получить по айдишнику
	api.HandleFunc("/notes/{noteId}", h.Note.DeleteNoteHandler).Methods("DELETE") // удалить
	return r
}

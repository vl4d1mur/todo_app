package routes

import (
	"github.com/gorilla/mux"

	"auth_service/internal/handlers"
	"auth_service/internal/health"
	"auth_service/internal/middleware"
)

func SetupRoutes(h *handlers.Handler, health *health.Checker) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.Logger)

	r.HandleFunc("/healthz", health.Liveness).Methods("GET")
	r.HandleFunc("/readyz", health.Readiness).Methods("GET")

	r.HandleFunc("/api/register", h.Auth.RegisterHandler).Methods("POST")
	r.HandleFunc("/api/login", h.Auth.LoginHandler).Methods("POST")
	r.HandleFunc("/api/refresh", h.Auth.RefreshHandler).Methods("POST")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware)

	api.HandleFunc("/logout", h.Auth.LogoutHandler).Methods("POST")
	api.HandleFunc("/profile", h.Auth.GetProfileHandler).Methods("GET")
	return r
}

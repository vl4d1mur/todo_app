package routes

import (
	"github.com/gorilla/mux"

	authgrpc "notifier_service/internal/grpc"
	"notifier_service/internal/handlers"
	"notifier_service/internal/health"
	"notifier_service/internal/middleware"
	"notifier_service/pkg/metrics"
)

func SetupRoutes(h *handlers.Handler, health *health.Checker, authClient *authgrpc.AuthClient) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.Logger)
	r.Use(metrics.Middleware)

	r.Handle("/metrics", metrics.Handler()).Methods("GET")

	r.HandleFunc("/healthz", health.Liveness).Methods("GET")
	r.HandleFunc("/readyz", health.Readiness).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(authClient))

	api.HandleFunc("/notifications", h.Notifier.GetNotificationsHandler).Methods("GET")
	return r
}

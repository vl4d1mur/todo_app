package handlers

import (
	"notifier_service/internal/service"
)

type Handler struct {
	Notifier *NotificationHandler
}

type NotificationHandler struct {
	service service.NotifierServiceInterface
}

func NewHandler(
	notifSvc service.NotifierServiceInterface,
) *Handler {
	return &Handler{
		Notifier: &NotificationHandler{service: notifSvc},
	}
}

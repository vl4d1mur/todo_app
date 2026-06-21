package handlers

import (
	"auth_service/internal/service"
)

type Handler struct {
	Auth     *AuthHandler
	Telegram *TelegramHandler
}

type TelegramHandler struct {
	service service.TelegramServiceInterface
}

type AuthHandler struct {
	service service.AuthServiceInterface
}

func NewHandler(authSvc service.AuthServiceInterface, telegramSvc service.TelegramServiceInterface) *Handler {
	return &Handler{
		Auth:     &AuthHandler{service: authSvc},
		Telegram: &TelegramHandler{service: telegramSvc},
	}
}

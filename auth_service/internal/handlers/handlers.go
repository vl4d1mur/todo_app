package handlers

import (
	"auth_service/internal/service"
)

type Handler struct {
    Auth *AuthHandler
}

type AuthHandler struct {
    service service.AuthServiceInterface
}

func NewHandler(authSvc service.AuthServiceInterface,) *Handler {
    return &Handler{
        Auth: &AuthHandler{service: authSvc},
    }
}
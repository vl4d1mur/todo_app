package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"auth_service/internal/dto"
	"auth_service/internal/middleware"
	"auth_service/internal/service"
	"auth_service/pkg/log"
)

func (h *AuthHandler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	user, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			middleware.RespondWithError(w, http.StatusNotFound, "User not found")
		} else {
			log.Logger.Error().Err(err).Msg("Get profile error")
			middleware.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "User profile",
		Data:    user,
	})
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Bad format")
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlrdeadyExists) {
			middleware.RespondWithError(w, http.StatusConflict, "User with this email already exists")
		} else {
			log.Logger.Error().Err(err).Msg("Register error")
			middleware.RespondWithError(w, http.StatusInternalServerError, "Registartion failed")
		}
		return
	}

	middleware.RespondWithJSON(w, http.StatusCreated, dto.SuccessResponse{
		Message: "User registration succesful",
		Data:    user,
	})
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid data format")
		return
	}

	tokens, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			middleware.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		} else {
			log.Logger.Error().Err(err).Msg("Login error")
			middleware.RespondWithJSON(w, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid data format")
		return
	}

	if req.RefreshToken == "" {
		middleware.RespondWithError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	tokens, err := h.service.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSession) {
			middleware.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		} else {
			log.Logger.Error().Err(err).Msg("Refresh error")
			middleware.RespondWithError(w, http.StatusInternalServerError, "Refresh failed")
		}
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Logger.Error().Err(err).Msg("Logout decode error")
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid data format")
		return
	}

	if req.RefreshToken == "" {
		middleware.RespondWithError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		if errors.Is(err, service.ErrInvalidSession) {
			middleware.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		} else {
			log.Logger.Error().Err(err).Msg("Logout error")
			middleware.RespondWithError(w, http.StatusInternalServerError, "Logout failed")
		}
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "Logged out successfully",
	})
}

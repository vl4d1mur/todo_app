package handlers

import (
	"net/http"

	"auth_service/internal/dto"
	"auth_service/internal/middleware"
	"auth_service/pkg/log"
)

func (h *TelegramHandler) GenerateTelegramCodeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	code, err := h.service.GenerateTelegramCode(r.Context(), userID)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to generate Telegram code")
		middleware.RespondWithError(w, http.StatusInternalServerError, "Failed to generate code")
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "Send this code to bot via /start <code>. Expires in 10 minutes.",
		Data: map[string]string{
			"code": code,
		},
	})
}

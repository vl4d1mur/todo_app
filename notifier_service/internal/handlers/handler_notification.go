package handlers

import (
	//"encoding/json"
	//"errors"
	"net/http"
	"strconv"

	"notifier_service/internal/middleware"
	"notifier_service/internal/service"
	"notifier_service/pkg/log"
	"notifier_service/pkg/pagination"
)

func NewNotificationHandler(svc service.NotifierServiceInterface) *NotificationHandler {
	return &NotificationHandler{service: svc}
}

func (h *NotificationHandler) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	q := pagination.NewPaginationQuery(page, limit, "")

	notifications, total, err := h.service.GetByUserID(r.Context(), userID, q)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get notifications")
		middleware.RespondWithError(w, http.StatusInternalServerError, "Failed to get notifications")
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, pagination.NewPaginatedResponse(notifications, total, q))
}

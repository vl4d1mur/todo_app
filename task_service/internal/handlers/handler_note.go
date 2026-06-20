package handlers

import (
	"encoding/json"
	"net/http"

	"task_service/internal/dto"
	"task_service/internal/middleware"
	"task_service/pkg/log"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (h *NoteHandler) CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req dto.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid data format")
		return
	}

	note, err := h.service.CreateNote(r.Context(), taskID, userID, req)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Note creating error:")
		middleware.RespondWithError(w, http.StatusInternalServerError, "Creating note error:")
		return
	}

	middleware.RespondWithJSON(w, http.StatusCreated, dto.SuccessResponse{
		Message: "Note created",
		Data:    note,
	})

}

func (h *NoteHandler) GetNoteHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	notes, err := h.service.GetAllByTask(r.Context(), taskID)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Notes receiving error:")
		middleware.RespondWithError(w, http.StatusInternalServerError, "Notes receiving error")
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "Notes list",
		Data:    notes,
	})
}

func (h *NoteHandler) DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		middleware.RespondWithError(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	vars := mux.Vars(r)
	noteID, err := bson.ObjectIDFromHex(vars["noteId"])
	if err != nil {
		middleware.RespondWithError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	if err := h.service.DeleteNote(r.Context(), noteID, userID); err != nil {
		middleware.RespondWithError(w, http.StatusNotFound, "Note not found or access denied")
		return
	}

	middleware.RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "Note deleted",
	})
}

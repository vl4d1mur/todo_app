package service

import (
	"context"
	"fmt"
	"time"

	"task_service/internal/dto"
	"task_service/internal/events"
	"task_service/internal/models"
	"task_service/internal/repository"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrNoteNotFound = fmt.Errorf("note not found or access denied")

type NoteService struct {
    repo repository.NoteRepository
}

func NewNoteService(repo repository.NoteRepository) *NoteService {
    return &NoteService{repo: repo}
}

func (s *NoteService) CreateNote(ctx context.Context, taskID, userID uuid.UUID, req dto.CreateNoteRequest) (*models.Note, error) {

	note := models.Note{
		ID:        bson.NewObjectID(),
		TaskID:    taskID,
		AuthorID:  userID,
		Text:      req.Text,
		Meta:      req.Meta,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateNote(ctx, &note); err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	events.PublishNoteAdded(taskID, userID, note.ID.Hex(), note.Text)
	
	return &note, nil
}

func (s *NoteService) DeleteNote(ctx context.Context, noteID bson.ObjectID, userID uuid.UUID) error {
	if err := s.repo.DeleteNote(ctx, noteID, userID); err != nil {
		return err
	}
	
	return nil
}

func (s *NoteService) GetAllByTask(ctx context.Context, taskID uuid.UUID) ([]models.Note, error) {
	return s.repo.GetAllByTask(ctx, taskID)
}

package dto

type CreateNoteRequest struct {
	Text string         `json:"text" validate:"required,max=5000"`
	Meta map[string]any `json:"meta"`
}

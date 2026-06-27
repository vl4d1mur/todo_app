package service

import (
	"context"
	"crypto/rand"
	"fmt"

	"auth_service/internal/db/redisConn"
	"auth_service/pkg/metrics"

	"github.com/google/uuid"
)

type TelegramService struct{}

func NewTelegramService() *TelegramService {
	return &TelegramService{}
}

func (s *TelegramService) GenerateTelegramCode(ctx context.Context, userID uuid.UUID) (string, error) {
	code, err := generateCode(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}

	if err := redisConn.SaveTelegramCode(code, userID.String()); err != nil {
		return "", fmt.Errorf("failed to save code: %w", err)
	}

	metrics.TelegramCodesGenerated.Inc()
	return code, err
}

func generateCode(length int) (string, error) {
	const charset = "0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = charset[int(b)%len(charset)]
	}
	return string(bytes), nil
}

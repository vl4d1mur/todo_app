package jwt_test

import (
	"testing"

	"auth_service/internal/config"
	"auth_service/internal/models"
	"auth_service/pkg/jwt"

	"github.com/google/uuid"
)

func init() {
	config.JwtSecret = []byte("test-secret-key")
	config.JwtRefreshSecret = []byte("test-refresh-secret-key")
}

func TestGenerateAndParseAccess_Success(t *testing.T) {
	user := models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	token, err := jwt.GenerateAccess(user)
	if err != nil {
		t.Fatalf("GenerateAccess error: %v", err)
	}

	claims, err := jwt.ParseAccess(token)
	if err != nil {
		t.Fatalf("ParseAccess error: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected userID %v, got %v", user.ID, claims.UserID)
	}
}

func TestGenerateAndParseRefresh_Success(t *testing.T) {
	user := models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	token, err := jwt.GenerateRefresh(user)
	if err != nil {
		t.Fatalf("GenerateRefresh error: %v", err)
	}

	claims, err := jwt.ParseRefresh(token)
	if err != nil {
		t.Fatalf("ParseRefresh error: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected userID %v, got %v", user.ID, claims.UserID)
	}
}

func TestParseAccess_InvalidToken(t *testing.T) {
	_, err := jwt.ParseAccess("invalid.token.string")
	if err != jwt.ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestParseAccess_WrongSecret(t *testing.T) {
	user := models.User{ID: uuid.New()}

	// Генерируем refresh токен и пробуем распарсить как access
	token, _ := jwt.GenerateRefresh(user)
	_, err := jwt.ParseAccess(token)
	if err != jwt.ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken when parsing with wrong secret, got %v", err)
	}
}

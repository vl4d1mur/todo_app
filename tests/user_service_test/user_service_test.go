package service_test

import (
	"context"
	"errors"
	"testing"

	"todo/internal/config"
	"todo/internal/dto"
	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/service"
	"todo/pkg/hash"
	"todo/pkg/jwt"

	"github.com/google/uuid"
)

func init() {
	config.JwtSecret = []byte("test-secret-key")
	config.JwtRefreshSecret = []byte("test-refresh-secret-key")
}

type mockUserRepo struct {
	existsByEmail  bool
	existsErr      error
	createErr      error
	userByEmail    *models.User
	userByEmailErr error
	userByID       *models.User
	userByIDErr    error
}

func (m *mockUserRepo) ExistByEmail(ctx context.Context, email string) (bool, error) {
	return m.existsByEmail, m.existsErr
}
func (m *mockUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	return m.createErr
}
func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.userByEmail, m.userByEmailErr
}
func (m *mockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.userByID, m.userByIDErr
}

type mockSessionRepo struct {
	createErr error
	session   *models.Session
	getErr    error
	deleteErr error
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, s *models.Session) error {
	return m.createErr
}
func (m *mockSessionRepo) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	return m.session, m.getErr
}
func (m *mockSessionRepo) DeleteSessionByToken(ctx context.Context, token string) error {
	return m.deleteErr
}
func (m *mockSessionRepo) DeleteAllSessionsByUser(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestRegister_Success(t *testing.T) {
	svc := service.NewAuthService(
		&mockUserRepo{existsByEmail: false},
		&mockSessionRepo{},
	)

	user, err := svc.Register(context.Background(), dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %v", user.Email)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := service.NewAuthService(
		&mockUserRepo{existsByEmail: true},
		&mockSessionRepo{},
	)

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	if !errors.Is(err, service.ErrUserAlrdeadyExists) {
		t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hashed, _ := hash.HashPassword("password123")

	svc := service.NewAuthService(
		&mockUserRepo{
			userByEmail: &models.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Password: hashed,
			},
		},
		&mockSessionRepo{},
	)

	tokens, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if tokens.AccessToken == "" {
		t.Error("AccessToken must not be empty")
	}
	if tokens.RefreshToken == "" {
		t.Error("RefreshToken must not be empty")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hashed, _ := hash.HashPassword("correctpassword")

	svc := service.NewAuthService(
		&mockUserRepo{
			userByEmail: &models.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Password: hashed,
			},
		},
		&mockSessionRepo{},
	)

	_, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	if !errors.Is(err, service.ErrInvalidPassword) {
		t.Errorf("Expected ErrInvalidPassword, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := service.NewAuthService(
		&mockUserRepo{userByEmailErr: repository.ErrUserNotFound},
		&mockSessionRepo{},
	)

	_, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "notexist@example.com",
		Password: "password123",
	})

	if !errors.Is(err, service.ErrInvalidPassword) {
		t.Errorf("Expected ErrInvalidPassword, got %v", err)
	}
}

func TestLogout_Success(t *testing.T) {
	svc := service.NewAuthService(
		&mockUserRepo{},
		&mockSessionRepo{},
	)

	err := svc.Logout(context.Background(), "some-refresh-token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestLogout_InvalidSession(t *testing.T) {
	svc := service.NewAuthService(
		&mockUserRepo{},
		&mockSessionRepo{deleteErr: repository.ErrSessionNotFound},
	)

	err := svc.Logout(context.Background(), "invalid-token")
	if !errors.Is(err, service.ErrInvalidSession) {
		t.Errorf("Expected ErrInvalidSession, got %v", err)
	}
}

func TestRefreshToken_InvalidJWT(t *testing.T) {
	svc := service.NewAuthService(
		&mockUserRepo{},
		&mockSessionRepo{},
	)

	_, err := svc.RefreshToken(context.Background(), "invalid.token")
	if !errors.Is(err, service.ErrInvalidSession) {
		t.Errorf("Expected ErrInvalidSession, got %v", err)
	}
}

func TestRefreshToken_SessionNotFound(t *testing.T) {
	user := models.User{ID: uuid.New()}
	token, _ := jwt.GenerateRefresh(user)

	svc := service.NewAuthService(
		&mockUserRepo{},
		&mockSessionRepo{getErr: repository.ErrSessionNotFound},
	)

	_, err := svc.RefreshToken(context.Background(), token)
	if !errors.Is(err, service.ErrInvalidSession) {
		t.Errorf("Expected ErrInvalidSession, got %v", err)
	}
}
//go:build integration

package integration

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"auth_service/internal/config"
	"auth_service/internal/db/postgres"
	"auth_service/internal/db/redisConn"
	"auth_service/internal/dto"
	"auth_service/internal/repository"
	"auth_service/internal/service"
	"auth_service/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tcPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAuthService_FullFlow(t *testing.T) {
	log.InitLogger()
	ctx := context.Background()

	// ==================== POSTGRES ====================

	_, currentFile, _, _ := runtime.Caller(0)
	schemaPath := filepath.Join(filepath.Dir(currentFile), "testdata", "schema.sql")

	pgContainer, err := tcPostgres.Run(ctx,
		"postgres:16-alpine",
		tcPostgres.WithDatabase("auth_test"),
		tcPostgres.WithUsername("postgres"),
		tcPostgres.WithPassword("root"),
		tcPostgres.WithInitScripts(schemaPath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	assert.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	pgDSN, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err)

	// ==================== REDIS ====================

	redisContainer, err := tcRedis.Run(ctx, "redis:7-alpine")
	assert.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	redisAddr, err := redisContainer.Endpoint(ctx, "")
	assert.NoError(t, err)

	// ==================== CONFIG ====================

	config.PostgresDSN = pgDSN
	config.RedisAddr = redisAddr
	config.RedisPassword = ""
	config.JwtSecret = []byte("test-jwt-secret")
	config.JwtRefreshSecret = []byte("test-refresh-secret")

	// ==================== CONNECT ====================

	postgres.ConnectPostgres()
	defer postgres.ClosePostgres()

	redisConn.ConnectRedis()
	defer redisConn.CloseRedis()

	// ==================== SERVICES ====================

	userRepo := repository.NewUserRepository()
	sessionRepo := repository.NewSessionRepository()
	authSvc := service.NewAuthService(userRepo, sessionRepo)

	// ==================== 1. REGISTER ====================

	user, err := authSvc.Register(ctx, dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)

	// ==================== 2. DUPLICATE REGISTER ====================

	_, err = authSvc.Register(ctx, dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "another",
		Name:     "Another User",
	})
	assert.Error(t, err, "Should fail on duplicate email")

	// ==================== 3. LOGIN ====================

	tokens, err := authSvc.Login(ctx, dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	// ==================== 4. LOGIN WITH WRONG PASSWORD ====================

	_, err = authSvc.Login(ctx, dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})
	assert.Error(t, err, "Should fail with wrong password")

	// ==================== 5. CHECK SESSION IN DB ====================

	session, err := sessionRepo.GetSessionByToken(ctx, tokens.RefreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, user.ID, session.UserID)

	// ==================== 6. REFRESH TOKEN ====================

	newTokens, err := authSvc.RefreshToken(ctx, tokens.RefreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)
	assert.NotEqual(t, tokens.RefreshToken, newTokens.RefreshToken, "New refresh token should be different")

	// ==================== 7. OLD REFRESH TOKEN SHOULD BE DELETED ====================

	_, err = sessionRepo.GetSessionByToken(ctx, tokens.RefreshToken)
	assert.Error(t, err, "Old refresh token should not exist after rotation")

	// ==================== 8. LOGOUT ====================

	err = authSvc.Logout(ctx, newTokens.RefreshToken)
	assert.NoError(t, err)

	// ==================== 9. SESSION SHOULD BE DELETED ====================

	_, err = sessionRepo.GetSessionByToken(ctx, newTokens.RefreshToken)
	assert.Error(t, err, "Session should not exist after logout")

	// ==================== 10. GET PROFILE ====================

	profile, err := authSvc.GetProfile(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, profile.Email)
}

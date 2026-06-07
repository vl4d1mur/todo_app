package config

import (
	"bytes"
	//"log"
	"os"

	"auth_service/pkg/log"

	"github.com/joho/godotenv"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	ServerPort       string
	PostgresDSN      string
	AppEnv           string
	JwtSecret        []byte
	JwtRefreshSecret []byte
	RedisAddr        string
	RedisPassword    string
	GrpcPort         string
)

func LoadConfig() {
	if err := godotenv.Load("config/.env.example"); err != nil {
		log.Logger.Error().Err(err).Msg(".env file not found")
	}

	ServerPort = getEnv("SERVER_PORT", ":8090")
	GrpcPort = getEnv("GRPC_PORT", ":50051")
	PostgresDSN = getEnv("POSTGRES_DSN", "postgres://postgres:root@localhost:5432/users")
	JwtSecret = []byte(getEnv("JWT_SECRET", "abeba229"))
	JwtRefreshSecret = []byte(getEnv("JWT_REFRESH_SECRET", "eriolergjiergjilohuio2347890"))
	AppEnv = getEnv("APP_ENV", "development")
	RedisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	RedisPassword = getEnv("REDIS_PASSWORD", "")

	if bytes.Equal(JwtSecret, []byte("abeba229")) && AppEnv == "production" {
		log.Logger.Fatal().Msg("Change JWT secret for prod")
	}
	log.Logger.Info().Msg("Config loaded")
}

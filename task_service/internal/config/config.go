package config

import (
	"bytes"
	//"log"
	"os"

	"task_service/pkg/log"

	"github.com/joho/godotenv"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	ServerPort        string
	PostgresDSN       string
	AppEnv            string
	JwtSecret         []byte
	JwtRefreshSecret  []byte
	MongoUri          string
	MongoDbName       string
	MongoDbCollection string
	RedisAddr         string
	RedisPassword     string
	NatsURL           string
	AuthGrpcAddr      string
)

func LoadConfig() {
	if err := godotenv.Load("config/.env.example"); err != nil {
		log.Logger.Error().Err(err).Msg(".env file not found")
	}

	ServerPort = getEnv("SERVER_PORT", ":8091")
	AuthGrpcAddr = getEnv("AUTH_GRPC_ADDR", "localhost:50051")
	PostgresDSN = getEnv("POSTGRES_DSN", "postgres://postgres:root@localhost:5432/tasks")
	JwtSecret = []byte(getEnv("JWT_SECRET", "abeba229"))
	JwtRefreshSecret = []byte(getEnv("JWT_REFRESH_SECRET", "eriolergjiergjilohuio2347890"))
	AppEnv = getEnv("APP_ENV", "development")
	MongoUri = getEnv("MONGO_URI", "mongodb://127.0.0.1:27017")
	MongoDbName = getEnv("MONGO_DB_NAME", "todo_notes")
	MongoDbCollection = getEnv("MONGO_DB_COLLECTION", "notes")
	RedisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	RedisPassword = getEnv("REDIS_PASSWORD", "")
	NatsURL = getEnv("NATS_URL", "nats://localhost:4222")

	if bytes.Equal(JwtSecret, []byte("abeba229")) && AppEnv == "production" {
		log.Logger.Fatal().Msg("Change JWT secret for prod")
	}
	log.Logger.Info().Msg("Config loaded")
}

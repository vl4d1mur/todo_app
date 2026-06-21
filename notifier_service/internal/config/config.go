package config

import (
	"os"

	"notifier_service/pkg/log"

	"github.com/joho/godotenv"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	ServerPort              string
	MongoUri                string
	MongoDbName             string
	MongoDbCollection       string
	MongoDeadlineCollection string
	NatsURL                 string
	AuthGrpcAddr            string
	SmtpHost                string
	SmtpPort                string
	SmtpFrom                string
	TelegramBotToken        string
	AppEnv                  string
)

func LoadConfig() {
	if err := godotenv.Load("config/.env.example"); err != nil {
		log.Logger.Error().Err(err).Msg(".env file not found")
	}

	ServerPort = getEnv("SERVER_PORT", ":8092")
	MongoUri = getEnv("MONGO_URI", "mongodb://localhost:27017")
	MongoDbName = getEnv("MONGO_DB_NAME", "notifier_db")
	MongoDbCollection = getEnv("MONGO_COLLECTION", "notifications")
	MongoDeadlineCollection = getEnv("MONGO_DEADLINE_COLLECTION", "task_deadlines")
	NatsURL = getEnv("NATS_URL", "nats://localhost:4222")
	AuthGrpcAddr = getEnv("AUTH_GRPC_ADDR", "localhost:50051")
	SmtpHost = getEnv("SMTP_HOST", "0.0.0.0")
	SmtpPort = getEnv("SMTP_PORT", "1025")
	SmtpFrom = getEnv("SMTP_FROM", "noreply@todo.local")
	TelegramBotToken = getEnv("TELEGRAM_BOT_TOKEN", "")
	AppEnv = getEnv("APP_ENV", "development")

	log.Logger.Info().Msg("Config loaded")
}

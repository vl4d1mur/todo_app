package config

import (
	"bytes"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	ServerPort    string
	PostgresDSN   string
	AppEnv        string
	JwtSecret     []byte
	MongoUri      string
	MongoDbName   string
	MongoDbCollection string
	RedisAddr     string
	RedisPassword string
	NatsURL       string
)

func LoadConfig() {
	if err := godotenv.Load("config/.env.example"); err != nil {
		log.Println(".env file not found")
	}

	ServerPort = getEnv("SERVER_PORT", ":8090")
	PostgresDSN = getEnv("POSTGRES_DSN", "postgres://postgres:root@localhost:5432/todo_app")
	JwtSecret = []byte(getEnv("JWT_SECRET", "abeba229"))
	AppEnv = getEnv("APP_ENV", "development")
	MongoUri = getEnv("MONGO_URI", "mongodb://127.0.0.1:27017")
	MongoDbName = getEnv("MONGO_DB_NAME", "todo_notes")
	MongoDbCollection = getEnv("MONGO_DB_COLLECTION", "notes")
	RedisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	RedisPassword = getEnv("REDIS_PASSWORD", "")
	NatsURL = getEnv("NATS_URL", "nats://localhost:4222")

	if bytes.Equal(JwtSecret, []byte("abeba229")) && AppEnv == "production" {
		log.Fatal("Change JWT secret for prod")
	}
	log.Println("Config loaded")
}
